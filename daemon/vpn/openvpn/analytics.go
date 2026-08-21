package openvpn

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/vishvananda/netlink"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/analytics"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	ovpnNamespace         = internal.DebugEventMessageNamespace
	ovpnSubscope          = "openvpn"
	dcoStatusEventName    = ovpnSubscope + "_dco_status"
	ovpnContextPathPrefix = ovpnSubscope

	// linkKindUnknown is reported when the tunnel interface kind could not be read.
	linkKindUnknown = "unknown"
)

// dcoLinkKinds lists link types used by OpenVPN DCO tunnel interfaces.
var dcoLinkKinds = map[string]bool{
	"ovpn-dco": true,
	"ovpn":     true,
}

// dcoModuleSysfsDir exists while the ovpn-dco kernel module is loaded.
const dcoModuleSysfsDir = "/sys/module/ovpn_dco_v2"

// dcoStatusEvent is the debugger-event payload describing whether a successful
// OpenVPN connection runs with kernel data channel offload (DCO).
type dcoStatusEvent struct {
	Namespace string `json:"namespace"`
	Subscope  string `json:"subscope"`
	Event     string `json:"event"`
	// DCOActive reports whether a DCO kernel module is in use.
	DCOActive bool `json:"dco_active"`
	// LinkKind is the tunnel interface's link kind.
	LinkKind string `json:"link_kind"`
	// ModuleAvailable reports whether a DCO kernel module is loaded, regardless of use.
	ModuleAvailable bool `json:"module_available"`
	// ModuleVersion is the loaded module's version.
	ModuleVersion string `json:"module_version,omitempty"`
}

func getLinkKind() (string, error) {
	link, err := netlink.LinkByName(InterfaceName)
	if err != nil {
		return "", err
	}
	return link.Type(), nil
}

// newDCOStatusEvent inspects the established tunnel and the kernel.
func newDCOStatusEvent() *dcoStatusEvent {
	linkKind, err := getLinkKind()
	if err != nil {
		log.Error("reading tunnel interface link kind:", err)
		linkKind = linkKindUnknown
	}

	event := &dcoStatusEvent{
		Namespace: ovpnNamespace,
		Subscope:  ovpnSubscope,
		Event:     dcoStatusEventName,
		DCOActive: dcoLinkKinds[linkKind],
		LinkKind:  linkKind,
	}

	if internal.FileExists(dcoModuleSysfsDir) {
		event.ModuleAvailable = true
		if version, err := os.ReadFile(filepath.Join(dcoModuleSysfsDir, "version")); err == nil {
			event.ModuleVersion = strings.TrimSpace(string(version))
		}
	}
	return event
}

type DCOAnalytics struct {
	publisher events.Publisher[events.DebuggerEvent]
}

func NewDCOAnalytics(publisher events.Publisher[events.DebuggerEvent]) *DCOAnalytics {
	return &DCOAnalytics{publisher: publisher}
}

// NotifyConnect emits the DCO status event once per successful OpenVPN connection
func (a *DCOAnalytics) NotifyConnect(e events.DataConnect) error {
	if e.EventStatus == events.StatusSuccess && e.Technology == config.Technology_OPENVPN {
		a.publisher.Publish(*newDCOStatusEvent().ToDebuggerEvent())
	}
	return nil
}

// ToDebuggerEvent converts the event to a DebuggerEvent for publishing.
func (e *dcoStatusEvent) ToDebuggerEvent() *events.DebuggerEvent {
	// cannot fail: the struct holds only strings and bools
	jsonData, _ := json.Marshal(e)
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: ovpnContextPathPrefix + ".namespace", Value: e.Namespace},
			events.ContextValue{Path: ovpnContextPathPrefix + ".subscope", Value: e.Subscope},
			events.ContextValue{Path: ovpnContextPathPrefix + ".event", Value: e.Event},
			events.ContextValue{Path: ovpnContextPathPrefix + ".dco_active", Value: e.DCOActive},
			events.ContextValue{Path: ovpnContextPathPrefix + ".link_kind", Value: e.LinkKind},
			events.ContextValue{Path: ovpnContextPathPrefix + ".module_available", Value: e.ModuleAvailable},
		).
		WithGlobalContextPaths(analytics.MergeContextPaths()...)
}
