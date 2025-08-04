// NordVPN daemon.
package main

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"golang.org/x/net/netutil"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/clientid"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/allowlist"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/forwarder"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/iptables"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/notables"
	"github.com/NordSecurity/nordvpn-linux/daemon/netstate"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	telemetrypb "github.com/NordSecurity/nordvpn-linux/daemon/pb/telemetry/v1"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes/ifgroup"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes/iprule"
	netlinkrouter "github.com/NordSecurity/nordvpn-linux/daemon/routes/netlink"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes/norouter"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes/norule"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/daemon/telemetry"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/firstopen"
	"github.com/NordSecurity/nordvpn-linux/events/logger"
	"github.com/NordSecurity/nordvpn-linux/events/meshunsetter"
	"github.com/NordSecurity/nordvpn-linux/events/refresher"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	grpcmiddleware "github.com/NordSecurity/nordvpn-linux/grpc_middleware"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/ipv6"
	"github.com/NordSecurity/nordvpn-linux/kernel"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/meshnet/inviter"
	"github.com/NordSecurity/nordvpn-linux/meshnet/mapper"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/meshnet/registry"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	norduserservice "github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/sharedctx"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
	"github.com/google/uuid"

	"google.golang.org/grpc"
)

// Values set when building the application
var (
	Salt        = ""
	Version     = "0.0.0"
	Environment = ""
	PackageType = ""
	Arch        = ""
	Port        = 6960
	ConnType    = "unix"
	ConnURL     = internal.DaemonSocket
	RemotePath  = ""
)

// Environment constants
const (
	EnvKeyPath = "PATH"
	EnvValPath = ":/bin:/sbin:/usr/bin:/usr/sbin"
	// EnvIgnoreHeaderValidation can only be used in `dev` builds. Setting this to `1` makes
	// API client to ignore X-headers. This makes setting up MITM proxies up possible. This
	// should not be used for regular usage.
	EnvIgnoreHeaderValidation = "IGNORE_HEADER_VALIDATION"
	EnvNordCdnUrl             = "NORD_CDN_URL"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// set PATH env for cli
	currentHost := os.Getenv(EnvKeyPath)
	_ = os.Setenv(EnvKeyPath, currentHost+EnvValPath)
}

// socketType used by gRPC listener
type socketType string

const (
	// sockUnix defines that gRPC server is listening to UNIX socket
	sockUnix socketType = "unix"
	// sockTCP defines that gRPC server is listening to TCP socket
	sockTCP socketType = "tcp"
)

func initializeStaticConfig(machineID uuid.UUID) config.StaticConfigManager {
	staticCfgManager := config.NewFilesystemStaticConfigManager()
	if err := staticCfgManager.SetRolloutGroup(remote.GenerateRolloutGroup(machineID)); err != nil {
		if !errors.Is(err, config.ErrStaticValueAlreadySet) {
			log.Println(internal.ErrorPrefix, "failed to configure rollout group:", err)
		}
	}
	return staticCfgManager
}

func main() {
	// pprof
	if internal.IsDevEnv(Environment) {
		go func() {
			// #nosec G114 -- not used in production
			if err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil); err != nil {
				log.Println(internal.ErrorPrefix, err)
			}
		}()
	}

	// Logging

	log.SetOutput(os.Stdout)
	log.Println(internal.InfoPrefix, "Daemon has started")

	machineIdGenerator := config.NewMachineID(os.ReadFile, os.Hostname)

	// Config
	configEvents := daemonevents.NewConfigEvents()
	fsystem := config.NewFilesystemConfigManager(
		config.SettingsDataFilePath,
		config.InstallFilePath,
		Salt,
		machineIdGenerator,
		config.StdFilesystemHandle{},
		configEvents.Config,
	)

	// Remove any remains of IPv6 settings
	if err := fsystem.SaveWith(removeIPv6Remains); err != nil {
		log.Println(internal.ErrorPrefix, "failed to remove IPv6 entries from settings ", err)
	}

	var cfg config.Config
	if err := fsystem.Load(&cfg); err != nil {
		log.Println(err)
		if err := fsystem.Reset(false, false); err != nil {
			log.Fatalln(err)
		}
	}

	// Events

	daemonEvents := daemonevents.NewEventsEmpty()
	meshnetEvents := meshnet.NewEventsEmpty()
	debugSubject := &subs.Subject[string]{}
	infoSubject := &subs.Subject[string]{}
	errSubject := &subs.Subject[error]{}
	heartBeatSubject := &subs.Subject[time.Duration]{}
	httpCallsSubject := &subs.Subject[events.DataRequestAPI]{}

	loggerSubscriber := logger.Subscriber{}
	if internal.Environment(Environment) == internal.Development {
		debugSubject.Subscribe(loggerSubscriber.NotifyMessage)
	}
	apiLogFn := loggerSubscriber.NotifyRequestAPI
	if internal.IsDevEnv(Environment) {
		apiLogFn = loggerSubscriber.NotifyRequestAPIVerbose
	}

	httpCallsSubject.Subscribe(apiLogFn)
	infoSubject.Subscribe(loggerSubscriber.NotifyInfo)
	errSubject.Subscribe(loggerSubscriber.NotifyError)

	daemonEvents.Settings.Subscribe(logger.NewSubscriber())

	// try to restore resolv.conf if target file contains Nordvpn changes
	dns.RestoreResolvConfFile()

	// Firewall
	stateModule := "conntrack"
	stateFlag := "--ctstate"
	chainPrefix := ""
	iptablesAgent := iptables.New(
		stateModule,
		stateFlag,
		chainPrefix,
		iptables.FilterSupportedIPTables([]string{"iptables", "ip6tables"}),
	)
	fw := firewall.NewFirewall(
		&notables.Facade{},
		iptablesAgent,
		debugSubject,
		cfg.Firewall,
	)

	// API
	var err error
	var validator response.Validator
	if !internal.IsProdEnv(Environment) && os.Getenv(EnvIgnoreHeaderValidation) == "1" {
		validator = response.NoopValidator{}
	} else {
		validator, err = response.NewNordValidator()
		if err != nil {
			log.Fatalln("Error on creating validator:", err)
		}
	}

	userAgent, err := request.GetUserAgentValue(Version, sysinfo.GetHostOSPrettyName)
	if err != nil {
		userAgent = fmt.Sprintf("%s/%s (unknown)", request.AppName, Version)
		log.Printf("Error while constructing UA value: %s. Falls back to default: %s\n", err, userAgent)
	}

	httpGlobalCtx, httpCancel := context.WithCancel(context.Background())

	// simple standard http client with dialer wrapped inside
	httpClientSimple := request.NewStdHTTP()
	httpClientSimple.Transport = request.NewHTTPReTransport(
		1, 1, "HTTP/1.1", func() http.RoundTripper {
			return request.NewPublishingRoundTripper(
				request.NewContextRoundTripper(request.NewStdTransport(), httpGlobalCtx),
				httpCallsSubject,
			)
		}, nil)

	cdnUrl := core.CDNURL
	if !internal.IsProdEnv(Environment) && os.Getenv(EnvNordCdnUrl) != "" {
		cdnUrl = os.Getenv(EnvNordCdnUrl)
	}
	log.Println(internal.InfoPrefix, "CDN URL:", cdnUrl)

	threatProtectionLiteServers := func() *dns.NameServers {
		cdn := core.NewCDNAPI(
			userAgent,
			cdnUrl,
			httpClientSimple,
			validator,
		)
		nameservers, err := cdn.ThreatProtectionLite()
		if err != nil {
			log.Println(internal.ErrorPrefix, "error retrieving nameservers:", err)
			return dns.NewNameServers(nil)
		}
		return dns.NewNameServers(nameservers.Servers)
	}()

	resolver := network.NewResolver(fw, threatProtectionLiteServers)

	if err := SetBufferSizeForHTTP3(); err != nil {
		log.Println(internal.WarningPrefix, "failed to set buffer size for HTTP/3:", err)
	}

	httpClientWithRotator := request.NewStdHTTP()
	httpClientWithRotator.Transport = createTimedOutTransport(
		resolver,
		cfg.FirewallMark,
		httpCallsSubject,
		daemonEvents.Service.Connect,
		httpGlobalCtx,
	)

	cdnAPI := core.NewCDNAPI(
		userAgent,
		cdnUrl,
		httpClientWithRotator,
		validator,
	)

	repoAPI := daemon.NewRepoAPI(
		userAgent,
		daemon.RepoURL,
		Version,
		internal.Environment(Environment),
		PackageType,
		Arch,
		httpClientSimple,
	)
	gwret := netlinkrouter.Retriever{}
	dnsSetter := dns.NewSetter(infoSubject)
	dnsHostSetter := dns.NewHostsFileSetter(dns.HostsFilePath)

	eventsDbPath := filepath.Join(internal.DatFilesPathCommon, "moose.db")
	if err := assignMooseDBPermissions(eventsDbPath); err != nil {
		log.Fatalln(err)
	}

	machineID := machineIdGenerator.GetMachineID()
	staticCfg := initializeStaticConfig(machineID)

	// obfuscated machineID and add the mask to identify how the ID was generated
	deviceID := fmt.Sprintf(
		"%x_%d",
		sha256.Sum256([]byte(machineID.String()+Salt)),
		machineIdGenerator.GetUsedInformationMask(),
	)

	invalidSessionErrHandlingReg := internal.NewErrorHandlingRegistry[error]()

	// encapsulating initialization logic
	clientAPI, accessTokenSessionStore := func() (core.ClientAPI, session.SessionStore) {
		api := core.NewSimpleAPI(
			userAgent,
			daemon.BaseURL,
			httpClientWithRotator,
			validator,
		)

		sessionStore := buildAccessTokenSessionStore(fsystem, invalidSessionErrHandlingReg, api)
		return core.NewSmartClientAPI(api, sessionStore), sessionStore
	}()

	// populate build target configuration
	buildTarget := config.BuildTarget{
		Version:      Version,
		Environment:  Environment,
		Architecture: Arch}
	if archVariant, err := machineIdGenerator.GetArchitectureVariantName(sysinfo.GetHostArchitecture()); err == nil {
		buildTarget.Architecture = archVariant
	}

	analytics := newAnalytics(
		eventsDbPath,
		fsystem,
		clientAPI,
		*httpClientSimple,
		buildTarget,
		deviceID)

	heartBeatSubject.Subscribe(analytics.NotifyHeartBeat)
	httpCallsSubject.Subscribe(analytics.NotifyRequestAPI)
	daemonEvents.Subscribe(analytics)

	firstopen.RegisterNotifier(
		fsystem,
		daemonEvents.Service.DeviceLocation,
		daemonEvents.Service.UiItemsClick,
	)

	daemonEvents.Service.Connect.Subscribe(loggerSubscriber.NotifyConnect)
	daemonEvents.Settings.Publish(cfg)

	rolloutGroup, err := staticCfg.GetRolloutGroup()
	if err != nil {
		log.Println(internal.ErrorPrefix, "getting rollout group:", err)
		// in case of error, rollout group is `0`
	}
	rcConfig := getRemoteConfigGetter(
		buildTarget,
		RemotePath,
		cdnAPI,
		remote.NewRemoteConfigAnalytics(
			daemonEvents.Debugger.DebuggerEvents,
			rolloutGroup,
		),
		rolloutGroup,
	)

	// try to load config from disk if it was previously downloaded
	rcConfig.TryPreload()

	vpnLibConfigGetter := vpnLibConfigGetterImplementation(fsystem, rcConfig)

	internalVpnEvents := vpn.NewInternalVPNEvents()

	// Networker
	vpnFactory := getVpnFactory(eventsDbPath, cfg.FirewallMark,
		internal.IsDevEnv(Environment), vpnLibConfigGetter, Version, internalVpnEvents)

	vpn, err := vpnFactory(cfg.Technology)
	if err != nil {
		// if NordWhiser was disabled we'll fall back automatically to NordLynx if autoconnect is enabled or tell user
		// to switch to a different tech
		if !errors.Is(err, ErrNordWhisperDisabled) {
			log.Fatalln(err)
		} else {
			log.Println(internal.ErrorPrefix, "failed to build NordWhisper VPN, it was disabled during compilation")
		}
	}

	devices, err := device.ListPhysical()
	if err != nil {
		log.Fatalln(err)
	}
	ifaceNames := []string{}
	for _, d := range devices {
		ifaceNames = append(ifaceNames, d.Name)
	}

	mesh, err := meshnetImplementation(vpnFactory)
	if err != nil {
		log.Fatalln(err)
	}

	allowlistRouter := routes.NewRouter(
		&norouter.Facade{},
		&netlinkrouter.Router{},
		cfg.Routing.Get(),
	)
	vpnRouter := routes.NewRouter(
		&norouter.Facade{},
		&netlinkrouter.Router{},
		cfg.Routing.Get(),
	)
	meshRouter := routes.NewRouter(
		&norouter.Facade{},
		&netlinkrouter.Router{},
		cfg.Routing.Get(),
	)

	connectionInfo := state.NewConnectionInfo()
	statePublisher := state.NewState()
	internalVpnEvents.Subscribe(connectionInfo)
	connectionInfo.Subscribe(statePublisher)
	daemonEvents.Service.Connect.Subscribe(connectionInfo.ConnectionStatusNotifyConnect)
	daemonEvents.Service.Disconnect.Subscribe(connectionInfo.ConnectionStatusNotifyDisconnect)
	daemonEvents.User.Subscribe(statePublisher)
	configEvents.Subscribe(statePublisher)

	netw := networker.NewCombined(
		vpn,
		mesh,
		gwret,
		infoSubject,
		allowlistRouter,
		dnsSetter,
		fw,
		allowlist.NewAllowlistRouting(func(command string, arg ...string) ([]byte, error) {
			arg = append(arg, "-w", internal.SecondsToWaitForIptablesLock)
			return exec.Command(command, arg...).CombinedOutput()
		}),
		device.ListPhysical,
		routes.NewPolicyRouter(
			&norule.Facade{},
			iprule.NewRouter(
				routes.NewSysctlRPFilterManager(),
				ifgroup.NewNetlinkManager(device.ListPhysical),
				cfg.FirewallMark,
			),
			cfg.Routing.Get(),
		),
		dnsHostSetter,
		vpnRouter,
		meshRouter,
		forwarder.NewForwarder(ifaceNames, func(command string, arg ...string) ([]byte, error) {
			arg = append(arg, "-w", internal.SecondsToWaitForIptablesLock)
			return exec.Command(command, arg...).CombinedOutput()
		},
			kernel.NewSysctlSetter(
				forwarder.Ipv4fwdKernelParamName,
				1,
				0,
			)),
		cfg.FirewallMark,
		cfg.LanDiscovery,
		ipv6.NewIpv6(),
	)

	keygen, err := keygenImplementation(vpnFactory)
	if err != nil {
		log.Fatalln(err)
	}

	var norduserService norduserservice.Service
	if snapconf.IsUnderSnap() {
		norduserService = norduserservice.NewNorduserSnapService()
	} else {
		norduserService = norduserservice.NewChildProcessNorduser()
	}

	norduserClient := norduserservice.NewNorduserGRPCClient()

	meshRegistry := registry.NewNotifyingRegistry(clientAPI, meshnetEvents.PeerUpdate)
	meshnetChecker := meshnet.NewRegisteringChecker(
		fsystem,
		keygen,
		meshRegistry,
	)

	meshMapper := mapper.NewNotifyingMapper(
		mapper.NewCachingMapper(clientAPI, time.Minute*5),
		meshnetEvents.SelfRemoved,
		meshnetEvents.PeerUpdate,
	)

	meshnetEvents.PeerUpdate.Subscribe(refresher.NewMeshnet(
		meshMapper, meshnetChecker, fsystem, netw,
	).NotifyPeerUpdate)

	meshUnsetter := meshunsetter.NewMeshnet(
		fsystem,
		netw,
		errSubject,
		norduserClient,
	)
	meshnetEvents.SelfRemoved.Subscribe(meshUnsetter.NotifyDisabled)

	accountUpdateEvents := daemonevents.NewAccountUpdateEvents()
	accountUpdateEvents.Subscribe(statePublisher)

	trustedPassSessionStore := buildTrustedPassSessionStore(fsystem, invalidSessionErrHandlingReg, clientAPI)
	authChecker := auth.NewRenewingChecker(
		fsystem,
		clientAPI,
		daemonEvents.User.MFA,
		daemonEvents.User.Logout,
		errSubject,
		accountUpdateEvents,
		accessTokenSessionStore, trustedPassSessionStore,
	)

	endpointResolver := network.NewDefaultResolverChain(fw)
	notificationClient := nc.NewClient(
		nc.MqttClientBuilder{},
		infoSubject,
		errSubject,
		meshnetEvents.PeerUpdate,
		nc.NewCredsFetcher(clientAPI, fsystem))

	// on session invalidation (unauthorized access, missing server resources, invalid request)
	// perform user log-out action
	invalidSessionErrHandlingReg.Add(
		func(reason error) {
			discArgs := access.DisconnectInput{
				Networker:                  netw,
				ConfigManager:              fsystem,
				PublishDisconnectEventFunc: daemonEvents.Service.Disconnect.Publish,
			}
			result := access.ForceLogoutWithoutToken(access.ForceLogoutWithoutTokenInput{
				AuthChecker:            authChecker,
				Netw:                   netw,
				NcClient:               notificationClient,
				ConfigManager:          fsystem,
				PublishLogoutEventFunc: daemonEvents.User.Logout.Publish,
				DebugPublisherFunc:     debugSubject.Publish,
				DisconnectFunc:         func() (bool, error) { return access.Disconnect(discArgs) },
			})

			if result.Err != nil {
				log.Println(internal.ErrorPrefix, "logging out on invalid session hook: %w", err)
			}

			if result.Status == internal.CodeSuccess {
				log.Println(internal.DebugPrefix, "successfully logged out after detecting invalid session")
			}
		},
		session.ErrAccessTokenRevoked, core.ErrUnauthorized, core.ErrNotFound, core.ErrBadRequest,
	)

	dataUpdateEvents := daemonevents.NewDataUpdateEvents()
	dataUpdateEvents.Subscribe(statePublisher)
	dm := daemon.NewDataManager(
		daemon.InsightsFilePath,
		daemon.ServersDataFilePath,
		daemon.CountryDataFilePath,
		daemon.VersionFilePath,
		dataUpdateEvents,
	)

	consentChecker := newConsentChecker(
		internal.IsDevEnv(Environment),
		fsystem,
		clientAPI,
		authChecker,
		analytics,
	)
	consentChecker.PrepareDaemonIfConsentNotCompleted()

	sharedContext := sharedctx.New()
	rpc := daemon.NewRPC(
		internal.Environment(Environment),
		authChecker,
		fsystem,
		dm,
		clientAPI,
		clientAPI,
		clientAPI,
		cdnAPI,
		repoAPI,
		core.NewOAuth2(httpClientWithRotator, daemon.BaseURL),
		Version,
		daemonEvents,
		vpnFactory,
		&endpointResolver,
		netw,
		debugSubject,
		threatProtectionLiteServers,
		notificationClient,
		analytics,
		norduserService,
		statePublisher,
		sharedContext,
		rcConfig,
		connectionInfo,
		consentChecker,
	)
	meshService := meshnet.NewServer(
		authChecker,
		fsystem,
		meshnetChecker,
		inviter.NewNotifyingInviter(clientAPI, meshnetEvents.PeerUpdate),
		netw,
		meshRegistry,
		meshMapper,
		threatProtectionLiteServers,
		errSubject,
		daemonEvents,
		norduserClient,
		sharedContext,
	)
	rcConfig.Subscribe(meshService)

	opts := []grpc.ServerOption{
		grpc.Creds(internal.NewUnixSocketCredentials(internal.NewDaemonAuthenticator())),
	}

	norduserMonitor := norduser.NewNorduserProcessMonitor(norduserService)
	go func() {
		if snapconf.IsUnderSnap() {
			if err := norduserMonitor.StartSnap(); err != nil {
				log.Println(internal.ErrorPrefix, "Error when starting norduser monitor for snap:", err.Error())
			}
		} else {
			if err := norduserMonitor.Start(); err != nil {
				log.Println(internal.ErrorPrefix, "Error when starting norduser monitor:", err.Error())
			}
		}
	}()

	middleware := grpcmiddleware.Middleware{}
	if snapconf.IsUnderSnap() {
		checker := snapconf.NewSnapChecker(errSubject)
		middleware.AddStreamMiddleware(checker.StreamInterceptor)
		middleware.AddUnaryMiddleware(checker.UnaryInterceptor)
	} else {
		// in non snap environment, norduser is started on the daemon side on every command
		norduserMiddleware := norduser.NewStartNorduserMiddleware(norduserService)
		middleware.AddStreamMiddleware(norduserMiddleware.StreamMiddleware)
		middleware.AddUnaryMiddleware(norduserMiddleware.UnaryMiddleware)
	}

	clientIDMiddleware := clientid.NewClientIDMiddleware(daemonEvents.Service.UiItemsClick)
	middleware.AddStreamMiddleware(clientIDMiddleware.StreamMiddleware)
	middleware.AddUnaryMiddleware(clientIDMiddleware.UnaryMiddleware)

	opts = append(opts, grpc.StreamInterceptor(middleware.StreamIntercept))
	opts = append(opts, grpc.UnaryInterceptor(middleware.UnaryIntercept))
	s := grpc.NewServer(opts...)

	pb.RegisterDaemonServer(s, rpc)
	meshpb.RegisterMeshnetServer(s, meshService)

	// initialize and register telemetry service with grpc server
	telemetryService := telemetry.New(analytics.OnTelemetry)
	telemetrypb.RegisterTelemetryServiceServer(s, telemetryService)

	// Start jobs
	go func() {
		var (
			listener net.Listener
			err      error
		)
		switch socketType(ConnType) {
		case sockUnix:
			// use systemd listener by default
			listenerFunction := internal.SystemDListener
			// switch to manual if pids mismatch
			if os.Getenv(internal.ListenPID) != strconv.Itoa(os.Getpid()) {
				listenerFunction = internal.ManualListenerIfNotInUse(ConnURL,
					internal.PermUserRWGroupRW, internal.DaemonPid)
			}
			listener, err = listenerFunction()
			if err != nil {
				log.Fatalf("Error on listening to UNIX domain socket: %s\n", err)
			}
			if snapconf.IsUnderSnap() {
				internal.UpdateFilePermissions(
					ConnURL,
					internal.PermUserRWGroupRWOthersRW,
				)
			}
			// limit count of requests on socket at the same time from
			// non-authorized users to prevent from crashing daemon
			listener = netutil.LimitListener(listener, 100)
		case sockTCP:
			listener, err = net.Listen("tcp", ConnURL)
			if err != nil {
				log.Fatalf("Error on listening to TCP %s: %s\n", ConnURL, err)
			}
		default:
			log.Fatalf("Invalid predefined connection type: %s", ConnType)
		}

		if err := s.Serve(listener); err != nil {
			log.Fatalln(err)
		}
	}()

	go func() {
		if err := dm.LoadData(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
	}()
	rpc.StartJobs(statePublisher, heartBeatSubject)
	rpc.StartRemoteConfigLoaderJob(rcConfig)
	meshService.StartJobs()
	rpc.StartKillSwitch()
	if internal.IsSystemd() {
		go rpc.StartSystemShutdownMonitor()
	}

	if cfg.AutoConnect {
		go rpc.StartAutoConnect(network.ExponentialBackoff)
	}

	monitor, err := netstate.NewNetlinkMonitor([]string{openvpn.InterfaceName, nordlynx.InterfaceName})
	if err != nil {
		log.Fatalln(err)
	}
	monitor.Start(netw)

	if ok, _ := authChecker.IsLoggedIn(); ok {
		go daemon.StartNC("[startup]", notificationClient)
	}

	if cfg.Mesh {
		go rpc.StartAutoMeshnet(meshService, network.ExponentialBackoff)
	}

	// Graceful stop

	internal.WaitSignal()
	s.Stop()
	norduserService.StopAll()

	httpCancel()

	if err := notificationClient.Stop(); err != nil {
		log.Println(internal.ErrorPrefix, "stopping NC:", err)
	}
	if _, err := rpc.DoDisconnect(); err != nil {
		log.Println(internal.ErrorPrefix, "disconnecting from VPN:", err)
	}
	if err := netw.UnSetMesh(); err != nil && !errors.Is(err, networker.ErrMeshNotActive) {
		log.Println(internal.ErrorPrefix, "disconnecting from meshnet:", err)
	}
	if err := rpc.StopKillSwitch(); err != nil {
		log.Println(internal.ErrorPrefix, "stopping KillSwitch:", err)
	}
	if err := analytics.Stop(); err != nil {
		log.Println(internal.ErrorPrefix, "stopping analytics:", err)
	}
}

// assignMooseDBPermissions updates moose DB permissions.
// If the file doesn't exist it will be created withe the desired permissions.
func assignMooseDBPermissions(eventsDbPath string) error {
	const permissions os.FileMode = internal.PermUserRWGroupRW

	if !internal.FileExists(eventsDbPath) {
		_, err := internal.FileCreate(eventsDbPath, permissions)
		return err
	}
	// Change permission of the existing DB, because older versions had read for everyone
	if err := os.Chmod(eventsDbPath, permissions); err != nil {
		log.Println(err)
	}

	if gid, err := internal.GetNordvpnGid(); err == nil {
		if err := os.Chown(eventsDbPath, os.Getuid(), gid); err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	return nil
}

func removeIPv6Remains(c config.Config) config.Config {
	// Remove all nameservers with IPv6 addresses
	var dnsList []string
	for _, addr := range c.AutoConnectData.DNS {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if ip.To4() != nil {
			dnsList = append(dnsList, addr)
		}
	}

	c.AutoConnectData.DNS = dnsList

	// Remove all IPv6 subnets from AllowList
	var allowList []string
	for _, addr := range c.AutoConnectData.Allowlist.Subnets {
		_, subnet, err := net.ParseCIDR(addr)
		if err != nil {
			continue
		}

		if subnet.IP.To4() != nil {
			allowList = append(allowList, addr)
		}
	}

	c.AutoConnectData.Allowlist.Subnets = allowList

	return c
}
