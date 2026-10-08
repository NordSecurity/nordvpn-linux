package config

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestTransport_String(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		transport Transport
		expected  string
	}{
		{TransportUnknown, "UNKNOWN_PROTOCOL"},
		{TransportUDP, "UDP"},
		{TransportTCP, "TCP"},
		{TransportWebTunnel, "Webtunnel"},
		{Transport(42), "42"},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, test.transport.String())
	}
}

func TestTransport_JSONIsNumeric(t *testing.T) {
	category.Set(t, category.Unit)

	data, err := json.Marshal(TransportTCP)
	assert.NoError(t, err)
	assert.Equal(t, "2", string(data))

	var decoded Transport
	assert.NoError(t, json.Unmarshal([]byte("1"), &decoded))
	assert.Equal(t, TransportUDP, decoded)
}
