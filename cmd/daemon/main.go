// NordVPN daemon.
package main

import (
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

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/allowlist"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/iptables"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/notables"
	"github.com/NordSecurity/nordvpn-linux/daemon/netstate"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes/iprouter"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes/iprule"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes/norouter"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes/norule"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/distro"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/logger"
	"github.com/NordSecurity/nordvpn-linux/events/meshunsetter"
	"github.com/NordSecurity/nordvpn-linux/events/refresher"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	grpcmiddleware "github.com/NordSecurity/nordvpn-linux/grpc_middleware"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/ipv6"
	"github.com/NordSecurity/nordvpn-linux/kernel"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/meshnet/exitnode"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/meshnet/registry"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	norduserservice "github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"golang.org/x/net/netutil"

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
)

// Environment constants
const (
	EnvKeyPath = "PATH"
	EnvValPath = ":/bin:/sbin:/usr/bin:/usr/sbin"
	// EnvIgnoreHeaderValidation can only be used in `dev` builds. Setting this to `1` makes
	// API client to ignore X-headers. This makes setting up MITM proxies up possible. This
	// should not be used for regular usage.
	EnvIgnoreHeaderValidation = "IGNORE_HEADER_VALIDATION"
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

func main() {
	// pprof
	go func() {
		if internal.IsDevEnv(Environment) {
			// #nosec G114 -- not used in production
			if err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil); err != nil {
				log.Println(internal.ErrorPrefix, err)
			}
		}
	}()

	// Logging

	log.SetOutput(os.Stdout)
	log.Println(internal.InfoPrefix, "Daemon has started")

	// Config
	fsystem := config.NewFilesystemConfigManager(
		config.SettingsDataFilePath,
		config.InstallFilePath,
		Salt,
		config.LinuxMachineIDGetter{},
		config.StdFilesystemHandle{},
	)
	var cfg config.Config
	if err := fsystem.Load(&cfg); err != nil {
		log.Println(err)
		if err := fsystem.Reset(); err != nil {
			log.Fatalln(err)
		}
	}

	// Events

	daemonEvents := daemon.NewEvents(
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[events.DataDNS]{},
		&subs.Subject[bool]{},
		&subs.Subject[config.Protocol]{},
		&subs.Subject[events.DataAllowlist]{},
		&subs.Subject[config.Technology]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[any]{},
		&subs.Subject[events.DataConnect]{},
		&subs.Subject[events.DataDisconnect]{},
		&subs.Subject[any]{},
		&subs.Subject[core.ServicesResponse]{},
		&subs.Subject[events.ServerRating]{},
		&subs.Subject[int]{},
	)
	meshnetEvents := meshnet.NewEvents(
		&subs.Subject[[]string]{},
		&subs.Subject[any]{},
	)
	debugSubject := &subs.Subject[string]{}
	infoSubject := &subs.Subject[string]{}
	errSubject := &subs.Subject[error]{}
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
	daemonEvents.Settings.Publish(cfg)

	// Firewall
	stateModule := "conntrack"
	stateFlag := "--ctstate"
	chainPrefix := ""
	iptablesAgent := iptables.New(
		stateModule,
		stateFlag,
		chainPrefix,
		iptables.FilterSupportedIPTables(internal.GetSupportedIPTables()),
	)
	fw := firewall.NewFirewall(
		&notables.Facade{},
		iptablesAgent,
		debugSubject,
		cfg.Firewall,
	)

	// API
	var validator response.Validator
	var err error
	if !internal.IsProdEnv(Environment) && os.Getenv(EnvIgnoreHeaderValidation) == "1" {
		validator = response.NoopValidator{}
	} else {
		validator, err = response.NewNordValidator()
		if err != nil {
			log.Fatalln("Error on creating validator:", err)
		}
	}

	userAgent := fmt.Sprintf("NordApp Linux %s %s", Version, distro.KernelName())
	// simple standard http client with dialer wrapped inside
	httpClientSimple := request.NewStdHTTP()
	httpClientSimple.Transport = request.NewPublishingRoundTripper(httpClientSimple.Transport, httpCallsSubject)
	cdnAPI := core.NewCDNAPI(
		userAgent,
		core.CDNURL,
		httpClientSimple,
		validator,
	)

	var threatProtectionLiteServers *dns.NameServers
	nameservers, err := cdnAPI.ThreatProtectionLite()
	if err != nil {
		log.Printf("error retrieving nameservers: %s", err)
		threatProtectionLiteServers = dns.NewNameServers(nil)
	} else {
		threatProtectionLiteServers = dns.NewNameServers(nameservers.Servers)
	}

	resolver := network.NewResolver(fw, threatProtectionLiteServers)

	if err := kernel.SetParameter(netCoreRmemMaxKey, netCodeRmemMaxValue); err != nil {
		log.Println(internal.WarningPrefix, err)
	}
	httpClientWithRotator := request.NewStdHTTP()
	httpClientWithRotator.Transport = createTimedOutTransport(resolver, cfg.FirewallMark, httpCallsSubject, daemonEvents.Service.Connect)

	defaultAPI := core.NewDefaultAPI(
		userAgent,
		daemon.BaseURL,
		httpClientWithRotator,
		validator,
	)
	meshAPIex := registry.NewRegistry(
		defaultAPI,
		meshnetEvents.SelfRemoved,
	)

	repoAPI := daemon.NewRepoAPI(
		daemon.RepoURL,
		Version,
		internal.Environment(Environment),
		PackageType,
		Arch,
		httpClientSimple,
	)
	gwret := routes.IPGatewayRetriever{}
	dnsSetter := dns.NewSetter(infoSubject)
	dnsHostSetter := dns.NewHostsFileSetter(dns.HostsFilePath)

	eventsDbPath := filepath.Join(internal.DatFilesPath, "moose.db")
	// TODO: remove once this is fixed: https://github.com/ziglang/zig/issues/11878
	// P.S. this issue does not happen with Zig 0.10.0, but it requires Go 1.19+
	if !internal.FileExists(eventsDbPath) {
		_, err := internal.FileCreate(eventsDbPath, internal.PermUserRWGroupRWOthersR)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		// Previously we created this file only with R permission for group, but fileshare daemon
		// which runs with user permissions also needs to write to it. Need to always rewrite permission
		// because of users updating from older version.
		err = os.Chmod(eventsDbPath, internal.PermUserRWGroupRWOthersR)
		if err != nil {
			log.Println(err)
		}

		gid, err := internal.GetNordvpnGid()
		if err != nil {
			log.Println(err)
		}

		err = os.Chown(eventsDbPath, os.Getuid(), gid)
		if err != nil {
			log.Println(err)
		}
	}

	// obfuscated machineID
	deviceID := fmt.Sprintf("%x", sha256.Sum256([]byte(cfg.MachineID.String()+Salt)))

	analytics := newAnalytics(eventsDbPath, fsystem, Version, Environment, deviceID)
	if cfg.Analytics.Get() {
		if err := analytics.Enable(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
	}
	daemonEvents.Subscribe(analytics)

	vpnLibConfigGetter := vpnLibConfigGetterImplementation(fsystem)

	// Networker
	vpnFactory := getVpnFactory(eventsDbPath, cfg.FirewallMark,
		internal.IsDevEnv(Environment), vpnLibConfigGetter, deviceID, Version)

	vpn, err := vpnFactory(cfg.Technology)
	if err != nil {
		log.Fatalln(err)
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
		&iprouter.Router{},
		cfg.Routing.Get(),
	)
	vpnRouter := routes.NewRouter(
		&norouter.Facade{},
		&iprouter.Router{},
		cfg.Routing.Get(),
	)
	meshRouter := routes.NewRouter(
		&norouter.Facade{},
		&iprouter.Router{},
		cfg.Routing.Get(),
	)

	netw := networker.NewCombined(
		vpn,
		mesh,
		gwret,
		infoSubject,
		allowlistRouter,
		dnsSetter,
		ipv6.NewIpv6(),
		fw,
		allowlist.NewAllowlistRouting(func(command string, arg ...string) ([]byte, error) {
			return exec.Command(command, arg...).CombinedOutput()
		}),
		device.ListPhysical,
		routes.NewPolicyRouter(
			&norule.Facade{},
			iprule.NewRouter(
				routes.NewSysctlRPFilterManager(),
				cfg.FirewallMark,
			),
			cfg.Routing.Get(),
		),
		dnsHostSetter,
		vpnRouter,
		meshRouter,
		exitnode.NewServer(ifaceNames, func(command string, arg ...string) ([]byte, error) {
			return exec.Command(command, arg...).CombinedOutput()
		}, cfg.AutoConnectData.Allowlist,
			kernel.NewSysctlSetter(
				exitnode.Ipv4fwdKernelParamName,
				1,
				0,
			)),
		cfg.FirewallMark,
		cfg.LanDiscovery,
	)

	keygen, err := keygenImplementation(vpnFactory)
	if err != nil {
		log.Fatalln(err)
	}

	var norduserService norduserservice.NorduserService
	if snapconf.IsUnderSnap() {
		norduserService = norduserservice.NewNorduserSnapService()
	} else {
		norduserService = norduserservice.NewChildProcessNorduser()
	}

	norduserClient := norduserservice.NewNorduserGRPCClient()

	meshnetChecker := meshnet.NewRegisteringChecker(
		fsystem,
		keygen,
		meshAPIex,
	)

	meshnetEvents.PeerUpdate.Subscribe(refresher.NewMeshnet(
		meshAPIex, meshnetChecker, fsystem, netw,
	).NotifyPeerUpdate)

	meshUnsetter := meshunsetter.NewMeshnet(
		fsystem,
		netw,
		errSubject,
		norduserClient,
	)
	meshnetEvents.SelfRemoved.Subscribe(meshUnsetter.NotifyDisabled)

	authChecker := auth.NewRenewingChecker(fsystem, defaultAPI)
	endpointResolver := network.NewDefaultResolverChain(fw)
	notificationClient := nc.NewClient(
		nc.MqttClientBuilder{},
		infoSubject,
		errSubject,
		meshnetEvents.PeerUpdate,
		nc.NewCredsFetcher(defaultAPI, fsystem))

	dm := daemon.NewDataManager(
		daemon.InsightsFilePath,
		daemon.ServersDataFilePath,
		daemon.CountryDataFilePath,
		daemon.VersionFilePath,
	)

	rpc := daemon.NewRPC(
		internal.Environment(Environment),
		authChecker,
		fsystem,
		dm,
		defaultAPI,
		defaultAPI,
		defaultAPI,
		cdnAPI,
		repoAPI,
		core.NewOAuth2(httpClientWithRotator, daemon.BaseURL),
		Version,
		fw,
		daemonEvents,
		vpnFactory,
		&endpointResolver,
		netw,
		debugSubject,
		threatProtectionLiteServers,
		notificationClient,
		analytics,
		norduserService,
		meshAPIex,
	)
	meshService := meshnet.NewServer(
		authChecker,
		fsystem,
		meshnetChecker,
		defaultAPI,
		netw,
		meshAPIex,
		threatProtectionLiteServers,
		errSubject,
		meshnetEvents.PeerUpdate,
		daemonEvents.Settings.Meshnet,
		daemonEvents.Service.Connect,
		norduserClient,
	)

	opts := []grpc.ServerOption{
		grpc.Creds(internal.NewUnixSocketCredentials(internal.NewDaemonAuthenticator())),
	}

	norduserMonitor := norduser.NewNordvpnGroupMonitor(norduserService)
	go func() {
		if err := norduserMonitor.Start(); err != nil {
			log.Println("Error when starting norduser monitor: ", err.Error())
		}
	}()

	middleware := grpcmiddleware.Middleware{}
	if snapconf.IsUnderSnap() {
		checker := snapChecker(errSubject)
		middleware.AddStreamMiddleware(checker.StreamInterceptor)
		middleware.AddUnaryMiddleware(checker.UnaryInterceptor)
	} else {
		// in non snap environment, norduser is started on the daemon side on every command
		norduserMiddleware := norduser.NewStartNorduserMiddleware(norduserService)
		middleware.AddStreamMiddleware(norduserMiddleware.StreamMiddleware)
		middleware.AddUnaryMiddleware(norduserMiddleware.UnaryMiddleware)
	}

	opts = append(opts, grpc.StreamInterceptor(middleware.StreamIntercept))
	opts = append(opts, grpc.UnaryInterceptor(middleware.UnaryIntercept))
	s := grpc.NewServer(opts...)

	pb.RegisterDaemonServer(s, rpc)
	meshpb.RegisterMeshnetServer(s, meshService)

	// Start jobs

	go func() {
		var (
			listener net.Listener
			err      error
		)
		switch socketType(ConnType) {
		case sockUnix:
			// use systemd listener by default
			var listenerFunction = internal.SystemDListener
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
	go rpc.StartJobs()
	go meshService.StartJobs()
	rpc.StartKillSwitch()

	if cfg.AutoConnect {
		go rpc.StartAutoConnect(network.ExponentialBackoff)
	}

	monitor, err := netstate.NewNetlinkMonitor([]string{openvpn.InterfaceName, nordlynx.InterfaceName})
	if err != nil {
		log.Fatalln(err)
	}
	monitor.Start(netw)

	if authChecker.IsLoggedIn() {
		go daemon.StartNC("[startup]", notificationClient)
	}

	if cfg.Mesh {
		go rpc.StartAutoMeshnet(meshService, network.ExponentialBackoff)
	}

	// Graceful stop

	internal.WaitSignal()
	s.GracefulStop()
	norduserService.StopAll()

	if err := dnsSetter.Unset(""); err != nil {
		log.Printf("unsetting dns: %s", err)
	}
	if err := notificationClient.Stop(); err != nil {
		log.Println(internal.ErrorPrefix, "stopping NC:", err)
	}
	if err := netw.Stop(); err != nil {
		log.Println(internal.ErrorPrefix, "disconnecting from VPN:", err)
	}
	if err := netw.UnSetMesh(); err != nil && !errors.Is(err, networker.ErrMeshNotActive) {
		log.Println(internal.ErrorPrefix, "disconnecting from meshnet:", err)
	}
	if err := rpc.StopKillSwitch(); err != nil {
		log.Println(internal.ErrorPrefix, "stopping KillSwitch:", err)
	}
}

func snapChecker(publisherErr events.Publisher[error]) *snapconf.ConnChecker {
	return snapconf.NewConnChecker(
		[]snapconf.Interface{
			snapconf.InterfaceNetwork,
			snapconf.InterfaceNetworkBind,
			snapconf.InterfaceNetworkControl,
			snapconf.InterfaceNetworkObserve,
			snapconf.InterfaceFirewallControl,
			snapconf.InterfaceHome,
		},
		[]snapconf.Interface{
			snapconf.InterfaceNetwork,
			snapconf.InterfaceNetworkBind,
			snapconf.InterfaceNetworkControl,
			snapconf.InterfaceNetworkObserve,
			snapconf.InterfaceFirewallControl,
			snapconf.InterfaceHome,
		},
		publisherErr,
	)
}
