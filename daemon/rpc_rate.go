package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) RateConnection(ctx context.Context, in *pb.RateRequest) (*pb.Payload, error) {
	if r.netw.LastServer().Hostname != "" {
		return &pb.Payload{Type: internal.CodeSuccess}, nil
	}
	return &pb.Payload{Type: internal.CodeNothingToDo}, nil
}
