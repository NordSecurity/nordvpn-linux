// NordVPN fileshare daemon.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"net/netip"
	"os"
	"strconv"

	daemonpb "github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/drop"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Values set when building the application
var (
	Environment = ""
	PprofPort   = 6961
	ConnURL     = internal.GetFilesharedSocket(os.Getuid())
	DaemonURL   = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
)

func main() {
	// Pprof
	go func() {
		if internal.IsDevEnv(Environment) {
			// #nosec G114 -- not used in production
			if err := http.ListenAndServe(fmt.Sprintf(":%d", PprofPort), nil); err != nil {
				log.Println(internal.ErrorPrefix, err)
			}
		}
	}()

	// Logging

	log.SetOutput(os.Stdout)
	log.Println(internal.InfoPrefix, "Daemon has started")

	// Connection to Meshnet gRPC server

	grpcConn, err := grpc.Dial(
		DaemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("can't connect to daemon: %s", err)
	}
	meshClient := meshpb.NewMeshnetClient(grpcConn)

	resp, err := meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil {
		log.Fatalf("can't check if meshnet is enabled: %s", err)
	}
	if !resp.GetValue() {
		log.Fatalf("meshnet is not enabled")
	}

	// Libdrop init

	daemonClient := daemonpb.NewDaemonClient(grpcConn)

	settings, err := daemonClient.Settings(context.Background(), &daemonpb.SettingsRequest{
		Uid: int64(os.Getuid()),
	})

	var notificationManager *fileshare.NotificationManager
	if err == nil {
		if settings.Data.Notify {
			notificationManager, err = fileshare.NewNotificationManager()
			if err != nil {
				log.Println("failed to initialize notification manager: ", err)
			}
		}
	} else {
		log.Println("failed to determine status notifications setting: ", err)
	}

	eventsDbPath := fmt.Sprintf("%smoose.db", internal.DatFilesPath)
	eventManager := fileshare.NewEventManager(fileshare.FileshareHistoryImplementation(), meshClient, notificationManager)
	fileshareImplementation := drop.New(
		eventManager.EventFunc,
		eventsDbPath,
		internal.IsProdEnv(Environment),
	)
	eventManager.CancelFunc = fileshareImplementation.Cancel

	meshnetIP, err := firstAddressByInterfaceName(nordlynx.InterfaceName)
	if err != nil {
		log.Fatalf("looking up meshnet ip: %s", err)
	}

	err = fileshareImplementation.Enable(meshnetIP)
	if err != nil {
		log.Fatalf("enabling libdrop: %s", err)
	}

	// Fileshare gRPC server init
	fileshareServer := fileshare.NewServer(fileshareImplementation, eventManager, meshClient, fileshare.NewStdFilesystem("/"), fileshare.StdOsInfo{})
	grpcServer := grpc.NewServer()
	pb.RegisterFileshareServer(grpcServer, fileshareServer)

	var listenerFunction = internal.SystemDListener
	if os.Getenv(internal.ListenPID) != strconv.Itoa(os.Getpid()) {
		listenerFunction = internal.ManualListener(ConnURL, internal.PermUserRWX)
	}
	listener, err := listenerFunction()
	if err != nil {
		log.Fatalf("Error on listening to UNIX domain socket: %s\n", err)
	}

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalln(err)
		}
	}()

	// Teardown

	internal.WaitSignal()

	grpcServer.GracefulStop()

	if err := fileshareImplementation.Disable(); err != nil {
		log.Println(internal.ErrorPrefix, "disabling fileshare:", err)
	}
	if err := grpcConn.Close(); err != nil {
		log.Println(internal.ErrorPrefix, "closing grpc connection:", err)
	}
}

func firstAddressByInterfaceName(name string) (netip.Addr, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("interface not found: %w", err)
	}

	ips, err := iface.Addrs()
	if err != nil || len(ips) == 0 {
		return netip.Addr{}, fmt.Errorf("interface is missing ips: %w", err)
	}

	ip, err := netip.ParsePrefix(ips[0].String())
	if err != nil {
		return netip.Addr{}, fmt.Errorf("invalid ip format: %w", err)
	}

	return ip.Addr(), nil
}
