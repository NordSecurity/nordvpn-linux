package daemon

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// GetRecentConnections retrieves recent vpn connections from store
func (r *RPC) GetRecentConnections(
	ctx context.Context,
	in *pb.RecentConnectionsRequest,
) (*pb.RecentConnectionsResponse, error) {
	values, err := r.recentVPNConnStore.Get()
	if err != nil {
		return nil, fmt.Errorf("getting recent vpn connections: %w", err)
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		return nil, fmt.Errorf("reading config for recent vpn connections: %w", err)
	}

	returnOnlyObfuscatedRecent := cfg.VPNProtocol.IsNordWhisper()

	var rcValues []*pb.RecentConnectionModel
	// filter by server technology used
	for _, v := range values {
		isObfuscated := v.Group == config.ServerGroup_NW_OBFUSCATED
		if returnOnlyObfuscatedRecent != isObfuscated {
			continue
		}
		if config.IsRegionalGroup(v.Group) {
			continue
		}

		// This is a safe-guard in case the migration failed
		if config.IsDeprecatedP2PGroup(v.Group) {
			v.Group = config.ServerGroup_UNDEFINED
			if v.Country == "" && v.City == "" {
				continue
			}
		}

		item := &pb.RecentConnectionModel{
			Country:            v.Country,
			CountryCode:        v.CountryCode,
			City:               v.City,
			SpecificServer:     v.SpecificServer,
			SpecificServerName: v.SpecificServerName,
			Group:              v.Group,
			ConnectionType:     v.ConnectionType,
		}
		rcValues = append(rcValues, item)
	}

	// limit results if value is specified
	limit := int(in.GetLimit())
	if limit > 0 && limit < len(rcValues) {
		rcValues = rcValues[:limit]
	}

	return &pb.RecentConnectionsResponse{Connections: rcValues}, nil
}
