package fileshare_process

import (
	"context"
	"fmt"
	"os"

	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/snapconf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func GetFileshareURL() string {
	var fileshareSocketPath string
	if snapconf.IsUnderSnap() {
		fileshareSocketPath = internal.GetFileshareSocketSnap()
	} else {
		uid := os.Getuid()
		fileshareSocketPath = internal.GetFileshareSocketFork(uid)
	}

	return fmt.Sprintf("%s://%s", internal.Proto, fileshareSocketPath)
}

type FileshareProcessClient struct{}

func NewFileshareProcessClient() *FileshareProcessClient {
	return &FileshareProcessClient{}
}

func getFileshareClient() (pb.FileshareClient, *grpc.ClientConn, error) {
	//nolint:staticcheck
	fileshareConn, err := grpc.NewClient(
		GetFileshareURL(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, nil, fmt.Errorf("connecting to fileshare daemon: %w", err)
	}

	client := pb.NewFileshareClient(fileshareConn)
	return client, fileshareConn, nil
}

func (f *FileshareProcessClient) Ping(nowait bool) error {
	client, clientConn, err := getFileshareClient()
	if err != nil {
		return fmt.Errorf("failed to initialize the connection: %w", err)
	}
	defer func() {
		if clientConn != nil {
			if err := clientConn.Close(); err != nil {
				log.Error("Failed to close client connection after a failed gRPC call: ", err)
			}
		}
	}()

	_, err = client.Ping(context.Background(), &pb.Empty{}, grpc.WaitForReady(!nowait))

	return err
}

func (f *FileshareProcessClient) Stop(bool) error {
	var socketPath string
	if snapconf.IsUnderSnap() {
		socketPath = internal.GetFileshareSocketSnap()
	} else {
		socketPath = internal.GetFileshareSocketFork(os.Getuid())
	}

	// There are cases when the fileshare has already been stopped when meshnet was disabled
	// We don't want to try stop it again
	if !internal.FileExists(socketPath) {
		log.Info("Fileshare has already been stopped")
		return nil
	}

	client, clientConn, err := getFileshareClient()
	if err != nil {
		return fmt.Errorf("failed to initialize the connection: %w", err)
	}
	defer func() {
		if clientConn != nil {
			if err := clientConn.Close(); err != nil {
				log.Error("Failed to close client connection after a failed gRPC call: ", err)
			}
		}
	}()

	_, err = client.Stop(context.Background(), &pb.Empty{})

	return err
}

func (f *FileshareProcessClient) Restart() error {
	return nil
}

func NewFileshareGRPCProcessManager() *childprocess.GRPCChildProcessManager {
	return childprocess.NewGRPCChildProcessManager(NewFileshareProcessClient(), internal.FileshareBinaryPath)
}
