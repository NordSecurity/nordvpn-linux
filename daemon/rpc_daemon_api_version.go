package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

func (r *RPC) GetDaemonApiVersion(_ context.Context, _ *pb.GetDaemonApiVersionRequest) (*pb.GetDaemonApiVersionResponse, error) {
	return &pb.GetDaemonApiVersionResponse{
		ApiVersion: uint32(pb.DaemonApiVersion_CURRENT_VERSION.Number()), // #nosec G115 - operation on an enum within a valid range
	}, nil
}
