package firewall

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/analytics"
)

// Event identification constants
const (
	Namespace = internal.DebugEventMessageNamespace
	Subscope  = "firewall"
	// Events types
	EventConfigure       = Subscope + "_configure"
	EventNftablesApply   = Subscope + "_nftables_apply"
	EventIptablesCleanup = Subscope + "_iptables_cleanup"

	contextPathPrefix = "firewall"
)

// Purpose identifies what the firewall is being configured for
const (
	PurposeVPN        = "vpn"
	PurposeMeshnet    = "meshnet"
	PurposeKillSwitch = "killswitch"
	PurposeAllowlist  = "allowlist"
)

// globalContextPaths defines the common context paths included in all firewall events.
var globalContextPaths = analytics.MergeContextPaths(
	// User preference states (settings vs actual config)
	"application.nordvpnapp.config.user_preferences.kill_switch_enabled.value",
	"application.nordvpnapp.config.user_preferences.meshnet_enabled.value",
	"application.nordvpnapp.config.user_preferences.split_tunneling_enabled.value",
	"application.nordvpnapp.config.current_state.is_on_vpn.value",
)

// ConfigureEvent represents a firewall configuration operation.
// Emitted each time nftables firewall is configured.
type ConfigureEvent struct {
	Namespace string   `json:"namespace"`
	Subscope  string   `json:"subscope"`
	Event     string   `json:"event"`
	Status    string   `json:"status"`
	Purpose   []string `json:"purpose"`
	Error     string   `json:"error,omitempty"`
}

// newConfigureEvent creates a new firewall configuration event.
func newConfigureEvent(config Config, err error) *ConfigureEvent {
	purposes := determinePurposes(config)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return &ConfigureEvent{
		Namespace: Namespace,
		Subscope:  Subscope,
		Event:     EventConfigure,
		Status:    analytics.BoolToResult(err == nil),
		Purpose:   purposes,
		Error:     errMsg,
	}
}

// determinePurposes analyzes the firewall config to determine what it's being configured for.
func determinePurposes(config Config) []string {
	var purposes []string

	if len(config.TunnelInterface) > 0 {
		purposes = append(purposes, PurposeVPN)
	}
	if config.MeshnetInfo != nil {
		purposes = append(purposes, PurposeMeshnet)
	}
	if config.KillSwitch {
		purposes = append(purposes, PurposeKillSwitch)
	}
	if hasAllowlist(config.Allowlist) {
		purposes = append(purposes, PurposeAllowlist)
	}

	return purposes
}

// hasAllowlist checks if the allowlist has any configured ports or subnets.
func hasAllowlist(allowlist config.Allowlist) bool {
	return len(allowlist.Ports.TCP) > 0 ||
		len(allowlist.Ports.UDP) > 0 ||
		len(allowlist.Subnets) > 0
}

// ToDebuggerEvent converts ConfigureEvent to a DebuggerEvent for moose publishing.
func (e *ConfigureEvent) ToDebuggerEvent() *events.DebuggerEvent {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to marshal firewall configure event:", err)
		// Fallback: provide basic information we know for certain
		jsonData = []byte(fmt.Sprintf(
			`{"namespace":"%s","subscope":"%s","event":"%s","status":"%s","purpose":[],"error":"marshal_error"}`,
			e.Namespace, e.Subscope, e.Event, e.Status,
		))
	}
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: contextPathPrefix + ".namespace", Value: e.Namespace},
			events.ContextValue{Path: contextPathPrefix + ".subscope", Value: e.Subscope},
			events.ContextValue{Path: contextPathPrefix + ".event", Value: e.Event},
			events.ContextValue{Path: contextPathPrefix + ".status", Value: e.Status},
			events.ContextValue{Path: contextPathPrefix + ".error", Value: e.Error},
		).
		WithGlobalContextPaths(globalContextPaths...)
}

// MigrationEvent represents a firewall migration operation event.
// It's a one-time process
type MigrationEvent struct {
	Namespace string `json:"namespace"`
	Subscope  string `json:"subscope"`
	Event     string `json:"event"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// NewNftablesApplyEvent creates an event for nftables rule application.
// Emitted after attempting to apply nftables rules during migration.
func NewNftablesApplyEvent(err error) *MigrationEvent {
	return newMigrationEvent(EventNftablesApply, err)
}

// NewIptablesCleanupEvent creates an event for iptables rule cleanup.
// Emitted after attempting to remove iptables rules during migration.
func NewIptablesCleanupEvent(err error) *MigrationEvent {
	return newMigrationEvent(EventIptablesCleanup, err)
}

// newMigrationEvent creates a new migration event.
func newMigrationEvent(eventType string, err error) *MigrationEvent {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return &MigrationEvent{
		Namespace: Namespace,
		Subscope:  Subscope,
		Event:     eventType,
		Status:    analytics.BoolToResult(err == nil),
		Error:     errMsg,
	}
}

// ToDebuggerEvent converts MigrationEvent to a DebuggerEvent for moose publishing.
func (e *MigrationEvent) ToDebuggerEvent() *events.DebuggerEvent {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to marshal migration event:", err)
		// Fallback: provide basic information we know for certain
		jsonData = []byte(fmt.Sprintf(
			`{"namespace":"%s","subscope":"%s","event":"%s","status":"%s","error":"marshal_error"}`,
			e.Namespace, e.Subscope, e.Event, e.Status,
		))
	}
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: contextPathPrefix + ".namespace", Value: e.Namespace},
			events.ContextValue{Path: contextPathPrefix + ".subscope", Value: e.Subscope},
			events.ContextValue{Path: contextPathPrefix + ".event", Value: e.Event},
			events.ContextValue{Path: contextPathPrefix + ".status", Value: e.Status},
			events.ContextValue{Path: contextPathPrefix + ".error", Value: e.Error},
		).
		WithGlobalContextPaths(globalContextPaths...)
}
