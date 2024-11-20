package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) IsQuenchEnabled(ctx context.Context, in *pb.Empty) (*pb.QuenchEnabled, error) {
	if !features.QuenchEnabled {
		return &pb.QuenchEnabled{Enabled: false}, nil
	}
	quenchEnabled, err := r.remoteConfigGetter.GetQuenchEnabled(r.version)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to determine if quench is enabled based on firebase config:", err)
		return &pb.QuenchEnabled{Enabled: false}, nil
	}

	return &pb.QuenchEnabled{Enabled: quenchEnabled}, nil
}
