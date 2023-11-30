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
	combined.mesh.(*workingMesh).errNetworkChanged = fmt.Errorf("test error")

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
	combined.vpnet.(*mock.WorkingVPN).ErrNetworkChanges = fmt.Errorf("test error")
	err := combined.refreshVPN()
	assert.Error(t, err)

	assert.False(t, combined.isConnectedToVPN())
	assert.True(t, combined.isMeshnetSet)
	assert.NotEmpty(t, combined.fw.(*workingFirewall).rules) // We want to keep rules to avoid leaking
}

func TestNetworkChange(t *testing.T) {
	category.Set(t, category.Unit)

	sharedVPNAndMesh := &mock.MeshnetAndVPN{}
	tests := []struct {
		name string
		vpn  *mock.MeshnetAndVPN
		mesh *mock.MeshnetAndVPN
	}{
		{
			name: "NetworkChange executes once for VPN only",
			vpn:  &mock.MeshnetAndVPN{},
			mesh: nil,
		},
		{
			name: "For VPN+Mesh using same tunnel NetworkChanged executes once",
			vpn:  sharedVPNAndMesh,
			mesh: sharedVPNAndMesh,
		},
		{
			name: "VPN+Mesh with different tunnels NetworkChanged executes once",
			vpn:  &mock.MeshnetAndVPN{},
			mesh: &mock.MeshnetAndVPN{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			combined := GetTestCombined()

			combined.SetVPN(test.vpn)
			combined.mesh = test.mesh

			assert.NoError(t, combined.start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, config.DNS{"1.1.1.1"}))
			assert.True(t, combined.isNetworkSet)

			if test.mesh != nil {
				assert.NoError(t, combined.setMesh(mesh.MachineMap{}, netip.IPv4Unspecified(), ""))
				assert.True(t, combined.isMeshnetSet)
			}

			assert.NoError(t, combined.refreshVPN())
			assert.Equal(t, 1, test.vpn.ExecutionStats[mock.StatsNetworkChange])
			assert.Equal(t, 1, test.vpn.ExecutionStats[mock.StatsStart])
			assert.True(t, combined.isConnectedToVPN())

			if test.mesh != nil {
				assert.Equal(t, 1, test.vpn.ExecutionStats[mock.StatsNetworkChange])
			}
		})
	}
}

func TestFallbackCaseForRefreshVPN(t *testing.T) {
	category.Set(t, category.Unit)

	combined := GetTestCombined()
	vpnet := combined.vpnet.(*mock.WorkingVPN)
	assert.NoError(t, combined.start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, config.DNS{"1.1.1.1"}))
	assert.Equal(t, 1, vpnet.ExecutionStats[mock.StatsStart])

	vpnet.ErrNetworkChanges = mock.ErrOnPurpose

	assert.NoError(t, combined.refreshVPN())
	assert.Equal(t, 1, vpnet.ExecutionStats[mock.StatsNetworkChange])
	assert.Equal(t, 2, vpnet.ExecutionStats[mock.StatsStart])

	assert.True(t, combined.isConnectedToVPN())
}
