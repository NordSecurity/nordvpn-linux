package mock

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc"
)

type MockDaemonClient struct {
	pb.DaemonClient
	CitiesResponse    []*pb.ServerGroup
	GroupsResponse    []*pb.ServerGroup
	CountriesResponse []*pb.ServerGroup
	PingFn            func() (*pb.PingResponse, error)
}

func (c MockDaemonClient) Cities(ctx context.Context, in *pb.CitiesRequest, opts ...grpc.CallOption) (*pb.ServerGroupsList, error) {
	if c.CitiesResponse != nil {
		return &pb.ServerGroupsList{
			Type:    internal.CodeSuccess,
			Servers: c.CitiesResponse,
		}, nil
	} else {
		return &pb.ServerGroupsList{
			Type:    internal.CodeEmptyPayloadError,
			Servers: nil,
		}, nil
	}
}

func (c MockDaemonClient) Countries(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ServerGroupsList, error) {
	if c.CountriesResponse != nil {
		return &pb.ServerGroupsList{
			Type:    internal.CodeSuccess,
			Servers: c.CountriesResponse,
		}, nil
	} else {
		return &pb.ServerGroupsList{
			Type:    internal.CodeEmptyPayloadError,
			Servers: nil,
		}, nil
	}
}
func (c MockDaemonClient) Groups(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ServerGroupsList, error) {
	if c.GroupsResponse != nil {
		return &pb.ServerGroupsList{
			Type:    internal.CodeSuccess,
			Servers: c.GroupsResponse,
		}, nil
	} else {
		return &pb.ServerGroupsList{
			Type:    internal.CodeEmptyPayloadError,
			Servers: nil,
		}, nil
	}
}

func (c MockDaemonClient) Ping(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.PingResponse, error) {
	return c.PingFn()
}
