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
