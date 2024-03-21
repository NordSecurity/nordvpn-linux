package service

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NorduserClient interface {
	StartFileshare(uid uint32) error
	StopFileshare(uid uint32) error
}

type NorduserGRPCClient struct {
}

func NewNorduserGRPCClient() NorduserGRPCClient {
	return NorduserGRPCClient{}
}

func getNorduserClient(uid int) (pb.NorduserClient, error) {
	socket := internal.GetNorduserdSocket(uid)
	if snapconf.IsUnderSnap() {
		socket = internal.GetNorduserSocketSnap(uint32(uid))
	}

	if socket == "" {
		return nil, fmt.Errorf("norduser socket not found")
	}

	url := fmt.Sprintf("%s://%s", internal.Proto, socket)
	norduserConn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dialing norduser socket: %w", err)
	}

	return pb.NewNorduserClient(norduserConn), nil
}

func (n NorduserGRPCClient) StartFileshare(uid uint32) error {
	client, err := getNorduserClient(int(uid))
	if err != nil {
		return fmt.Errorf("getting norduser client: %s", err)
	}

	resp, err := client.StartFileshare(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to start fileshare: %w", err)
	}

	if resp.StartFileshareStatus != pb.StartFileshareStatus_SUCCESS &&
		resp.StartFileshareStatus != pb.StartFileshareStatus_ALREADY_RUNNING {
		return fmt.Errorf("failed to stat fileshare, error code: %d", resp.StartFileshareStatus)
	}

	return nil
}

func (n NorduserGRPCClient) StopFileshare(uid uint32) error {
	client, err := getNorduserClient(int(uid))
	if err != nil {
		return fmt.Errorf("getting norduser client: %s", err)
	}

	resp, err := client.StopFileshare(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to stop fileshare: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("fileshare not stopped")
	}

	return nil
}
