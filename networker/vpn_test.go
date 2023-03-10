package networker

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	testtunnel "github.com/NordSecurity/nordvpn-linux/test/tunnel"
	"github.com/NordSecurity/nordvpn-linux/tunnel"

	"github.com/stretchr/testify/assert"
)

type activeVPN struct{}

func (activeVPN) Start(
	vpn.Credentials,
	vpn.ServerData,
) error {
	return nil
}
func (activeVPN) State() vpn.State { return vpn.UnknownState }
func (activeVPN) Stop() error      { return nil }
func (activeVPN) Tun() tunnel.T    { return testtunnel.Working{} }
func (activeVPN) IsActive() bool   { return true }

type inactiveVPN struct{}

func (inactiveVPN) Start(
	vpn.Credentials,
	vpn.ServerData,
) error {
	return nil
}
func (inactiveVPN) State() vpn.State { return vpn.UnknownState }
func (inactiveVPN) Stop() error      { return nil }
func (inactiveVPN) Tun() tunnel.T    { return nil }
func (inactiveVPN) IsActive() bool   { return false }

func TestVPNNetworker_IsVPNActive(t *testing.T) {
	tests := []struct {
		name     string
		vpn      vpn.VPN
		expected bool
	}{
		{
			name:     "active vpn",
			vpn:      activeVPN{},
			expected: true,
		},
		{
			name:     "inactive vpn",
			vpn:      inactiveVPN{},
			expected: false,
		},
		{
			name:     "nil vpn",
			vpn:      nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test does not rely on any of the values provided via constructor
			// so it's fine to pass nils to all of them.
			netw := NewCombined(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
			)
			// injecting VPN implementation without calling netw.Start
			netw.vpnet = test.vpn
			assert.Equal(t, test.expected, netw.IsVPNActive())
		})
	}
}
