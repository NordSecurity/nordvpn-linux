package process

import (
	"context"
	"fmt"
	"log"
	"sync"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NorduserURL(uid uint32) string {
	return fmt.Sprintf("%s://%s", internal.Proto, internal.GetNorduserSocketSnap(uid))
}

type NorduserProcessClient struct {
	mu  sync.Mutex
	uid uint32
}

func NewNorduserProcessClient(uid uint32) *NorduserProcessClient {
	return &NorduserProcessClient{
		uid: uid,
	}
}

func getNorduserClient(uid uint32) (pb.NorduserClient, *grpc.ClientConn, error) {
	norduserConn, err := grpc.Dial(
		NorduserURL(uid),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to fileshare daemon: %w", err)
	}

	client := pb.NewNorduserClient(norduserConn)
	return client, norduserConn, nil
}

func (n *NorduserProcessClient) Ping(nowait bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	client, clientConn, err := getNorduserClient(n.uid)
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

func (n *NorduserProcessClient) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	client, clientConn, err := getNorduserClient(n.uid)
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

func NewNorduserGRPCProcessManager(uid uint32) *childprocess.GRPCChildProcessManager {
	return childprocess.NewGRPCChildProcessManager(NewNorduserProcessClient(uid), internal.NorduserBinaryPath)
}
