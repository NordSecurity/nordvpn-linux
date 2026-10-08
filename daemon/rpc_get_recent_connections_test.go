package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestGetRecentConnections_Filtering(t *testing.T) {
	category.Set(t, category.Unit)
	r := testRPCLocal(t)

	r.recentVPNConnStore.Add(recents.Model{
		Country: "France",
	})
	r.recentVPNConnStore.Add(recents.Model{
		Country: "Lithuania",
	})

	resp, err := r.GetRecentConnections(context.Background(), &pb.RecentConnectionsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Connections, 2)
	assert.Equal(t, "Lithuania", resp.Connections[0].Country)
	assert.Equal(t, "France", resp.Connections[1].Country)
}

func TestGetRecentConnections_FiltersDeprecatedRegionalGroups(t *testing.T) {
	category.Set(t, category.Unit)
	r := testRPCLocal(t)

	r.recentVPNConnStore.Add(recents.Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	})
	r.recentVPNConnStore.Add(recents.Model{
		Group:          regionalGroupEurope,
		ConnectionType: config.ServerSelectionRule_GROUP,
	})
	r.recentVPNConnStore.Add(recents.Model{
		Country:        "Germany",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	})

	resp, err := r.GetRecentConnections(context.Background(), &pb.RecentConnectionsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Connections, 2)
	for _, c := range resp.Connections {
		assert.False(t, config.IsRegionalGroup(c.Group), "regional group leaked into response: %v", c)
	}
}

func TestGetRecentConnections_Limit(t *testing.T) {
	category.Set(t, category.Unit)
	r := testRPCLocal(t)

	r.recentVPNConnStore.Add(recents.Model{
		Country: "France",
	})
	r.recentVPNConnStore.Add(recents.Model{
		Country: "Germany",
	})
	r.recentVPNConnStore.Add(recents.Model{
		Country: "Lithuania",
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
