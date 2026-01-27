package allowlist

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Event identification constants
const (
	Namespace         = "nordvpn-linux"
	Subscope          = "allowlist"
	EventOperation    = Subscope + "_operation"
	EventSnapshot     = Subscope + "_snapshot"
	contextPathPrefix = "allowlist"
)

// Operation types
const (
	OpAdd    = "add"
	OpRemove = "remove"
	OpClear  = "clear"
)

// Entry types
const (
	EntryPort      = "port"
	EntryPortRange = "port_range"
	EntrySubnet    = "subnet"
)

// Protocols
const (
	ProtoTCP  = "tcp"
	ProtoUDP  = "udp"
	ProtoBoth = "both"
	ProtoNA   = "n/a"
)

// Results
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
)

// globalContextPaths defines the common context paths included in all allowlist events.
var globalContextPaths = []string{
	// Device context
	"device.*",
	"application.nordvpnapp.version",
	"application.nordvpnapp.platform",
	// Related feature states
	"application.nordvpnapp.config.user_preferences.local_network_discovery_allowed.value",
	"application.nordvpnapp.config.user_preferences.kill_switch_enabled.value",
	"application.nordvpnapp.config.user_preferences.meshnet_enabled.value",
	"application.nordvpnapp.config.current_state.is_on_vpn.value",
}

// OperationEvent represents an allowlist add/remove/clear operation.
type OperationEvent struct {
	Namespace  string `json:"namespace"`
	Subscope   string `json:"subscope"`
	Event      string `json:"event"`
	Operation  string `json:"operation"`
	EntryType  string `json:"entry_type"`
	Protocol   string `json:"protocol"`
	Result     string `json:"result"`
	Error      string `json:"error"`
	Port       int64  `json:"port"`
	PortStart  int64  `json:"port_start"`
	PortEnd    int64  `json:"port_end"`
	Subnet     string `json:"subnet"`
	SubnetMask int64  `json:"subnet_mask"`
}

// SnapshotEvent represents the current allowlist configuration state.
type SnapshotEvent struct {
	Namespace    string   `json:"namespace"`
	Subscope     string   `json:"subscope"`
	Event        string   `json:"event"`
	TCPPorts     []int64  `json:"tcp_ports"`
	UDPPorts     []int64  `json:"udp_ports"`
	Subnets      []string `json:"subnets"`
	TCPPortCount int64    `json:"tcp_port_count"`
	UDPPortCount int64    `json:"udp_port_count"`
	SubnetCount  int64    `json:"subnet_count"`
	TotalCount   int64    `json:"total_count"`
	IsEnabled    bool     `json:"is_enabled"`
}

// NewPortOperation creates an event for single port add/remove
func NewPortOperation(op string, port int64, protocol string, success bool, errCode int64) *OperationEvent {
	return &OperationEvent{
		Namespace: Namespace,
		Subscope:  Subscope,
		Event:     EventOperation,
		Operation: op,
		EntryType: EntryPort,
		Protocol:  protocol,
		Result:    boolToResult(success),
		Error:     NewError(errCode).Error(),
		Port:      port,
	}
}

// NewPortRangeOperation creates an event for port range add/remove
func NewPortRangeOperation(op string, start, end int64, protocol string, success bool, errCode int64) *OperationEvent {
	return &OperationEvent{
		Namespace: Namespace,
		Subscope:  Subscope,
		Event:     EventOperation,
		Operation: op,
		EntryType: EntryPortRange,
		Protocol:  protocol,
		Result:    boolToResult(success),
		Error:     NewError(errCode).Error(),
		PortStart: start,
		PortEnd:   end,
	}
}

// NewSubnetOperation creates an event for subnet add/remove.
func NewSubnetOperation(op string, subnet string, success bool, errCode int64) *OperationEvent {
	return &OperationEvent{
		Namespace:  Namespace,
		Subscope:   Subscope,
		Event:      EventOperation,
		Operation:  op,
		EntryType:  EntrySubnet,
		Protocol:   ProtoNA,
		Result:     boolToResult(success),
		Error:      NewError(errCode).Error(),
		Subnet:     subnet,
		SubnetMask: extractMask(subnet),
	}
}

// NewClearOperation creates an event for clearing all entries
func NewClearOperation(success bool, errCode int64) *OperationEvent {
	return &OperationEvent{
		Namespace: Namespace,
		Subscope:  Subscope,
		Event:     EventOperation,
		Operation: OpClear,
		EntryType: EntryPort,
		Protocol:  ProtoBoth,
		Result:    boolToResult(success),
		Error:     NewError(errCode).Error(),
	}
}

// SnapshotConfig holds the configuration state for snapshot events
type SnapshotConfig struct {
	TCPPorts []int64
	UDPPorts []int64
	Subnets  []string
}

// NewSnapshot creates a snapshot event from current config.
func NewSnapshot(cfg SnapshotConfig) *SnapshotEvent {
	tcpCount := int64(len(cfg.TCPPorts))
	udpCount := int64(len(cfg.UDPPorts))
	subnetCount := int64(len(cfg.Subnets))
	total := tcpCount + udpCount + subnetCount
	return &SnapshotEvent{
		Namespace:    Namespace,
		Subscope:     Subscope,
		Event:        EventSnapshot,
		TCPPorts:     cfg.TCPPorts,
		UDPPorts:     cfg.UDPPorts,
		Subnets:      cfg.Subnets,
		TCPPortCount: tcpCount,
		UDPPortCount: udpCount,
		SubnetCount:  subnetCount,
		TotalCount:   total,
		IsEnabled:    total > 0,
	}
}

// ToDebuggerEvent converts OperationEvent to a DebuggerEvent for moose publishing
func (e *OperationEvent) ToDebuggerEvent() *events.DebuggerEvent {
	jsonData, _ := json.Marshal(e)
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: contextPathPrefix + ".namespace", Value: e.Namespace},
			events.ContextValue{Path: contextPathPrefix + ".subscope", Value: e.Subscope},
			events.ContextValue{Path: contextPathPrefix + ".event", Value: e.Event},
			events.ContextValue{Path: contextPathPrefix + ".operation", Value: e.Operation},
			events.ContextValue{Path: contextPathPrefix + ".entry_type", Value: e.EntryType},
			events.ContextValue{Path: contextPathPrefix + ".protocol", Value: e.Protocol},
			events.ContextValue{Path: contextPathPrefix + ".result", Value: e.Result},
			events.ContextValue{Path: contextPathPrefix + ".error", Value: e.Error},
			events.ContextValue{Path: contextPathPrefix + ".port", Value: e.Port},
			events.ContextValue{Path: contextPathPrefix + ".port_start", Value: e.PortStart},
			events.ContextValue{Path: contextPathPrefix + ".port_end", Value: e.PortEnd},
			events.ContextValue{Path: contextPathPrefix + ".subnet", Value: e.Subnet},
			events.ContextValue{Path: contextPathPrefix + ".subnet_mask", Value: e.SubnetMask},
		).
		WithGlobalContextPaths(globalContextPaths...)
}

// ToDebuggerEvent converts SnapshotEvent to a DebuggerEvent for moose publishing
func (e *SnapshotEvent) ToDebuggerEvent() *events.DebuggerEvent {
	jsonData, _ := json.Marshal(e)
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: contextPathPrefix + ".namespace", Value: e.Namespace},
			events.ContextValue{Path: contextPathPrefix + ".subscope", Value: e.Subscope},
			events.ContextValue{Path: contextPathPrefix + ".event", Value: e.Event},
			events.ContextValue{Path: contextPathPrefix + ".tcp_ports", Value: e.TCPPorts},
			events.ContextValue{Path: contextPathPrefix + ".udp_ports", Value: e.UDPPorts},
			events.ContextValue{Path: contextPathPrefix + ".subnets", Value: e.Subnets},
			events.ContextValue{Path: contextPathPrefix + ".tcp_port_count", Value: e.TCPPortCount},
			events.ContextValue{Path: contextPathPrefix + ".udp_port_count", Value: e.UDPPortCount},
			events.ContextValue{Path: contextPathPrefix + ".subnet_count", Value: e.SubnetCount},
			events.ContextValue{Path: contextPathPrefix + ".total_count", Value: e.TotalCount},
			events.ContextValue{Path: contextPathPrefix + ".is_enabled", Value: e.IsEnabled},
		).
		WithGlobalContextPaths(globalContextPaths...)
}

func boolToResult(success bool) string {
	if success {
		return ResultSuccess
	}
	return ResultFailure
}

func extractMask(cidr string) int64 {
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return 0
	}
	mask, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0
	}
	return mask
}

type Error struct {
	Code int64
}

func NewError(code int64) *Error {
	return &Error{Code: code}
}

func (e *Error) Error() string {
	switch e.Code {
	case internal.CodeSuccess:
		return "success"
	case internal.CodeFailure:
		return "operation failed"
	case internal.CodeConfigError:
		return "configuration error"
	case internal.CodePrivateSubnetLANDiscovery:
		return "private subnet conflicts with LAN discovery"
	case internal.CodeAllowlistInvalidSubnet:
		return "invalid subnet format"
	case internal.CodeAllowlistSubnetNoop:
		return "subnet already exists or does not exist"
	case internal.CodeAllowlistPortOutOfRange:
		return "port out of valid range (1-65535)"
	case internal.CodeAllowlistPortNoop:
		return "port already exists or does not exist"
	default:
		return fmt.Sprintf("unknown allowlist error (code %d)", e.Code)
	}
}
