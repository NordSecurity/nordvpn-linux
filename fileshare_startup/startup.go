package fileshare_startup

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	daemonpb "github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/libdrop"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/fileshare/storage"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

// Values set when building the application
var (
	version   = "0.0.0"
	daemonURL = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
	ConnURL   = internal.GetFilesharedSocket(os.Getuid())
)

var (
	ErrMeshNotEnabled          = errors.New("meshnet not enabled")
	ErrMeshAddressAlreadyInUse = errors.New("meshnet address is already in use, probably another fileshare instance is already running")
)

const transferHistoryChunkSize = 10000

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

type FileshareHandle struct {
	shutdownChan            <-chan struct{}
	eventManager            *fileshare.EventManager
	grpcServer              *grpc.Server
	fileshareImplementation *libdrop.Fileshare
	grpcConn                *grpc.ClientConn
}

// GetShutdownChan provides a way for the gRPC fileshare server to notify the main goroutine about a shutdown triggered
// from the upstream(most likely by the Disable RPC called by the main daemon, when meshent was disabled).
func (f *FileshareHandle) GetShutdownChan() <-chan struct{} {
	return f.shutdownChan
}

// Shutdown performs graceful shutdown
func (f *FileshareHandle) Shutdown() {
	f.eventManager.CancelLiveTransfers()

	f.grpcServer.GracefulStop()

	if err := f.fileshareImplementation.Disable(); err != nil {
		log.Println(internal.ErrorPrefix, "disabling fileshare:", err)
	}
	if err := f.grpcConn.Close(); err != nil {
		log.Println(internal.ErrorPrefix, "closing grpc connection:", err)
	}
}

// Startup contains common parts of the startup fileshare process(daemon or orphan)
func Startup(storagePath string,
	legacyStoragePath string,
	eventsDBPath string,
	serverListener net.Listener,
	environment string,
	grpcAuthenticator internal.SocketAuthenticator) (FileshareHandle, error) {
	// Connection to Meshnet gRPC server
	grpcConn, err := grpc.Dial(
		daemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return FileshareHandle{}, fmt.Errorf("can't connect to daemon: %w", err)
	}

	defer func() {
		if err != nil {
			grpcConn.Close()
		}
	}()

	daemonClient := daemonpb.NewDaemonClient(grpcConn)
	meshClient := meshpb.NewMeshnetClient(grpcConn)

	resp, err := meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return FileshareHandle{}, ErrMeshNotEnabled
	}

	// Libdrop init
	defaultDownloadDirectory, err := fileshare.GetDefaultDownloadDirectory()
	if err != nil {
		log.Println("failed to find default download directory: ", err.Error())
	}

	eventManager := fileshare.NewEventManager(
		internal.IsProdEnv(environment),
		meshClient,
		fileshare.StdOsInfo{},
		fileshare.NewStdFilesystem("/"),
		defaultDownloadDirectory,
	)

	privKeyResponse, err := meshClient.GetPrivateKey(context.Background(), &meshpb.Empty{})
	if err != nil || privKeyResponse.GetPrivateKey() == "" {
		return FileshareHandle{},
			fmt.Errorf("retrieving mesh private key: error: %w, response: %s", err, privKeyResponse.GetPrivateKey())
	}
	meshPrivKey, err := base64.StdEncoding.DecodeString(privKeyResponse.GetPrivateKey())
	if err != nil || len(meshPrivKey) != 32 {
		return FileshareHandle{}, fmt.Errorf("can't decode mesh private key: %w", err)
	}

	if err := internal.EnsureDir(storagePath); err != nil {
		return FileshareHandle{}, fmt.Errorf("failed to ensure dir for transfer history: %w", err)
	}

	fileshareImplementation := libdrop.New(
		eventManager.EventFunc,
		eventsDBPath,
		version,
		internal.IsProdEnv(environment),
		fileshare.NewPubkeyProvider(meshClient).PubkeyFunc,
		string(meshPrivKey),
		storagePath,
	)

	eventManager.SetFileshare(fileshareImplementation)

	if legacyStoragePath != "" {
		eventManager.SetStorage(storage.NewCombined(legacyStoragePath, fileshareImplementation))
	} else {
		eventManager.SetStorage(storage.NewLibdrop(fileshareImplementation))
	}

	settings, err := daemonClient.Settings(context.Background(), &daemonpb.SettingsRequest{
		Uid: int64(os.Getuid()),
	})
	if err != nil {
		return FileshareHandle{}, fmt.Errorf("failed to retrieve daemon setting: %w", err)
	}
	if settings != nil && settings.Data.Notify {
		err = eventManager.EnableNotifications(fileshareImplementation)
		if err != nil {
			return FileshareHandle{}, fmt.Errorf("failed to enable notifications: %w", err)
		}
	}

	meshnetIP, err := firstAddressByInterfaceName(nordlynx.InterfaceName)
	if err != nil {
		return FileshareHandle{}, fmt.Errorf("failed to look up meshnet ip: %w", err)
	}

	err = fileshareImplementation.Enable(meshnetIP)
	if err != nil {
		if errors.Is(err, libdrop.ErrLAddressAlreadyInUse) {
			return FileshareHandle{}, ErrMeshAddressAlreadyInUse
		}
		return FileshareHandle{}, fmt.Errorf("failed to enable libdrop: %w", err)
	}

	shutdownChan := make(chan struct{})

	// Fileshare gRPC server init
	fileshareServer := fileshare.NewServer(fileshareImplementation,
		eventManager,
		meshClient, fileshare.NewStdFilesystem("/"),
		fileshare.StdOsInfo{},
		transferHistoryChunkSize,
		shutdownChan)

	grpcServer := grpc.NewServer()
	if grpcAuthenticator != nil {
		grpcServer = grpc.NewServer(grpc.Creds(internal.NewUnixSocketCredentials(grpcAuthenticator)))
	}

	pb.RegisterFileshareServer(grpcServer, fileshareServer)

	go func() {
		if err := grpcServer.Serve(serverListener); err != nil {
			log.Fatalln(err)
		}
	}()

	return FileshareHandle{
		shutdownChan:            shutdownChan,
		eventManager:            eventManager,
		grpcServer:              grpcServer,
		fileshareImplementation: fileshareImplementation,
		grpcConn:                grpcConn,
	}, nil
}
