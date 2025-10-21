package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/stretchr/testify/assert"
)

func TestGetRecentConnections_Filtering(t *testing.T) {
	r := testRPCLocal(t)

	// Enable virtual locations
	r.cm.SaveWith(func(c config.Config) config.Config {
		c.VirtualLocation.Set(true)
		return c
	})

	r.recentVPNConnStore.Add(recents.Model{
		Country:            "France",
		IsVirtual:          false,
		ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
	})
	r.recentVPNConnStore.Add(recents.Model{
		Country:            "Lithuania",
		IsVirtual:          true,
		ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
	})

	resp, err := r.GetRecentConnections(context.Background(), &pb.RecentConnectionsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Connections, 2)
	assert.Equal(t, "Lithuania", resp.Connections[0].Country)
	assert.True(t, resp.Connections[0].IsVirtual)
	assert.Equal(t, "France", resp.Connections[1].Country)
	assert.False(t, resp.Connections[1].IsVirtual)

	// Disable virtual locations
	r.cm.SaveWith(func(c config.Config) config.Config {
		c.VirtualLocation.Set(false)
		return c
	})

	resp, err = r.GetRecentConnections(context.Background(), &pb.RecentConnectionsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Connections, 1)
	assert.Equal(t, "France", resp.Connections[0].Country)
	assert.False(t, resp.Connections[0].IsVirtual)
}

func TestGetRecentConnections_Limit(t *testing.T) {
	r := testRPCLocal(t)

	r.recentVPNConnStore.Add(recents.Model{
		Country:            "France",
		ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
	})
	r.recentVPNConnStore.Add(recents.Model{
		Country:            "Germany",
		ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
	})
	r.recentVPNConnStore.Add(recents.Model{
		Country:            "Lithuania",
		ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
	})

	// Limit to 2
	limit := int64(2)
	resp, err := r.GetRecentConnections(context.Background(), &pb.RecentConnectionsRequest{Limit: &limit})
	assert.NoError(t, err)
	assert.Len(t, resp.Connections, 2)
	assert.Equal(t, "Lithuania", resp.Connections[0].Country)
	assert.Equal(t, "Germany", resp.Connections[1].Country)

	// Limit 0 means no limit
	limit = int64(0)
	resp, err = r.GetRecentConnections(context.Background(), &pb.RecentConnectionsRequest{Limit: &limit})
	assert.NoError(t, err)
	assert.Len(t, resp.Connections, 3)

	// No limit
	resp, err = r.GetRecentConnections(context.Background(), &pb.RecentConnectionsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Connections, 3)
}