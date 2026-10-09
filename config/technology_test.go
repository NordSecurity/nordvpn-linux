package config

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestTechnology_String(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		tech     Technology
		expected string
	}{
		{TechnologyUnknown, "UNKNOWN_TECHNOLOGY"},
		{TechnologyOpenVPN, "OPENVPN"},
		{TechnologyNordLynx, "NORDLYNX"},
		{TechnologyNordWhisper, "NORDWHISPER"},
		{Technology(42), "42"},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, test.tech.String())
	}
}
