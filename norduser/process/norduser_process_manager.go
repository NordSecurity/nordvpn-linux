package process

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

type NorduserProcessClient struct {
	uid uint32
}

func NewNorduserProcessClient(uid uint32) *NorduserProcessClient {
	return &NorduserProcessClient{
		uid: uid,
	}
}

func GetNorduserClientConnection(uid int) (*grpc.ClientConn, error) {
	socket := internal.GetNorduserdSocket(uid)
	if snapconf.IsUnderSnap() {
		socket = internal.GetNorduserSocketSnap(uid)
	} else if _, err := os.Stat(socket); os.IsNotExist(err) {
		socket = internal.GetNorduserSocketFork(uid)
	}

	if socket == "" {
		return nil, fmt.Errorf("norduser socket not found")
	}

	url := fmt.Sprintf("%s://%s", internal.Proto, socket)
	return grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func (n *NorduserProcessClient) Ping(nowait bool) error {
	clientConn, err := GetNorduserClientConnection(int(n.uid))
	if err != nil {
		return fmt.Errorf("failed to initialize the connection: %w", err)
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			log.Println(internal.ErrorPrefix, "Failed to close client connection after a failed gRPC call: ", err)
		}
	}()

	client := pb.NewNorduserClient(clientConn)
	_, err = client.Ping(context.Background(), &pb.Empty{}, grpc.WaitForReady(!nowait))

	return err
}

func (n *NorduserProcessClient) Stop(disable bool) error {
	clientConn, err := GetNorduserClientConnection(int(n.uid))
	if err != nil {
		return fmt.Errorf("failed to initialize the connection: %w", err)
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			log.Println(internal.ErrorPrefix, "Failed to close client connection after a failed gRPC call: ", err)
		}
	}()

	client := pb.NewNorduserClient(clientConn)
	_, err = client.Stop(context.Background(), &pb.StopNorduserRequest{Disable: disable, Restart: false})

	return err
}

func (n *NorduserProcessClient) Restart() error {
	clientConn, err := GetNorduserClientConnection(int(n.uid))
	if err != nil {
		return fmt.Errorf("failed to initialize the connection: %w", err)
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			log.Println(internal.ErrorPrefix, "Failed to close client connection after a failed gRPC call: ", err)
		}
	}()

	client := pb.NewNorduserClient(clientConn)
	_, err = client.Stop(context.Background(), &pb.StopNorduserRequest{Disable: false, Restart: true})

	return err
}

func NewNorduserGRPCProcessManager(uid uint32) *childprocess.GRPCChildProcessManager {
	return childprocess.NewGRPCChildProcessManager(NewNorduserProcessClient(uid), internal.NorduserdBinaryPath)
}
