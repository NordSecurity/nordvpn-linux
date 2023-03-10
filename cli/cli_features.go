package cli

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// IsFeatureSupported is used to find out whether the feature should be provided to user
func (c *cmd) IsFeatureSupported(ctx context.Context, feature config.Feature) bool {
	supported, err := c.client.IsFeatureSupported(ctx, &pb.FeatureRequest{Feature: feature})
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false
	}
	return supported.GetValue()
}
