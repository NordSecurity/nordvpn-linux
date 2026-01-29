package allowlist

import (
	"encoding/json"
	"fmt"
	"log"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/analytics"
)

// Event identification constants
const (
	Namespace         = internal.DebugEventMessageNamespace
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
	Namespace       string `json:"namespace"`
	Subscope        string `json:"subscope"`
	Event           string `json:"event"`
	Operation       string `json:"operation"`
	EntryType       string `json:"entry_type"`
	Protocol        string `json:"protocol"`
	Result          string `json:"result"`
	Error           string `json:"error,omitempty"`
	Port            int64  `json:"port,omitempty"`
	PortRangeStart  int64  `json:"port_range_start,omitempty"`
	PortRangeEnd    int64  `json:"port_range_end,omitempty"`
	SubnetMask      int    `json:"subnet_mask,omitempty"`
	IsPrivateSubnet bool   `json:"is_private_subnet,omitempty"`
}

// SnapshotEvent represents the current allowlist configuration state.
type SnapshotEvent struct {
	Namespace          string  `json:"namespace"`
	Subscope           string  `json:"subscope"`
	Event              string  `json:"event"`
	TCPPorts           []int64 `json:"tcp_ports"`
	UDPPorts           []int64 `json:"udp_ports"`
	TCPPortCount       int64   `json:"tcp_port_count"`
	UDPPortCount       int64   `json:"udp_port_count"`
	SubnetCount        int64   `json:"subnet_count"`
	PrivateSubnetCount int64   `json:"private_subnet_count"`
	PublicSubnetCount  int64   `json:"public_subnet_count"`
	TotalEntryCount    int64   `json:"total_entry_count"`
	IsEnabled          bool    `json:"is_enabled"`
}

// newPortOperation creates an event for single port add/remove
func newPortOperation(op string, port int64, protocol string, success bool, errCode int64) *OperationEvent {
	return &OperationEvent{
		Namespace: Namespace,
		Subscope:  Subscope,
		Event:     EventOperation,
		Operation: op,
		EntryType: EntryPort,
		Protocol:  protocol,
		Result:    analytics.BoolToResult(success),
		Error:     codeToString(errCode),
		Port:      port,
	}
}

// newPortRangeOperation creates an event for port range add/remove
func newPortRangeOperation(op string, start, end int64, protocol string, success bool, errCode int64) *OperationEvent {
	return &OperationEvent{
		Namespace:      Namespace,
		Subscope:       Subscope,
		Event:          EventOperation,
		Operation:      op,
		EntryType:      EntryPortRange,
		Protocol:       protocol,
		Result:         analytics.BoolToResult(success),
		Error:          codeToString(errCode),
		PortRangeStart: start,
		PortRangeEnd:   end,
	}
}

// newSubnetOperation creates an event for subnet add/remove.
func newSubnetOperation(op string, subnet string, success bool, errCode int64) *OperationEvent {
	mask, isPrivate := parseSubnetInfo(subnet)
	return &OperationEvent{
		Namespace:       Namespace,
		Subscope:        Subscope,
		Event:           EventOperation,
		Operation:       op,
		EntryType:       EntrySubnet,
		Protocol:        ProtoNA,
		Result:          analytics.BoolToResult(success),
		Error:           codeToString(errCode),
		SubnetMask:      mask,
		IsPrivateSubnet: isPrivate,
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
		Result:    analytics.BoolToResult(success),
		Error:     codeToString(errCode),
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

	var privateCount, publicCount int64
	for _, subnet := range cfg.Subnets {
		_, isPrivate := parseSubnetInfo(subnet)
		if isPrivate {
			privateCount++
		} else {
			publicCount++
		}
	}

	total := tcpCount + udpCount + subnetCount
	return &SnapshotEvent{
		Namespace:          Namespace,
		Subscope:           Subscope,
		Event:              EventSnapshot,
		TCPPorts:           cfg.TCPPorts,
		UDPPorts:           cfg.UDPPorts,
		TCPPortCount:       tcpCount,
		UDPPortCount:       udpCount,
		SubnetCount:        subnetCount,
		PrivateSubnetCount: privateCount,
		PublicSubnetCount:  publicCount,
		TotalEntryCount:    total,
		IsEnabled:          total > 0,
	}
}

// NewOperationEventFromRequest creates an analytics event based on the protobuf request type.
func NewOperationEventFromRequest(
	req *pb.SetAllowlistRequest,
	op string,
	success bool,
	errCode int64,
) *OperationEvent {
	switch request := req.Request.(type) {
	case *pb.SetAllowlistRequest_SetAllowlistSubnetRequest:
		subnet := request.SetAllowlistSubnetRequest.GetSubnet()
		return newSubnetOperation(op, subnet, success, errCode)

	case *pb.SetAllowlistRequest_SetAllowlistPortsRequest:
		portRange := request.SetAllowlistPortsRequest.GetPortRange()
		start := portRange.GetStartPort()
		end := portRange.GetEndPort()

		protocol := ProtoBoth
		if request.SetAllowlistPortsRequest.IsTcp && !request.SetAllowlistPortsRequest.IsUdp {
			protocol = ProtoTCP
		} else if request.SetAllowlistPortsRequest.IsUdp && !request.SetAllowlistPortsRequest.IsTcp {
			protocol = ProtoUDP
		}

		if end == 0 || end == start {
			return newPortOperation(op, start, protocol, success, errCode)
		}
		return newPortRangeOperation(op, start, end, protocol, success, errCode)
	}

	return nil
}

// ToDebuggerEvent converts OperationEvent to a DebuggerEvent for moose publishing
func (e *OperationEvent) ToDebuggerEvent() *events.DebuggerEvent {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to marshal operation-event:", err)
		// Fallback: provide basic information we know for certain
		jsonData = []byte(fmt.Sprintf(
			`{"namespace":"%s","subscope":"%s","event":"%s","operation":"%s","result":"%s","error":"marshal_error"}`,
			e.Namespace, e.Subscope, e.Event, e.Operation, e.Result,
		))
	}
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
			events.ContextValue{Path: contextPathPrefix + ".port_range_start", Value: e.PortRangeStart},
			events.ContextValue{Path: contextPathPrefix + ".port_range_end", Value: e.PortRangeEnd},
			events.ContextValue{Path: contextPathPrefix + ".subnet_mask", Value: e.SubnetMask},
			events.ContextValue{Path: contextPathPrefix + ".is_private_subnet", Value: e.IsPrivateSubnet},
		).
		WithGlobalContextPaths(globalContextPaths...)
}

// ToDebuggerEvent converts SnapshotEvent to a DebuggerEvent for moose publishing
func (e *SnapshotEvent) ToDebuggerEvent() *events.DebuggerEvent {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to marshal snapshot-event:", err)
		// Fallback: provide basic information we know for certain
		jsonData = []byte(fmt.Sprintf(
			`{"namespace":"%s","subscope":"%s","event":"%s","error":"marshal_error"}`,
			e.Namespace, e.Subscope, e.Event,
		))
	}
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: contextPathPrefix + ".namespace", Value: e.Namespace},
			events.ContextValue{Path: contextPathPrefix + ".subscope", Value: e.Subscope},
			events.ContextValue{Path: contextPathPrefix + ".event", Value: e.Event},
			events.ContextValue{Path: contextPathPrefix + ".tcp_ports", Value: e.TCPPorts},
			events.ContextValue{Path: contextPathPrefix + ".udp_ports", Value: e.UDPPorts},
			events.ContextValue{Path: contextPathPrefix + ".tcp_port_count", Value: e.TCPPortCount},
			events.ContextValue{Path: contextPathPrefix + ".udp_port_count", Value: e.UDPPortCount},
			events.ContextValue{Path: contextPathPrefix + ".subnet_count", Value: e.SubnetCount},
			events.ContextValue{Path: contextPathPrefix + ".private_subnet_count", Value: e.PrivateSubnetCount},
			events.ContextValue{Path: contextPathPrefix + ".public_subnet_count", Value: e.PublicSubnetCount},
			events.ContextValue{Path: contextPathPrefix + ".total_entry_count", Value: e.TotalEntryCount},
			events.ContextValue{Path: contextPathPrefix + ".is_enabled", Value: e.IsEnabled},
		).
		WithGlobalContextPaths(globalContextPaths...)
}

// parseSubnetInfo extracts the mask and determines if the subnet is private.
// Returns mask=-1 if parsing fails, isPrivate=false for invalid or public addresses.
func parseSubnetInfo(cidr string) (mask int, isPrivate bool) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return -1, false
	}
	return prefix.Bits(), prefix.Addr().IsPrivate()
}

func codeToString(code int64) string {
	switch code {
	case internal.CodeSuccess:
		return ""
	case internal.CodeFailure:
		return "operation failed"
	case internal.CodeConfigError:
		return "configuration error"
	case internal.CodePrivateSubnetLANDiscovery:
		return "private subnet conflicts with LAN discovery"
	case internal.CodeAllowlistInvalidSubnet:
		return "invalid subnet format"
	case internal.CodeAllowlistSubnetNoop:
		return "subnet unchanged: already in desired state"
	case internal.CodeAllowlistPortOutOfRange:
		return "port out of valid range (1-65535)"
	case internal.CodeAllowlistPortNoop:
		return "port unchanged: already in desired state"
	default:
		return fmt.Sprintf("unknown allowlist error (code %d)", code)
	}
}
