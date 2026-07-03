package fileshare_startup

import (
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

const transferHistoryChunkSize = 10000
const disableTimeout = 5 * time.Second

type FileshareHandle struct {
	shutdownChan            <-chan struct{}
	eventManager            *fileshare.EventManager
	grpcServer              *grpc.Server
	fileshareImplementation fileshare.Fileshare
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

	f.grpcServer.Stop()

	// Disable() is an FFI call into libdrop that can hang
	done := make(chan error, 1)
	go func() {
		done <- f.fileshareImplementation.Disable()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Error("disabling fileshare:", err)
		}
	case <-time.After(disableTimeout):
		log.Error("fileshare Disable() timed out after", disableTimeout)
	}

	if err := f.grpcConn.Close(); err != nil {
		log.Error("closing grpc connection:", err)
	}
}

// Startup contains common parts of the startup fileshare process(daemon or orphan)
func Startup(storagePath string,
	serverListener net.Listener,
	grpcAuthenticator internal.SocketAuthenticator,
	fileshareImpl fileshare.Fileshare,
	eventManager *fileshare.EventManager,
	meshClient meshpb.MeshnetClient,
	grpcConn *grpc.ClientConn,
) FileshareHandle {
	shutdownChan := make(chan struct{})

	// Fileshare gRPC server init
	fileshareServer := fileshare.NewServer(fileshareImpl,
		eventManager,
		meshClient,
		fileshare.NewStdFilesystem("/"),
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
			log.Fatal(err)
		}
	}()

	return FileshareHandle{
		shutdownChan:            shutdownChan,
		eventManager:            eventManager,
		grpcServer:              grpcServer,
		fileshareImplementation: fileshareImpl,
		grpcConn:                grpcConn,
	}
}
