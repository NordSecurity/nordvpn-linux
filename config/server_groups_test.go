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
		{group: ServerGroup_NW_OBFUSCATED, expected: "Obfuscated"},
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

func TestGroupTitleForId(t *testing.T) {
	category.Set(t, category.Unit)

	for title, group := range GroupMap {
		t.Run(title, func(t *testing.T) {
			assert.Equal(t, title, GroupTitleForId(group))
		})
	}

	t.Run("undefined", func(t *testing.T) {
		assert.Equal(t, "", GroupTitleForId(ServerGroup_UNDEFINED))
	})
	t.Run("unknown", func(t *testing.T) {
		assert.Equal(t, "", GroupTitleForId(9999))
	})
}

func TestIsRegionalGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		group    ServerGroup
		expected bool
	}{
		{name: "europe", group: 19, expected: true},
		{name: "the americas", group: 21, expected: true},
		{name: "asia pacific", group: 23, expected: true},
		{name: "africa, the middle east and india", group: 25, expected: true},
		{name: "undefined", group: ServerGroup_UNDEFINED, expected: false},
		{name: "deprecated p2p", group: 15, expected: false},
		{name: "double vpn", group: ServerGroup_DOUBLE_VPN, expected: false},
		{name: "obfuscated", group: ServerGroup_NW_OBFUSCATED, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRegionalGroup(tt.group))
		})
	}
}

func TestIsDeprecatedP2PGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		group    ServerGroup
		expected bool
	}{
		{name: "p2p", group: 15, expected: true},
		{name: "undefined", group: ServerGroup_UNDEFINED, expected: false},
		{name: "regional", group: 19, expected: false},
		{name: "double vpn", group: ServerGroup_DOUBLE_VPN, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsDeprecatedP2PGroup(tt.group))
		})
	}
}
