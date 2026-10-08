package config

import (
	"encoding/json"
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

func TestTechnology_JSONIsNumeric(t *testing.T) {
	category.Set(t, category.Unit)

	data, err := json.Marshal(TechnologyNordLynx)
	assert.NoError(t, err)
	assert.Equal(t, "2", string(data))

	var decoded Technology
	assert.NoError(t, json.Unmarshal([]byte("1"), &decoded))
	assert.Equal(t, TechnologyOpenVPN, decoded)
}
