package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) Ping(ctx context.Context, in *pb.Empty) (*pb.Payload, error) {
	if r.dm.GetVersionData().newerVersionAvailable {
		return &pb.Payload{
			Type: internal.CodeOutdated,
		}, nil
	}

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
