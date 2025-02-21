package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

func (r *RPC) GetDaemonApiVersion(_ context.Context, _ *pb.GetDaemonApiVersionRequest) (*pb.GetDaemonApiVersionResponse, error) {
	return &pb.GetDaemonApiVersionResponse{
		ApiVersion: pb.DaemonApiVersion_CURRENT_VERSION,
	}, nil
}
