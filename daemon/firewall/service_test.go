package firewall

import (
	"math/rand"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConfigOptions(t *testing.T) {
	category.Set(t, category.Unit)

	allowlist := config.NewAllowlist([]int64{1}, []int64{101}, []string{"123.456.123.1"})
	meshInfo := NewMeshInfo(mesh.MachineMap{}, "nordlynx")
	fullConfig := Config{
		TunnelInterface: "test",
		KillSwitch:      true,
		BlockFileshare:  true,
		Allowlist:       allowlist,
		MeshnetInfo:     meshInfo,
	}

	tests := []struct {
		name     string
		options  []Option
		expected Config
	}{
		{
			name:     "tunnel interface changes",
			options:  []Option{WithTunnelInterface("test")},
			expected: Config{TunnelInterface: "test"},
		},
		{
			name:     "KillSwitch interface changes",
			options:  []Option{WithKillSwitch(true)},
			expected: Config{KillSwitch: true},
		},
		{
			name:     "BlockFileshare interface changes",
			options:  []Option{WithBlockFileshare(true)},
			expected: Config{BlockFileshare: true},
		},
		{
			name:     "Allowlist interface changes",
			options:  []Option{WithAllowlist(allowlist)},
			expected: Config{Allowlist: allowlist},
		},
		{
			name:     "MeshnetInfo interface changes",
			options:  []Option{WithMeshnetInfo(meshInfo)},
			expected: Config{MeshnetInfo: meshInfo},
		},
		{
			name: "combine all members works",
			options: []Option{
				WithTunnelInterface("test"),
				WithKillSwitch(true),
				WithBlockFileshare(true),
				WithAllowlist(allowlist),
				WithMeshnetInfo(meshInfo),
			},
			expected: fullConfig,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewConfig(test.options...)
			assert.Equal(t, test.expected, c)
		})
	}
}

func TestMeshInfoIsSimilar(t *testing.T) {
	category.Set(t, category.Unit)

	randomIP := func() netip.Addr {
		return netip.AddrFrom4([4]byte{
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
		})
	}

	createPeer := func() mesh.MachinePeer {
		return mesh.MachinePeer{
			ID:                   uuid.New(),
			Address:              randomIP(),
			DoIAllowInbound:      true,
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: true,
			DoIAllowFileshare:    true,
			AlwaysAcceptFiles:    true,
		}
	}

	peer1 := createPeer()
	peer2 := createPeer()
	baseMeshInfo := NewMeshInfo(mesh.MachineMap{
		Peers: mesh.MachinePeers{peer1, peer2}}, "nordlynx")

	tests := []struct {
		name     string
		fn       func(*MeshInfo)
		expected bool
	}{
		{
			name:     "are equal for same data",
			fn:       func(m *MeshInfo) {},
			expected: true,
		},
		{
			name: "not equal when mesh interface changes",
			fn: func(m *MeshInfo) {
				m.MeshInterface = "other"
			},
			expected: false,
		},
		{
			name: "ID changes",
			fn: func(m *MeshInfo) {
				m.MeshnetMap.Peers[0].ID = uuid.New()
			},
			expected: false,
		},
		{
			name: "address changes",
			fn: func(m *MeshInfo) {
				m.MeshnetMap.Peers[0].Address = randomIP()
			},
			expected: false,
		},
		{
			name: "DoIAllowInbound changes",
			fn: func(m *MeshInfo) {
				m.MeshnetMap.Peers[0].DoIAllowInbound = false
			},
			expected: false,
		},
		{
			name: "DoIAllowRouting changes",
			fn: func(m *MeshInfo) {
				m.MeshnetMap.Peers[0].DoIAllowRouting = false
			},
			expected: false,
		},
		{
			name: "DoIAllowLocalNetwork changes",
			fn: func(m *MeshInfo) {
				m.MeshnetMap.Peers[0].DoIAllowLocalNetwork = false
			},
			expected: false,
		},
		{
			name: "DoIAllowFileshare changes",
			fn: func(m *MeshInfo) {
				m.MeshnetMap.Peers[0].DoIAllowFileshare = false
			},
			expected: false,
		},
		{
			name: "irrelevant changes have no effect",
			fn: func(m *MeshInfo) {
				peer := m.MeshnetMap.Peers[0]
				peer.Hostname = "host1"
				peer.OS = mesh.OperatingSystem{Name: "name"}
				peer.PublicKey = "key"
				peer.Endpoints = []netip.AddrPort{netip.AddrPortFrom(randomIP(), 1234)}
				peer.Email = "email"
				peer.IsLocal = true
				peer.DoesPeerAllowRouting = true
				peer.DoesPeerAllowInbound = true
				peer.DoesPeerAllowLocalNetwork = true
				peer.DoesPeerAllowFileshare = true
				peer.DoesPeerSupportRouting = true
				peer.AlwaysAcceptFiles = true
				peer.Nickname = "nick"
				m.MeshnetMap.Peers[0] = peer
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := NewMeshInfo(mesh.MachineMap{
				Peers: mesh.MachinePeers{peer1, peer2}}, "nordlynx")
			test.fn(m)
			assert.Equal(t, test.expected, baseMeshInfo.IsSimilar(m))
		})
	}
}
