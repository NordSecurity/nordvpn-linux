package config

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestProtocol_String(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		protocol Protocol
		expected string
	}{
		{Protocol_UNKNOWN_PROTOCOL, "UNKNOWN_PROTOCOL"},
		{Protocol_UDP, "UDP"},
		{Protocol_TCP, "TCP"},
		{Protocol_Webtunnel, "Webtunnel"},
		{Protocol(42), "42"},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, test.protocol.String())
	}
}

func TestProtocol_JSONIsNumeric(t *testing.T) {
	category.Set(t, category.Unit)

	data, err := json.Marshal(Protocol_TCP)
	assert.NoError(t, err)
	assert.Equal(t, "2", string(data))

	var decoded Protocol
	assert.NoError(t, json.Unmarshal([]byte("1"), &decoded))
	assert.Equal(t, Protocol_UDP, decoded)
}
