package daemon

import (
	"encoding/json"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/analytics"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	ensNamespace         = internal.DebugEventMessageNamespace
	ensSubscope          = "ens"
	ensEvent             = ensSubscope + "_connection_error"
	ensContextPathPrefix = ensSubscope
)

func vpnConnectionErrorCodeName(code events.VPNConnectionError) string {
	switch code {
	case events.VPNConnectionErrorUnknown:
		return "unknown"
	case events.VPNConnectionErrorConnectionLimitReached:
		return "connection_limit_reached"
	case events.VPNConnectionErrorServerMaintenance:
		return "server_maintenance"
	case events.VPNConnectionErrorUnauthenticated:
		return "unauthenticated"
	case events.VPNConnectionErrorSuperseded:
		return "superseded"
	default:
		return "unrecognized"
	}
}

type vpnConnectionErrorEvent struct {
	Namespace   string `json:"namespace"`
	Subscope    string `json:"subscope"`
	Event       string `json:"event"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// newVPNConnectionErrorEvent builds the debugger-event payload for a connection
// error code.
func newVPNConnectionErrorEvent(code events.VPNConnectionError) *vpnConnectionErrorEvent {
	return &vpnConnectionErrorEvent{
		Namespace:   ensNamespace,
		Subscope:    ensSubscope,
		Event:       ensEvent,
		Code:        vpnConnectionErrorCodeName(code),
		Description: code.String(),
	}
}

// ToDebuggerEvent converts the event to a DebuggerEvent for moose publishing.
func (e *vpnConnectionErrorEvent) ToDebuggerEvent() *events.DebuggerEvent {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to marshal ENS connection error event:", err)
		jsonData = []byte(fmt.Sprintf(
			`{"namespace":"%s","subscope":"%s","event":"%s","code":"%s","error":"marshal_error"}`,
			e.Namespace, e.Subscope, e.Event, e.Code,
		))
	}
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: ensContextPathPrefix + ".namespace", Value: e.Namespace},
			events.ContextValue{Path: ensContextPathPrefix + ".subscope", Value: e.Subscope},
			events.ContextValue{Path: ensContextPathPrefix + ".event", Value: e.Event},
			events.ContextValue{Path: ensContextPathPrefix + ".code", Value: e.Code},
		).
		WithGlobalContextPaths(analytics.MergeContextPaths()...)
}
