package allowlist

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPortOperation(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		op        string
		port      int64
		protocol  string
		success   bool
		errCode   int64
		wantEvent *OperationEvent
	}{
		{
			name:     "add TCP port success",
			op:       OpAdd,
			port:     22,
			protocol: ProtoTCP,
			success:  true,
			errCode:  internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpAdd,
				EntryType: EntryPort,
				Protocol:  ProtoTCP,
				Result:    ResultSuccess,
				Error:     "success",
				Port:      22,
			},
		},
		{
			name:     "add UDP port failure",
			op:       OpAdd,
			port:     70000,
			protocol: ProtoUDP,
			success:  false,
			errCode:  internal.CodeAllowlistPortOutOfRange,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpAdd,
				EntryType: EntryPort,
				Protocol:  ProtoUDP,
				Result:    ResultFailure,
				Error:     "port out of valid range (1-65535)",
				Port:      70000,
			},
		},
		{
			name:     "remove TCP port success",
			op:       OpRemove,
			port:     443,
			protocol: ProtoTCP,
			success:  true,
			errCode:  internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpRemove,
				EntryType: EntryPort,
				Protocol:  ProtoTCP,
				Result:    ResultSuccess,
				Error:     "success",
				Port:      443,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPortOperation(tt.op, tt.port, tt.protocol, tt.success, tt.errCode)
			assert.Equal(t, tt.wantEvent, got)
		})
	}
}

func TestNewPortRangeOperation(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		op        string
		start     int64
		end       int64
		protocol  string
		success   bool
		errCode   int64
		wantEvent *OperationEvent
	}{
		{
			name:     "add UDP port range success",
			op:       OpAdd,
			start:    3000,
			end:      8000,
			protocol: ProtoUDP,
			success:  true,
			errCode:  internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpAdd,
				EntryType: EntryPortRange,
				Protocol:  ProtoUDP,
				Result:    ResultSuccess,
				Error:     "success",
				PortStart: 3000,
				PortEnd:   8000,
			},
		},
		{
			name:     "remove TCP port range success",
			op:       OpRemove,
			start:    1024,
			end:      2048,
			protocol: ProtoTCP,
			success:  true,
			errCode:  internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpRemove,
				EntryType: EntryPortRange,
				Protocol:  ProtoTCP,
				Result:    ResultSuccess,
				Error:     "success",
				PortStart: 1024,
				PortEnd:   2048,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPortRangeOperation(tt.op, tt.start, tt.end, tt.protocol, tt.success, tt.errCode)
			assert.Equal(t, tt.wantEvent, got)
		})
	}
}

func TestNewSubnetOperation(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		op        string
		subnet    string
		success   bool
		errCode   int64
		wantEvent *OperationEvent
	}{
		{
			name:    "add subnet /24 success",
			op:      OpAdd,
			subnet:  "192.168.1.0/24",
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace:  Namespace,
				Subscope:   Subscope,
				Event:      EventOperation,
				Operation:  OpAdd,
				EntryType:  EntrySubnet,
				Protocol:   ProtoNA,
				Result:     ResultSuccess,
				Error:      "success",
				Subnet:     "192.168.1.0/24",
				SubnetMask: 24,
			},
		},
		{
			name:    "add subnet /16 success",
			op:      OpAdd,
			subnet:  "10.0.0.0/16",
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace:  Namespace,
				Subscope:   Subscope,
				Event:      EventOperation,
				Operation:  OpAdd,
				EntryType:  EntrySubnet,
				Protocol:   ProtoNA,
				Result:     ResultSuccess,
				Error:      "success",
				Subnet:     "10.0.0.0/16",
				SubnetMask: 16,
			},
		},
		{
			name:    "add single host /32 success",
			op:      OpAdd,
			subnet:  "192.168.1.100/32",
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace:  Namespace,
				Subscope:   Subscope,
				Event:      EventOperation,
				Operation:  OpAdd,
				EntryType:  EntrySubnet,
				Protocol:   ProtoNA,
				Result:     ResultSuccess,
				Error:      "success",
				Subnet:     "192.168.1.100/32",
				SubnetMask: 32,
			},
		},
		{
			name:    "add invalid subnet failure",
			op:      OpAdd,
			subnet:  "invalid",
			success: false,
			errCode: internal.CodeAllowlistInvalidSubnet,
			wantEvent: &OperationEvent{
				Namespace:  Namespace,
				Subscope:   Subscope,
				Event:      EventOperation,
				Operation:  OpAdd,
				EntryType:  EntrySubnet,
				Protocol:   ProtoNA,
				Result:     ResultFailure,
				Error:      "invalid subnet format",
				Subnet:     "invalid",
				SubnetMask: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSubnetOperation(tt.op, tt.subnet, tt.success, tt.errCode)
			assert.Equal(t, tt.wantEvent, got)
		})
	}
}

func TestNewClearOperation(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		success   bool
		errCode   int64
		wantEvent *OperationEvent
	}{
		{
			name:    "clear success",
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpClear,
				EntryType: EntryPort,
				Protocol:  ProtoBoth,
				Result:    ResultSuccess,
				Error:     "success",
			},
		},
		{
			name:    "clear failure",
			success: false,
			errCode: internal.CodeFailure,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpClear,
				EntryType: EntryPort,
				Protocol:  ProtoBoth,
				Result:    ResultFailure,
				Error:     "operation failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClearOperation(tt.success, tt.errCode)
			assert.Equal(t, tt.wantEvent, got)
		})
	}
}

func TestNewSnapshot(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		cfg       SnapshotConfig
		wantEvent *SnapshotEvent
	}{
		{
			name: "empty config",
			cfg: SnapshotConfig{
				TCPPorts: []int64{},
				UDPPorts: []int64{},
				Subnets:  []string{},
			},
			wantEvent: &SnapshotEvent{
				Namespace:    Namespace,
				Subscope:     Subscope,
				Event:        EventSnapshot,
				TCPPorts:     []int64{},
				UDPPorts:     []int64{},
				Subnets:      []string{},
				TCPPortCount: 0,
				UDPPortCount: 0,
				SubnetCount:  0,
				TotalCount:   0,
				IsEnabled:    false,
			},
		},
		{
			name: "only TCP ports",
			cfg: SnapshotConfig{
				TCPPorts: []int64{22, 80, 443},
				UDPPorts: []int64{},
				Subnets:  []string{},
			},
			wantEvent: &SnapshotEvent{
				Namespace:    Namespace,
				Subscope:     Subscope,
				Event:        EventSnapshot,
				TCPPorts:     []int64{22, 80, 443},
				UDPPorts:     []int64{},
				Subnets:      []string{},
				TCPPortCount: 3,
				UDPPortCount: 0,
				SubnetCount:  0,
				TotalCount:   3,
				IsEnabled:    true,
			},
		},
		{
			name: "mixed config",
			cfg: SnapshotConfig{
				TCPPorts: []int64{22, 80, 443},
				UDPPorts: []int64{53, 123},
				Subnets:  []string{"192.168.1.0/24"},
			},
			wantEvent: &SnapshotEvent{
				Namespace:    Namespace,
				Subscope:     Subscope,
				Event:        EventSnapshot,
				TCPPorts:     []int64{22, 80, 443},
				UDPPorts:     []int64{53, 123},
				Subnets:      []string{"192.168.1.0/24"},
				TCPPortCount: 3,
				UDPPortCount: 2,
				SubnetCount:  1,
				TotalCount:   6,
				IsEnabled:    true,
			},
		},
		{
			name: "multiple subnets",
			cfg: SnapshotConfig{
				TCPPorts: []int64{22},
				UDPPorts: []int64{},
				Subnets:  []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/16"},
			},
			wantEvent: &SnapshotEvent{
				Namespace:    Namespace,
				Subscope:     Subscope,
				Event:        EventSnapshot,
				TCPPorts:     []int64{22},
				UDPPorts:     []int64{},
				Subnets:      []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/16"},
				TCPPortCount: 1,
				UDPPortCount: 0,
				SubnetCount:  3,
				TotalCount:   4,
				IsEnabled:    true,
			},
		},
		{
			name: "only subnets",
			cfg: SnapshotConfig{
				TCPPorts: []int64{},
				UDPPorts: []int64{},
				Subnets:  []string{"192.168.0.0/16", "10.10.10.0/24"},
			},
			wantEvent: &SnapshotEvent{
				Namespace:    Namespace,
				Subscope:     Subscope,
				Event:        EventSnapshot,
				TCPPorts:     []int64{},
				UDPPorts:     []int64{},
				Subnets:      []string{"192.168.0.0/16", "10.10.10.0/24"},
				TCPPortCount: 0,
				UDPPortCount: 0,
				SubnetCount:  2,
				TotalCount:   2,
				IsEnabled:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSnapshot(tt.cfg)
			assert.Equal(t, tt.wantEvent, got)
		})
	}
}

func TestOperationEvent_ToDebuggerEvent(t *testing.T) {
	category.Set(t, category.Unit)
	event := NewPortOperation(OpAdd, 22, ProtoTCP, true, 0)
	debuggerEvent := event.ToDebuggerEvent()

	require.NotNil(t, debuggerEvent)
	assert.NotEmpty(t, debuggerEvent.JsonData)
	assert.NotEmpty(t, debuggerEvent.KeyBasedContextPaths)
	assert.NotEmpty(t, debuggerEvent.GeneralContextPaths)

	var parsed OperationEvent
	err := json.Unmarshal([]byte(debuggerEvent.JsonData), &parsed)
	require.NoError(t, err)
	assert.Equal(t, event, &parsed)

	foundEvent := false
	foundOperation := false
	for _, ctx := range debuggerEvent.KeyBasedContextPaths {
		if ctx.Path == contextPathPrefix+".event" {
			foundEvent = true
			assert.Equal(t, EventOperation, ctx.Value)
		}
		if ctx.Path == contextPathPrefix+".operation" {
			foundOperation = true
			assert.Equal(t, OpAdd, ctx.Value)
		}
	}
	assert.True(t, foundEvent, "expected %s.event context path", contextPathPrefix)
	assert.True(t, foundOperation, "expected %s.operation context path", contextPathPrefix)

	assert.Contains(t, debuggerEvent.GeneralContextPaths, "device.*")
	assert.Contains(t, debuggerEvent.GeneralContextPaths, "application.nordvpnapp.version")
}

func TestSnapshotEvent_ToDebuggerEvent(t *testing.T) {
	category.Set(t, category.Unit)
	event := NewSnapshot(SnapshotConfig{
		TCPPorts: []int64{22, 80, 443},
		UDPPorts: []int64{53, 123},
		Subnets:  []string{"192.168.1.0/24"},
	})
	debuggerEvent := event.ToDebuggerEvent()

	require.NotNil(t, debuggerEvent)
	assert.NotEmpty(t, debuggerEvent.JsonData)
	assert.NotEmpty(t, debuggerEvent.KeyBasedContextPaths)
	assert.NotEmpty(t, debuggerEvent.GeneralContextPaths)

	var parsed SnapshotEvent
	err := json.Unmarshal([]byte(debuggerEvent.JsonData), &parsed)
	require.NoError(t, err)
	assert.Equal(t, event, &parsed)

	foundEvent := false
	foundIsEnabled := false
	for _, ctx := range debuggerEvent.KeyBasedContextPaths {
		if ctx.Path == contextPathPrefix+".event" {
			foundEvent = true
			assert.Equal(t, EventSnapshot, ctx.Value)
		}
		if ctx.Path == contextPathPrefix+".is_enabled" {
			foundIsEnabled = true
			assert.Equal(t, true, ctx.Value)
		}
	}
	assert.True(t, foundEvent, "expected %s.event context path", contextPathPrefix)
	assert.True(t, foundIsEnabled, "expected %s.is_enabled context path", contextPathPrefix)

	assert.Contains(t, debuggerEvent.GeneralContextPaths,
		"application.nordvpnapp.config.user_preferences.local_network_discovery_allowed.value")
	assert.Contains(t, debuggerEvent.GeneralContextPaths,
		"application.nordvpnapp.config.user_preferences.kill_switch_enabled.value")
	assert.Contains(t, debuggerEvent.GeneralContextPaths,
		"application.nordvpnapp.config.user_preferences.meshnet_enabled.value")
	assert.Contains(t, debuggerEvent.GeneralContextPaths,
		"application.nordvpnapp.config.current_state.is_on_vpn.value")
}

func TestExtractMask(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		cidr     string
		wantMask int64
	}{
		{"192.168.1.0/24", 24},
		{"10.0.0.0/8", 8},
		{"172.16.0.0/16", 16},
		{"192.168.1.100/32", 32},
		{"invalid", 0},
		{"192.168.1.0", 0},
		{"192.168.1.0/", 0},
		{"192.168.1.0/abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			got := extractMask(tt.cidr)
			assert.Equal(t, tt.wantMask, got)
		})
	}
}

func TestBoolToResult(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Equal(t, ResultSuccess, boolToResult(true))
	assert.Equal(t, ResultFailure, boolToResult(false))
}

func TestError(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		code    int64
		wantMsg string
	}{
		{internal.CodeSuccess, "success"},
		{internal.CodeFailure, "operation failed"},
		{internal.CodeConfigError, "configuration error"},
		{internal.CodePrivateSubnetLANDiscovery, "private subnet conflicts with LAN discovery"},
		{internal.CodeAllowlistInvalidSubnet, "invalid subnet format"},
		{internal.CodeAllowlistSubnetNoop, "subnet already exists or does not exist"},
		{internal.CodeAllowlistPortOutOfRange, "port out of valid range (1-65535)"},
		{internal.CodeAllowlistPortNoop, "port already exists or does not exist"},
		{9999, "unknown allowlist error (code 9999)"},
	}

	for _, tt := range tests {
		t.Run(tt.wantMsg, func(t *testing.T) {
			err := NewError(tt.code)
			assert.Equal(t, tt.wantMsg, err.Error())
			assert.Equal(t, tt.code, err.Code)
		})
	}
}
