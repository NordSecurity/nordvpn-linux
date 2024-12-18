package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) IsNordWhisperEnabled(ctx context.Context, in *pb.Empty) (*pb.NordWhisperEnabled, error) {
	if !features.NordWhisperEnabled {
		return &pb.NordWhisperEnabled{Enabled: false}, nil
	}
	nordWhisperEnabled, err := r.remoteConfigGetter.GetNordWhisperEnabled(r.version)
	if err != nil {
		log.Println(internal.ErrorPrefix,
			"failed to determine if NordWhisper is enabled based on firebase config:", err)
		return &pb.NordWhisperEnabled{Enabled: false}, nil
	}

	return &pb.NordWhisperEnabled{Enabled: nordWhisperEnabled}, nil
}
