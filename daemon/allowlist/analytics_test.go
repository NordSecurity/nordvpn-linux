package allowlist

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/analytics"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				Result:    analytics.ResultSuccess,
				Error:     "",
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
				Result:    analytics.ResultFailure,
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

func TestNewOperationEventFromRequest(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		req       *pb.SetAllowlistRequest
		op        string
		success   bool
		errCode   int64
		wantEvent *OperationEvent
	}{
		{
			name: "subnet request add success",
			req: &pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistSubnetRequest{
					SetAllowlistSubnetRequest: &pb.SetAllowlistSubnetRequest{
						Subnet: "192.168.1.0/24",
					},
				},
			},
			op:      OpAdd,
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace:       Namespace,
				Subscope:        Subscope,
				Event:           EventOperation,
				Operation:       OpAdd,
				EntryType:       EntrySubnet,
				Protocol:        ProtoNA,
				Result:          analytics.ResultSuccess,
				Error:           "",
				SubnetMask:      24,
				IsPrivateSubnet: true,
			},
		},
		{
			name: "subnet request remove failure",
			req: &pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistSubnetRequest{
					SetAllowlistSubnetRequest: &pb.SetAllowlistSubnetRequest{
						Subnet: "10.0.0.0/8",
					},
				},
			},
			op:      OpRemove,
			success: false,
			errCode: internal.CodeAllowlistSubnetNoop,
			wantEvent: &OperationEvent{
				Namespace:       Namespace,
				Subscope:        Subscope,
				Event:           EventOperation,
				Operation:       OpRemove,
				EntryType:       EntrySubnet,
				Protocol:        ProtoNA,
				Result:          analytics.ResultFailure,
				Error:           "subnet unchanged: already in desired state",
				SubnetMask:      8,
				IsPrivateSubnet: true,
			},
		},
		{
			name: "single port TCP only",
			req: &pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
					SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
						IsTcp: true,
						IsUdp: false,
						PortRange: &pb.PortRange{
							StartPort: 22,
							EndPort:   0,
						},
					},
				},
			},
			op:      OpAdd,
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpAdd,
				EntryType: EntryPort,
				Protocol:  ProtoTCP,
				Result:    analytics.ResultSuccess,
				Error:     "",
				Port:      22,
			},
		},
		{
			name: "single port UDP only",
			req: &pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
					SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
						IsTcp: false,
						IsUdp: true,
						PortRange: &pb.PortRange{
							StartPort: 53,
							EndPort:   53,
						},
					},
				},
			},
			op:      OpAdd,
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpAdd,
				EntryType: EntryPort,
				Protocol:  ProtoUDP,
				Result:    analytics.ResultSuccess,
				Error:     "",
				Port:      53,
			},
		},
		{
			name: "single port both protocols",
			req: &pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
					SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
						IsTcp: true,
						IsUdp: true,
						PortRange: &pb.PortRange{
							StartPort: 443,
							EndPort:   0,
						},
					},
				},
			},
			op:      OpAdd,
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace: Namespace,
				Subscope:  Subscope,
				Event:     EventOperation,
				Operation: OpAdd,
				EntryType: EntryPort,
				Protocol:  ProtoBoth,
				Result:    analytics.ResultSuccess,
				Error:     "",
				Port:      443,
			},
		},
		{
			name: "port range TCP",
			req: &pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
					SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
						IsTcp: true,
						IsUdp: false,
						PortRange: &pb.PortRange{
							StartPort: 3000,
							EndPort:   8000,
						},
					},
				},
			},
			op:      OpAdd,
			success: true,
			errCode: internal.CodeSuccess,
			wantEvent: &OperationEvent{
				Namespace:      Namespace,
				Subscope:       Subscope,
				Event:          EventOperation,
				Operation:      OpAdd,
				EntryType:      EntryPortRange,
				Protocol:       ProtoTCP,
				Result:         analytics.ResultSuccess,
				Error:          "",
				PortRangeStart: 3000,
				PortRangeEnd:   8000,
			},
		},
		{
			name: "port range failure",
			req: &pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
					SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
						IsTcp: true,
						IsUdp: true,
						PortRange: &pb.PortRange{
							StartPort: 70000,
							EndPort:   80000,
						},
					},
				},
			},
			op:      OpAdd,
			success: false,
			errCode: internal.CodeAllowlistPortOutOfRange,
			wantEvent: &OperationEvent{
				Namespace:      Namespace,
				Subscope:       Subscope,
				Event:          EventOperation,
				Operation:      OpAdd,
				EntryType:      EntryPortRange,
				Protocol:       ProtoBoth,
				Result:         analytics.ResultFailure,
				Error:          "port out of valid range (1-65535)",
				PortRangeStart: 70000,
				PortRangeEnd:   80000,
			},
		},
		{
			name:      "nil request returns nil",
			req:       &pb.SetAllowlistRequest{},
			op:        OpAdd,
			success:   true,
			errCode:   internal.CodeSuccess,
			wantEvent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOperationEventFromRequest(tt.req, tt.op, tt.success, tt.errCode)
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
				Namespace:          Namespace,
				Subscope:           Subscope,
				Event:              EventSnapshot,
				TCPPorts:           []int64{},
				UDPPorts:           []int64{},
				TCPPortCount:       0,
				UDPPortCount:       0,
				SubnetCount:        0,
				PrivateSubnetCount: 0,
				PublicSubnetCount:  0,
				TotalEntryCount:    0,
				IsEnabled:          false,
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
				Namespace:          Namespace,
				Subscope:           Subscope,
				Event:              EventSnapshot,
				TCPPorts:           []int64{22, 80, 443},
				UDPPorts:           []int64{},
				TCPPortCount:       3,
				UDPPortCount:       0,
				SubnetCount:        0,
				PrivateSubnetCount: 0,
				PublicSubnetCount:  0,
				TotalEntryCount:    3,
				IsEnabled:          true,
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
				Namespace:          Namespace,
				Subscope:           Subscope,
				Event:              EventSnapshot,
				TCPPorts:           []int64{22, 80, 443},
				UDPPorts:           []int64{53, 123},
				TCPPortCount:       3,
				UDPPortCount:       2,
				SubnetCount:        1,
				PrivateSubnetCount: 1,
				PublicSubnetCount:  0,
				TotalEntryCount:    6,
				IsEnabled:          true,
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
				Namespace:          Namespace,
				Subscope:           Subscope,
				Event:              EventSnapshot,
				TCPPorts:           []int64{22},
				UDPPorts:           []int64{},
				TCPPortCount:       1,
				UDPPortCount:       0,
				SubnetCount:        3,
				PrivateSubnetCount: 3,
				PublicSubnetCount:  0,
				TotalEntryCount:    4,
				IsEnabled:          true,
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
				Namespace:          Namespace,
				Subscope:           Subscope,
				Event:              EventSnapshot,
				TCPPorts:           []int64{},
				UDPPorts:           []int64{},
				TCPPortCount:       0,
				UDPPortCount:       0,
				SubnetCount:        2,
				PrivateSubnetCount: 2,
				PublicSubnetCount:  0,
				TotalEntryCount:    2,
				IsEnabled:          true,
			},
		},
		{
			name: "mixed private and public subnets",
			cfg: SnapshotConfig{
				TCPPorts: []int64{},
				UDPPorts: []int64{},
				Subnets:  []string{"192.168.1.0/24", "8.8.8.0/24", "10.0.0.0/8", "1.1.1.0/24"},
			},
			wantEvent: &SnapshotEvent{
				Namespace:          Namespace,
				Subscope:           Subscope,
				Event:              EventSnapshot,
				TCPPorts:           []int64{},
				UDPPorts:           []int64{},
				TCPPortCount:       0,
				UDPPortCount:       0,
				SubnetCount:        4,
				PrivateSubnetCount: 2,
				PublicSubnetCount:  2,
				TotalEntryCount:    4,
				IsEnabled:          true,
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
	req := &pb.SetAllowlistRequest{
		Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
			SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
				IsTcp:     true,
				IsUdp:     false,
				PortRange: &pb.PortRange{StartPort: 22, EndPort: 0},
			},
		},
	}
	event := NewOperationEventFromRequest(req, OpAdd, true, 0)
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

func TestParseSubnetInfo(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		cidr          string
		wantMask      int
		wantIsPrivate bool
	}{
		// Valid private subnets
		{"192.168.1.0/24", 24, true},
		{"10.0.0.0/8", 8, true},
		{"172.16.0.0/16", 16, true},
		{"192.168.1.100/32", 32, true},
		// Valid public subnets
		{"8.8.8.0/24", 24, false},
		{"1.1.1.0/24", 24, false},
		// Edge cases
		{"0.0.0.0/0", 0, false}, // 0.0.0.0 is not considered private
		// Invalid inputs
		{"invalid", -1, false},
		{"192.168.1.0", -1, false},
		{"192.168.1.0/", -1, false},
		{"192.168.1.0/abc", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			gotMask, gotIsPrivate := parseSubnetInfo(tt.cidr)
			assert.Equal(t, tt.wantMask, gotMask)
			assert.Equal(t, tt.wantIsPrivate, gotIsPrivate)
		})
	}
}

func TestBoolToResult(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Equal(t, analytics.ResultSuccess, analytics.BoolToResult(true))
	assert.Equal(t, analytics.ResultFailure, analytics.BoolToResult(false))
}

func TestError(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		code    int64
		wantMsg string
	}{
		{internal.CodeSuccess, ""},
		{internal.CodeFailure, "operation failed"},
		{internal.CodeConfigError, "configuration error"},
		{internal.CodePrivateSubnetLANDiscovery, "private subnet conflicts with LAN discovery"},
		{internal.CodeAllowlistInvalidSubnet, "invalid subnet format"},
		{internal.CodeAllowlistSubnetNoop, "subnet unchanged: already in desired state"},
		{internal.CodeAllowlistPortOutOfRange, "port out of valid range (1-65535)"},
		{internal.CodeAllowlistPortNoop, "port unchanged: already in desired state"},
		{9999, "unknown allowlist error (code 9999)"},
	}

	for _, tt := range tests {
		t.Run(tt.wantMsg, func(t *testing.T) {
			got := codeToString(tt.code)
			assert.Equal(t, tt.wantMsg, got)
		})
	}
}
