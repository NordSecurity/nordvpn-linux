// NordVPN fileshare daemon.
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"net/netip"
	"os"
	"os/user"
	"path"
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
	Version     = ""
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
	daemonClient := daemonpb.NewDaemonClient(grpcConn)
	meshClient := meshpb.NewMeshnetClient(grpcConn)

	resp, err := meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil {
		log.Fatalf("can't check if meshnet is enabled: %s", err)
	}
	if !resp.GetValue() {
		log.Fatalf("meshnet is not enabled")
	}

	// Libdrop init
	defaultDownloadDirectory, err := fileshare.GetDefaultDownloadDirectory()
	if err != nil {
		log.Println("failed to find default download directory: ", err.Error())
	}

	eventManager := fileshare.NewEventManager(
		internal.IsProdEnv(Environment),
		meshClient,
		fileshare.StdOsInfo{},
		fileshare.NewStdFilesystem("/"),
		defaultDownloadDirectory,
	)

	privKeyResponse, err := meshClient.GetPrivateKey(context.Background(), &meshpb.Empty{})
	if err != nil || privKeyResponse.GetPrivateKey() == "" {
		log.Fatalf("can't retrieve mesh private key: %v; service error %v", err, privKeyResponse.GetServiceErrorCode())
	}
	meshPrivKey, err := base64.StdEncoding.DecodeString(privKeyResponse.GetPrivateKey())
	if err != nil || len(meshPrivKey) != 32 {
		log.Fatalf("can't decode mesh private key: %v", err)
	}

	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("can't retrieve current user info: %s", err)
	}
	// we have to hardcode config directory, using os.UserConfigDir is not viable as nordfileshared
	// is spawned by nordvpnd(owned by root) and inherits roots environment variables
	storagePath := path.Join(
		currentUser.HomeDir,
		internal.ConfigDirectory,
		internal.UserDataPath,
		internal.FileshareHistoryFile,
	)
	if err := internal.EnsureDir(storagePath); err != nil {
		log.Fatalf("ensuring dir for transfer history file: %s", err)
	}
	eventsDbPath := fmt.Sprintf("%smoose.db", internal.DatFilesPath)

	fileshareImplementation := drop.New(
		eventManager.EventFunc,
		eventsDbPath,
		Version,
		internal.IsProdEnv(Environment),
		fileshare.NewPubkeyProvider(meshClient).PubkeyFunc,
		string(meshPrivKey),
		storagePath,
	)
	eventManager.SetFileshare(fileshareImplementation)
	legacyStoragePath := path.Join(currentUser.HomeDir, internal.ConfigDirectory, internal.UserDataPath)
	eventManager.SetStorage(fileshare.NewCombinedStorage(legacyStoragePath, fileshareImplementation))

	settings, err := daemonClient.Settings(context.Background(), &daemonpb.SettingsRequest{
		Uid: int64(os.Getuid()),
	})
	if err != nil {
		log.Printf("retrieving daemon settings: %s", err)
	}
	if settings != nil && settings.Data.Notify {
		err = eventManager.EnableNotifications(fileshareImplementation)
		if err != nil {
			log.Println("failed to enable notifications: ", err)
		}
	}

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
