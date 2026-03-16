package firewall

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal/analytics"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockevents "github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeterminePurposes(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		config   Config
		expected []string
	}{
		{
			name:     "empty config returns empty list",
			config:   Config{},
			expected: nil,
		},
		{
			name: "tunnel interface returns vpn",
			config: Config{
				TunnelInterface: "nordlynx",
			},
			expected: []string{PurposeVPN},
		},
		{
			name: "killswitch alone returns killswitch",
			config: Config{
				KillSwitch: true,
			},
			expected: []string{PurposeKillSwitch},
		},
		{
			name: "meshnet only returns meshnet",
			config: Config{
				MeshnetInfo: &MeshInfo{
					MeshnetMap:    mesh.MachineMap{},
					MeshInterface: "nordlynx",
				},
			},
			expected: []string{PurposeMeshnet},
		},
		{
			name: "vpn and meshnet returns both in list",
			config: Config{
				TunnelInterface: "nordlynx",
				MeshnetInfo: &MeshInfo{
					MeshnetMap:    mesh.MachineMap{},
					MeshInterface: "nordlynx",
				},
			},
			expected: []string{PurposeVPN, PurposeMeshnet},
		},
		{
			name: "killswitch with meshnet returns both",
			config: Config{
				KillSwitch: true,
				MeshnetInfo: &MeshInfo{
					MeshnetMap:    mesh.MachineMap{},
					MeshInterface: "nordlynx",
				},
			},
			expected: []string{PurposeMeshnet, PurposeKillSwitch},
		},
		{
			name: "allowlist with subnets returns allowlist",
			config: Config{
				Allowlist: config.Allowlist{
					Subnets: []string{"192.168.1.0/24"},
				},
			},
			expected: []string{PurposeAllowlist},
		},
		{
			name: "allowlist with TCP ports returns allowlist",
			config: Config{
				Allowlist: config.Allowlist{
					Ports: config.Ports{
						TCP: map[int64]bool{443: true},
					},
				},
			},
			expected: []string{PurposeAllowlist},
		},
		{
			name: "allowlist with UDP ports returns allowlist",
			config: Config{
				Allowlist: config.Allowlist{
					Ports: config.Ports{
						UDP: map[int64]bool{53: true},
					},
				},
			},
			expected: []string{PurposeAllowlist},
		},
		{
			name: "vpn with killswitch and allowlist returns all three",
			config: Config{
				TunnelInterface: "nordlynx",
				KillSwitch:      true,
				Allowlist: config.Allowlist{
					Subnets: []string{"10.0.0.0/8"},
				},
			},
			expected: []string{PurposeVPN, PurposeKillSwitch, PurposeAllowlist},
		},
		{
			name: "all four purposes combined",
			config: Config{
				TunnelInterface: "nordlynx",
				MeshnetInfo: &MeshInfo{
					MeshnetMap:    mesh.MachineMap{},
					MeshInterface: "nordlynx",
				},
				KillSwitch: true,
				Allowlist: config.Allowlist{
					Subnets: []string{"192.168.0.0/16"},
				},
			},
			expected: []string{PurposeVPN, PurposeMeshnet, PurposeKillSwitch, PurposeAllowlist},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determinePurposes(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigureEvent_ToDebuggerEvent(t *testing.T) {
	category.Set(t, category.Unit)

	event := newConfigureEvent(Config{TunnelInterface: "nordlynx"}, true, nil)
	debuggerEvent := event.ToDebuggerEvent()

	require.NotNil(t, debuggerEvent)
	assert.NotEmpty(t, debuggerEvent.JsonData)

	// Verify JSON structure
	var decoded ConfigureEvent
	err := json.Unmarshal([]byte(debuggerEvent.JsonData), &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.Namespace, decoded.Namespace)
	assert.Equal(t, event.Subscope, decoded.Subscope)
	assert.Equal(t, event.Event, decoded.Event)
	assert.Equal(t, event.Status, decoded.Status)
	assert.Equal(t, event.Purpose, decoded.Purpose)

	// Verify context paths are set
	assert.NotEmpty(t, debuggerEvent.KeyBasedContextPaths)
	assert.NotEmpty(t, debuggerEvent.GeneralContextPaths)
}

// MockBackend is a test implementation of FirewallBackend
type MockBackend struct {
	ConfigureErr error
	FlushErr     error
}

func (m *MockBackend) Configure(config Config) error {
	return m.ConfigureErr
}

func (m *MockBackend) Flush() error {
	return m.FlushErr
}

func TestFirewall_Configure_EmitsSuccessEvent(t *testing.T) {
	category.Set(t, category.Unit)

	publisher := &mockevents.MockPublisher[events.DebuggerEvent]{}
	backend := &MockBackend{}
	fw := NewFirewall(backend, true, publisher)

	cfg := Config{TunnelInterface: "nordlynx"}
	err := fw.Configure(cfg)

	require.NoError(t, err)

	event, _, ok := publisher.PopEvent()
	require.True(t, ok)

	var decoded ConfigureEvent
	jsonErr := json.Unmarshal([]byte(event.JsonData), &decoded)
	require.NoError(t, jsonErr)
	assert.Equal(t, analytics.ResultSuccess, decoded.Status)
	assert.Equal(t, []string{PurposeVPN}, decoded.Purpose)
}

func TestFirewall_Configure_EmitsFailureEvent(t *testing.T) {
	category.Set(t, category.Unit)

	publisher := &mockevents.MockPublisher[events.DebuggerEvent]{}
	backend := &MockBackend{ConfigureErr: errors.New("nftables error")}
	fw := NewFirewall(backend, true, publisher)

	cfg := Config{TunnelInterface: "nordlynx"}
	err := fw.Configure(cfg)

	require.Error(t, err)

	event, _, ok := publisher.PopEvent()
	require.True(t, ok)

	var decoded ConfigureEvent
	jsonErr := json.Unmarshal([]byte(event.JsonData), &decoded)
	require.NoError(t, jsonErr)
	assert.Equal(t, analytics.ResultFailure, decoded.Status)
	assert.Equal(t, "nftables error", decoded.Error)
}

func TestFirewall_Configure_NoEventWhenDisabled(t *testing.T) {
	category.Set(t, category.Unit)

	publisher := &mockevents.MockPublisher[events.DebuggerEvent]{}
	backend := &MockBackend{}
	fw := NewFirewall(backend, false, publisher)

	cfg := Config{TunnelInterface: "nordlynx"}
	err := fw.Configure(cfg)

	require.NoError(t, err)
	_, _, ok := publisher.PopEvent()
	assert.False(t, ok)
}

func TestFirewall_Configure_NilPublisherNoPanic(t *testing.T) {
	category.Set(t, category.Unit)

	backend := &MockBackend{}
	fw := NewFirewall(backend, true, nil)

	cfg := Config{TunnelInterface: "nordlynx"}
	err := fw.Configure(cfg)

	require.NoError(t, err)
}
