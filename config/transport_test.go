package config

import (
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
		{TransportUnknown, "UNKNOWN_TRANSPORT"},
		{TransportUDP, "UDP"},
		{TransportTCP, "TCP"},
		{TransportWebTunnel, "Webtunnel"},
		{Transport(42), "42"},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, test.transport.String())
	}
}
