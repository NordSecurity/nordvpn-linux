package config

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestGroupDisplayName(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		group    ServerGroup
		expected string
	}{
		{group: ServerGroup_DOUBLE_VPN, expected: "Double VPN"},
		{group: ServerGroup_ONION_OVER_VPN, expected: "Onion Over VPN"},
		{group: ServerGroup_DEDICATED_IP, expected: "Dedicated IP"},
		{group: ServerGroup_STANDARD_VPN_SERVERS, expected: "Standard VPN Servers"},
		{group: ServerGroup_P2P, expected: "P2P"},
		{group: ServerGroup_OBFUSCATED, expected: "Obfuscated Servers"},
		{group: ServerGroup_DEDICATED_SERVER, expected: "Dedicated Server"},
		{group: ServerGroup_UNDEFINED, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.group.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, GroupDisplayName(tt.group))
		})
	}
}

func TestGroupLabels_CoverGroupMap(t *testing.T) {
	category.Set(t, category.Unit)

	for value, group := range GroupMap {
		assert.Contains(t, GroupLabels, group, "group %q has no label", value)
	}
}
