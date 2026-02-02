package tray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestItemValueFromServerGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		group    config.ServerGroup
		expected pb.UIEvent_ItemValue
	}{
		{"dedicated ip", config.ServerGroup_DEDICATED_IP, pb.UIEvent_DIP},
		{"obfuscated", config.ServerGroup_OBFUSCATED, pb.UIEvent_OBFUSCATED},
		{"onion over vpn", config.ServerGroup_ONION_OVER_VPN, pb.UIEvent_ONION_OVER_VPN},
		{"double vpn", config.ServerGroup_DOUBLE_VPN, pb.UIEvent_DOUBLE_VPN},
		{"p2p", config.ServerGroup_P2P, pb.UIEvent_P2P},
		{"standard vpn", config.ServerGroup_STANDARD_VPN_SERVERS, pb.UIEvent_ITEM_VALUE_UNSPECIFIED},
		{"undefined", config.ServerGroup_UNDEFINED, pb.UIEvent_ITEM_VALUE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ItemValueFromServerGroup(tt.group))
		})
	}
}

func TestItemValueFromRecentConnection(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		conn     *RecentConnection
		expected pb.UIEvent_ItemValue
	}{
		{
			name:     "nil connection",
			conn:     nil,
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name: "city connection",
			conn: &RecentConnection{
				ConnectionType: config.ServerSelectionRule_CITY,
				Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
			},
			expected: pb.UIEvent_CITY,
		},
		{
			name: "country connection",
			conn: &RecentConnection{
				ConnectionType: config.ServerSelectionRule_COUNTRY,
				Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
			},
			expected: pb.UIEvent_COUNTRY,
		},
		{
			name: "specialty group takes priority over connection type",
			conn: &RecentConnection{
				ConnectionType: config.ServerSelectionRule_CITY,
				Group:          config.ServerGroup_DOUBLE_VPN,
			},
			expected: pb.UIEvent_DOUBLE_VPN,
		},
		{
			name: "dedicated ip",
			conn: &RecentConnection{
				ConnectionType: config.ServerSelectionRule_GROUP,
				Group:          config.ServerGroup_DEDICATED_IP,
			},
			expected: pb.UIEvent_DIP,
		},
		{
			name: "specific server defaults to unspecified",
			conn: &RecentConnection{
				ConnectionType: config.ServerSelectionRule_SPECIFIC_SERVER,
				Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
			},
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ItemValueFromRecentConnection(tt.conn))
		})
	}
}

func TestAttachUIEventMetadata(t *testing.T) {
	category.Set(t, category.Unit)

	ctx := attachUIEventMetadata(
		t.Context(),
		pb.UIEvent_CONNECT,
		pb.UIEvent_COUNTRY,
	)

	md, ok := metadata.FromOutgoingContext(ctx)
	assert.True(t, ok)
	assert.NotEmpty(t, md.Get("ui-form-reference"))
	assert.NotEmpty(t, md.Get("ui-item-name"))
	assert.NotEmpty(t, md.Get("ui-item-type"))
	assert.NotEmpty(t, md.Get("ui-item-value"))

	// Verify TRAY form reference (enum value 2)
	assert.Equal(t, "2", md.Get("ui-form-reference")[0])
	// Verify CONNECT item name (enum value 1)
	assert.Equal(t, "1", md.Get("ui-item-name")[0])
	// Verify CLICK item type (enum value 1)
	assert.Equal(t, "1", md.Get("ui-item-type")[0])
	// Verify COUNTRY item value (enum value 1)
	assert.Equal(t, "1", md.Get("ui-item-value")[0])
}
