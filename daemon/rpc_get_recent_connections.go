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

	rcValues := make([]*pb.RecentConnection, len(values))
	for i, v := range values {
		conn := &v.Connection
		rcValues[i] = &pb.RecentConnection{
			ConnectionModel: &pb.RecentConnectionModel{
				Country:            conn.Country,
				City:               conn.City,
				SpecificServer:     conn.SpecificServer,
				SpecificServerName: conn.SpecificServerName,
				Group:              conn.Group,
				ConnectionType:     pb.ServerSelectionRule(conn.ConnectionType),
			},
			DisplayLabel: v.DisplayLabel,
		}
	}

	return &pb.RecentConnectionsResponse{Connections: rcValues}, nil
}
