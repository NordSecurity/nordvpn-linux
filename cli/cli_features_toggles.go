package cli

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
)

type FeatureToggles struct {
	meshnetEnabled bool
}

func defaultToggles() FeatureToggles {
	return FeatureToggles{meshnetEnabled: true}
}

func (c *cmd) GetFeatureToggles() FeatureToggles {
	featureToggles, err := c.client.GetFeatureToggles(context.Background(), &pb.Empty{})
	if err != nil {
		log.Error("failed to get the feature toggles", err)
		return defaultToggles()
	}

	return FeatureToggles{meshnetEnabled: featureToggles.GetMeshnetEnabled()}
}
