// NordVPN daemon.
package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
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
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/ipv6"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/meshnet/exitnode"
	"github.com/NordSecurity/nordvpn-linux/meshnet/fork"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/meshnet/registry"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/request/rotator"

	"google.golang.org/grpc"
)

// Values set when building the application
var (
	Salt        = ""
	Version     = ""
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

	fsystem := config.NewFilesystem(
		config.SettingsDataFilePath,
		config.InstallFilePath,
		Salt,
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
		&subs.Subject[events.DataWhitelist]{},
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
	)
	meshnetEvents := meshnet.NewEvents(
		&subs.Subject[[]string]{},
		&subs.Subject[any]{},
	)
	debugSubject := &subs.Subject[string]{}
	infoSubject := &subs.Subject[string]{}
	errSubject := &subs.Subject[error]{}
	httpCalls := &subs.Subject[events.DataRequestAPI]{}

	loggerSubscriber := logger.Subscriber{}
	if internal.Environment(Environment) == internal.Development {
		debugSubject.Subscribe(loggerSubscriber.NotifyMessage)
	}
	infoSubject.Subscribe(loggerSubscriber.NotifyInfo)
	errSubject.Subscribe(loggerSubscriber.NotifyError)

	daemonEvents.Settings.Subscribe(logger.NewSubscriber(true, fsystem))
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

	pkVault := response.NewFilePKVault(internal.DatFilesPath)
	userAgent := fmt.Sprintf("NordApp Linux %s %s", Version, distro.KernelName())
	// simple standard http client with dialer wrapped inside
	httpClientSimple := request.NewStdHTTP()
	cdnAPI := core.NewCDNAPI(
		core.CDNURL,
		userAgent,
		pkVault,
		httpClientSimple,
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
	transports := createTimedOutTransports(resolver, cfg.FirewallMark)
	// subscribe to Connect event for transport(s) to recreate/reconnect
	for _, item := range transports {
		daemonEvents.Service.Connect.Subscribe(item.Transport.NotifyConnect)
	}

	// http client with transport rotator
	httpClientWithRotator := request.NewHTTPClient(httpClientSimple, daemon.BaseURL, debugSubject, nil)
	transportRotator := rotator.NewTransportRotator(httpClientWithRotator, transports)
	httpClientWithRotator.CompleteRotator = transportRotator

	validatorFunc := response.ValidateResponseHeaders
	if !internal.IsProdEnv(Environment) && os.Getenv(EnvIgnoreHeaderValidation) == "1" {
		validatorFunc = func(headers http.Header, body []byte, vault response.PKVault) error {
			return nil
		}
	}

	defaultAPI := core.NewDefaultAPI(
		Version,
		userAgent,
		internal.Environment(Environment),
		pkVault,
		httpClientWithRotator,
		validatorFunc,
		httpCalls,
	)
	repoAPI := daemon.NewRepoAPI(
		daemon.RepoURL,
		Version,
		internal.Environment(Environment),
		PackageType,
		Arch,
		httpClientSimple,
	)
	meshAPI := core.NewMeshAPI(
		daemon.BaseURL,
		userAgent,
		httpClientWithRotator,
		pkVault,
		debugSubject,
	)
	meshAPIex := registry.NewRegistry(
		meshAPI,
		meshnetEvents.SelfRemoved,
	)

	// Networker

	gwret := routes.IPGatewayRetriever{}
	dnsSetter := dns.NewSetter(infoSubject)
	dnsHostSetter := dns.NewHostsFileSetter(dns.HostsFilePath)

	eventsDbPath := fmt.Sprintf("%smoose.db", internal.DatFilesPath)
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

		err = os.Chown(eventsDbPath, os.Getuid(), int(gid))
		if err != nil {
			log.Println(err)
		}
	}

	// obfuscated machineID
	deviceID := fmt.Sprintf("%x", sha256.Sum256([]byte(cfg.MachineID.String()+Salt)))

	remoteConfigGetter := remoteConfigGetterImplementation()

	vpnFactory := getVpnFactory(eventsDbPath, cfg.FirewallMark,
		internal.IsDevEnv(Environment), remoteConfigGetter, deviceID, Version)

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

	whitelistRouter := routes.NewRouter(
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
		whitelistRouter,
		dnsSetter,
		ipv6.NewIpv6(),
		fw,
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
		}),
		cfg.FirewallMark,
	)

	// RPC Servers

	fileshareImplementation := fileshareImplementation()

	keygen, err := keygenImplementation(vpnFactory)
	if err != nil {
		log.Fatalln(err)
	}

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
		fileshareImplementation,
	)
	meshnetEvents.SelfRemoved.Subscribe(meshUnsetter.NotifyDisabled)

	authChecker := auth.NewRenewingChecker(fsystem, defaultAPI)
	endpointResolver := network.NewDefaultResolverChain(fw)
	notificationClient := nc.NewClient(debugSubject, meshnetEvents.PeerUpdate)

	analytics := newAnalytics(eventsDbPath, fsystem, Version, Environment, deviceID)
	if cfg.Analytics.Get() {
		if err := analytics.Enable(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
	}
	daemonEvents.Subscribe(analytics)
	httpCalls.Subscribe(analytics.NotifyRequestAPI)

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
		core.NewOAuth2(httpClientWithRotator),
		Version,
		fw,
		defaultAPI.Client,
		daemonEvents,
		vpnFactory,
		&endpointResolver,
		netw,
		debugSubject,
		threatProtectionLiteServers,
		notificationClient,
		analytics,
		fileshareImplementation,
	)
	meshService := meshnet.NewServer(
		authChecker,
		fsystem,
		meshnetChecker,
		meshAPI,
		netw,
		meshAPIex,
		threatProtectionLiteServers,
		errSubject,
		meshnetEvents.PeerUpdate,
		daemonEvents.Settings.Meshnet,
		fileshareImplementation,
	)

	s := grpc.NewServer(grpc.Creds(internal.UnixSocketCredentials{}))
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
				listenerFunction = internal.ManualListener(ConnURL, internal.PermUserRWGroupRW)
			}
			listener, err = listenerFunction()
			if err != nil {
				log.Fatalf("Error on listening to UNIX domain socket: %s\n", err)
			}
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
	go rpc.StartAutoConnect()

	monitor, err := netstate.NewNetlinkMonitor([]string{openvpn.InterfaceName, nordlynx.InterfaceName})
	if err != nil {
		log.Fatalln(err)
	}
	monitor.Start(netw)

	if authChecker.IsLoggedIn() {
		go daemon.StartNotificationCenter(defaultAPI, notificationClient, fsystem)
	}

	go func() {
		if err := meshService.StartMeshnet(); err != nil && cfg.Mesh {
			log.Println("starting meshnet:", err)
			_, _ = meshService.DisableMeshnet(context.Background(), &meshpb.Empty{})
		}
	}()

	// Graceful stop

	internal.WaitSignal()

	s.GracefulStop()

	if err := dnsSetter.Unset(""); err != nil {
		log.Printf("unsetting dns: %s", err)
	}
	if err := fsystem.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, "loading config:", err)
	} else {
		err := fileshareImplementation.Stop(cfg.Meshnet.EnabledByUID, cfg.Meshnet.EnabledByGID)
		if err != nil && !errors.Is(err, fork.ErrNotStarted) {
			log.Println(internal.ErrorPrefix, "disabling fileshare:", err)
		}
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
