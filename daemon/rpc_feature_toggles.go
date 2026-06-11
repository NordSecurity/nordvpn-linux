package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// Returns the list of feature toggles fetched from the remote config
func (r *RPC) GetFeatureToggles(ctx context.Context, in *pb.Empty) (*pb.FeatureToggles, error) {
	meshnetEnabled := r.remoteConfigGetter.IsFeatureEnabled(remote.FeatureMeshnet)
	ensEnabled := r.remoteConfigGetter.IsFeatureEnabled(remote.FeatureENS)

	return &pb.FeatureToggles{
		MeshnetEnabled: meshnetEnabled,
		EnsEnabled:     ensEnabled,
	}, nil
}
