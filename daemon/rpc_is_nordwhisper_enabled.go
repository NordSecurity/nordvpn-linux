package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/features"
)

func (r *RPC) isNordWhisperEnabled() bool {
	nordWhisperEnabled, err := r.remoteConfigGetter.GetNordWhisperEnabled(r.version)
	if err != nil {
		log.Println("failed to determine if NordWhisper is enabled:", err)
		return false
	}

	return features.NordWhisperEnabled && nordWhisperEnabled
}

func (r *RPC) IsNordWhisperEnabled(ctx context.Context, in *pb.Empty) (*pb.NordWhisperEnabled, error) {
	return &pb.NordWhisperEnabled{Enabled: r.isNordWhisperEnabled()}, nil
}
