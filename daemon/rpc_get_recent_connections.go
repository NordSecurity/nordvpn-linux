package daemon

import (
	"context"
	"fmt"
	"slices"

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

	serverTech := techToServerTech(
		cfg.Technology,
		cfg.AutoConnectData.Protocol,
		cfg.AutoConnectData.Obfuscate)

	var rcValues []*pb.RecentConnectionModel
	// filter by server technology used
	for _, v := range values {
		if slices.Contains(v.ServerTechnologies, serverTech) {
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
	}

	// limit results if value is specified
	limit := int(in.GetLimit())
	if limit > 0 && limit < len(rcValues) {
		rcValues = rcValues[:limit]
	}

	return &pb.RecentConnectionsResponse{Connections: rcValues}, nil
}
