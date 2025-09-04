package daemon

import (
	"context"
	"fmt"

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

	// limit results if value is specified
	if in.Limit != nil {
		limit := int(*in.Limit)
		if limit > 0 && limit < len(values) {
			values = values[:limit]
		}
	}

	rcValues := make([]*pb.RecentConnectionModel, len(values))
	for i, v := range values {
		rcValues[i] = &pb.RecentConnectionModel{
			Country:            v.Country,
			City:               v.City,
			SpecificServer:     v.SpecificServer,
			SpecificServerName: v.SpecificServerName,
			Group:              v.Group,
			CountryCode:        v.CountryCode,
			ConnectionType:     v.ConnectionType,
		}
	}

	return &pb.RecentConnectionsResponse{Connections: rcValues}, nil
}
