package config

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoConnectDataMarshalJSONUsesAllowlistKey(t *testing.T) {
	category.Set(t, category.Unit)

	data, err := json.Marshal(AutoConnectData{Allowlist: NewAllowlist(nil, []int64{22}, nil)})
	require.NoError(t, err)
	assert.Contains(t, string(data), `"allowlist":`)
	assert.NotContains(t, string(data), `"whitelist"`)
}

func TestServerTagFromAutoConnectData(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    AutoConnectData
		expected string
	}{
		{name: "already snake_case", input: AutoConnectData{Country: "united_states", City: "new_york"}, expected: "united_states new_york"},
		{name: "mixed case with spaces", input: AutoConnectData{Country: "United States", City: "New York"}, expected: "united_states new_york"},
		{name: "single words", input: AutoConnectData{Country: "Italy", City: "Rome"}, expected: "italy rome"},
		{name: "missing city", input: AutoConnectData{Country: "Italy"}, expected: "italy"},
		{name: "missing country", input: AutoConnectData{City: "Rome"}, expected: "rome"},
		{name: "both empty", input: AutoConnectData{}, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ServerTagFromAutoconnectData(tt.input), tt.name)
		})
	}
}
