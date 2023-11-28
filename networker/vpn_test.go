package networker

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"

	"github.com/stretchr/testify/assert"
)

func TestVPNNetworker_IsVPNActive(t *testing.T) {
	tests := []struct {
		name     string
		vpn      vpn.VPN
		expected bool
	}{
		{
			name:     "active vpn",
			vpn:      mock.ActiveVPN{},
			expected: true,
		},
		{
			name:     "inactive vpn",
			vpn:      mock.WorkingInactiveVPN{},
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
				nil,
				0,
				false,
			)
			// injecting VPN implementation without calling netw.Start
			netw.vpnet = test.vpn
			assert.Equal(t, test.expected, netw.IsVPNActive())
		})
	}
}

func TestRefreshVPN_NotConnected(t *testing.T) {
	category.Set(t, category.Unit)

	combined := GetTestCombined()
	err := combined.refreshVPN()
	assert.NoError(t, err)

	assert.False(t, combined.isConnectedToVPN())
	assert.False(t, combined.isMeshnetSet)
}

func TestRefreshVPN_MeshnetFailure(t *testing.T) {
	category.Set(t, category.Unit)

	combined := GetTestCombined()
	assert.NoError(t, combined.setMesh(mesh.MachineMap{}, netip.IPv4Unspecified(), ""))
	assert.NoError(t, combined.start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, config.DNS{}))

	assert.True(t, combined.isConnectedToVPN())
	assert.True(t, combined.isMeshnetSet)

	combined.mesh.(*workingMesh).enableErr = fmt.Errorf("test error")
	err := combined.refreshVPN()
	assert.Error(t, err)

	assert.True(t, combined.isConnectedToVPN())
	assert.False(t, combined.isMeshnetSet)
}

func TestRefreshVPN_VPNFailure(t *testing.T) {
	category.Set(t, category.Unit)

	combined := GetTestCombined()
	assert.Empty(t, combined.fw.(*workingFirewall).rules)
	assert.NoError(t, combined.setMesh(mesh.MachineMap{}, netip.IPv4Unspecified(), ""))
	assert.NoError(t, combined.start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, config.DNS{}))
	assert.NotEmpty(t, combined.fw.(*workingFirewall).rules)

	assert.True(t, combined.isConnectedToVPN())
	assert.True(t, combined.isMeshnetSet)

	combined.vpnet.(*mock.WorkingVPN).StartErr = fmt.Errorf("test error")
	err := combined.refreshVPN()
	assert.Error(t, err)

	assert.False(t, combined.isConnectedToVPN())
	assert.True(t, combined.isMeshnetSet)
	assert.NotEmpty(t, combined.fw.(*workingFirewall).rules) // We want to keep rules to avoid leaking
}
