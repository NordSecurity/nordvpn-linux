package uievent

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestItemValueFromServerGroupString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		group    string
		expected pb.UIEvent_ItemValue
	}{
		{
			name:     "dedicated server",
			group:    "dedicated_server",
			expected: pb.UIEvent_DEDICATED_SERVER,
		},
		{
			name:     "dedicated ip",
			group:    "dedicated_ip",
			expected: pb.UIEvent_DIP,
		},
		{
			name:     "obfuscated",
			group:    "obfuscated_servers",
			expected: pb.UIEvent_OBFUSCATED,
		},
		{
			name:     "double vpn",
			group:    "double_vpn",
			expected: pb.UIEvent_DOUBLE_VPN,
		},
		{
			name:     "p2p",
			group:    "p2p",
			expected: pb.UIEvent_P2P,
		},
		{
			name:     "empty string",
			group:    "",
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name:     "unknown group",
			group:    "unknown_group",
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name:     "standard vpn has no item value",
			group:    "standard_vpn_servers",
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ItemValueFromServerGroupString(tt.group))
		})
	}
}
