package openvpn

import (
	"encoding/json"
	"fmt"

	"github.com/vishvananda/netlink"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/analytics"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
)

const (
	ovpnNamespace         = internal.DebugEventMessageNamespace
	ovpnSubscope          = "openvpn"
	dcoStatusEventName    = ovpnSubscope + "_dco_status"
	ovpnContextPathPrefix = ovpnSubscope

	// linkKindUnknown is reported when the tunnel interface kind could not be read.
	linkKindUnknown = "unknown"
)

// dcoStatusEvent is the debugger-event payload describing whether a successful
// OpenVPN connection runs with kernel data channel offload (DCO).
type dcoStatusEvent struct {
	Namespace string `json:"namespace"`
	Subscope  string `json:"subscope"`
	Event     string `json:"event"`
	// LinkKind is the tunnel interface's link kind.
	// out of tree module is not supported by OpenVPN 2.7+
	// in-tree module (kernel >= 6.16), used by OpenVPN 2.7+ : ovpn
	// non dco interface : tuntap
	LinkKind string `json:"link_kind"`
	KernelVersion string `json:"module_version,omitempty"`
}

func getLinkKind() (string, error) {
	link, err := netlink.LinkByName(InterfaceName)
	if err != nil {
		return "", err
	}
	log.Debug("DCO: printing link type: ", link.Type())
	return link.Type(), nil
}

// newDCOStatusEvent inspects the established tunnel and the kernel.
func newDCOStatusEvent() *dcoStatusEvent {
	linkKind, err := getLinkKind()
	if err != nil {
		log.Error("reading tunnel interface link kind:", err)
		linkKind = linkKindUnknown
	}
	kernelVersion := sysinfo.GetKernelVersion()

	event := &dcoStatusEvent{
		Namespace: ovpnNamespace,
		Subscope:  ovpnSubscope,
		Event:     dcoStatusEventName,
		LinkKind:  linkKind,
		KernelVersion: kernelVersion,
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
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Error("failed to marshal dco status event:", err)
		// Fallback
		jsonData = fmt.Appendf(nil,
			`{"namespace":"%s","subscope":"%s","event":"%s","error":"marshal_error"}`,
			ovpnNamespace, ovpnSubscope, dcoStatusEventName,
		)
	}
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: ovpnContextPathPrefix + ".namespace", Value: e.Namespace},
			events.ContextValue{Path: ovpnContextPathPrefix + ".subscope", Value: e.Subscope},
			events.ContextValue{Path: ovpnContextPathPrefix + ".event", Value: e.Event},
			events.ContextValue{Path: ovpnContextPathPrefix + ".link_kind", Value: e.LinkKind},
		).
		WithGlobalContextPaths(analytics.MergeContextPaths()...)
}
