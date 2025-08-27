package daemon

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// GetRecentConnections retrieves recent vpn connections from store
func (r *RPC) GetRecentConnections(
	ctx context.Context,
	in *pb.RecentConnectionsRequest,
) (*pb.RecentConnectionsResponse, error) {
	values, err := r.recentVPNConnStore.Get()
	if err != nil {
		return nil, fmt.Errorf("%s getting recent vpn connections: %w", internal.ErrorPrefix, err)
	}

	if in.Limit > 0 && int(in.Limit) < len(values) {
		values = values[:in.Limit]
	}

	rcValues := make([]*pb.RecentConnectionModel, len(values))
	for i, v := range values {
		rcValues[i] = &pb.RecentConnectionModel{
			Country:            v.Country,
			City:               v.City,
			SpecificServer:     v.SpecificServer,
			SpecificServerName: v.SpecificServerName,
			Group:              v.Group,
			ConnectionType:     v.ConnectionType,
		}
	}

	return &pb.RecentConnectionsResponse{Connections: rcValues}, nil
}
