package config

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestMigrateTPL(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name     string
		data     string
		initial  bool
		expected bool
	}{
		{
			name:     "legacy cybersec true is migrated to RTP",
			data:     `{"auto_connect_data":{"cybersec":true}}`,
			initial:  false,
			expected: true,
		},
		{
			name:     "legacy cybersec false is migrated to RTP",
			data:     `{"auto_connect_data":{"cybersec":false}}`,
			initial:  true,
			expected: false,
		},
		{
			name:     "absent cybersec does not override an already migrated RTP",
			data:     `{"auto_connect_data":{"realtimeprotection":true}}`,
			initial:  true,
			expected: true,
		},
		{
			name:     "absent cybersec leaves the current value untouched",
			data:     `{"auto_connect_data":{}}`,
			initial:  true,
			expected: true,
		},
		{
			name:     "invalid json is ignored",
			data:     `not json`,
			initial:  true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{AutoConnectData: AutoConnectData{RealTimeProtection: tt.initial}}
			migrateTPL(c, []byte(tt.data))
			assert.Equal(t, tt.expected, c.AutoConnectData.RealTimeProtection)
		})
	}
}
