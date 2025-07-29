package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// Returns the list of feature toggles fetched from the remote config
func (r *RPC) GetFeatureToggles(ctx context.Context, in *pb.Empty) (*pb.FeatureToggles, error) {
	meshnetEnabled := r.remoteConfigGetter.IsFeatureEnabled(remote.FeatureMeshnet.String())
	return &pb.FeatureToggles{MeshnetEnabled: meshnetEnabled}, nil
}
