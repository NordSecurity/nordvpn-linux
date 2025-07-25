package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// Returns the list of feature toggles fetched from the remote config
// TODO (dfe): Use the actual info from the remote config
func (r *RPC) GetFeatureToggles(ctx context.Context, in *pb.Empty) (*pb.FeatureToggles, error) {
	return &pb.FeatureToggles{MeshnetEnabled: true}, nil
}
