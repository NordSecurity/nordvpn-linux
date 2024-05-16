package service

import (
	"context"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser/process"
)

type NorduserFileshareClient interface {
	StartFileshare(uid uint32) error
	StopFileshare(uid uint32) error
}

type NorduserGRPCClient struct {
}

func NewNorduserGRPCClient() NorduserGRPCClient {
	return NorduserGRPCClient{}
}

func (n NorduserGRPCClient) StartFileshare(uid uint32) error {
	clientConn, err := process.GetNorduserClientConnection(int(uid))
	if err != nil {
		return fmt.Errorf("connecting to norduser client: %w", err)
	}

	defer func() {
		if err := clientConn.Close(); err != nil {
			log.Println(internal.ErrorPrefix, "failed to close client connection to nord user: ", err)
		}
	}()

	client := pb.NewNorduserClient(clientConn)
	_, err = client.StartFileshare(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to start fileshare: %w", err)
	}

	return nil
}

func (n NorduserGRPCClient) StopFileshare(uid uint32) error {
	clientConn, err := process.GetNorduserClientConnection(int(uid))
	if err != nil {
		return fmt.Errorf("connecting to norduser client: %w", err)
	}

	defer func() {
		if err := clientConn.Close(); err != nil {
			log.Println(internal.ErrorPrefix, "failed to close client connection to nord user: ", err)
		}
	}()

	client := pb.NewNorduserClient(clientConn)

	_, err = client.StopFileshare(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to stop fileshare: %w", err)
	}

	return nil
}
