package fileshare_process

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var FileshareURL = fmt.Sprintf("%s://%s", internal.Proto, internal.FileshareSocket)

type FileshareProcessClient struct{}

func NewFileshareProcessClient() *FileshareProcessClient {
	return &FileshareProcessClient{}
}

func getFileshareClient() (pb.FileshareClient, *grpc.ClientConn, error) {
	fileshareConn, err := grpc.Dial(
		FileshareURL,
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
				log.Println("Failed to close client connection after a failed gRPC call: ", err)
			}
		}
	}()

	_, err = client.Ping(context.Background(), &pb.Empty{}, grpc.WaitForReady(!nowait))

	return err
}

func (f *FileshareProcessClient) Stop(bool) error {
	client, clientConn, err := getFileshareClient()
	if err != nil {
		return fmt.Errorf("failed to initialize the connection: %w", err)
	}
	defer func() {
		if clientConn != nil {
			if err := clientConn.Close(); err != nil {
				log.Println("Failed to close client connection after a failed gRPC call: ", err)
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
