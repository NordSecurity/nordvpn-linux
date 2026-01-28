package dns

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type event struct {
	Event            string `json:"event"`
	MessageNamespace string `json:"namespace"`
}

type eventType int

const (
	resolvConfOverwritten eventType = iota
)

func (e eventType) String() string {
	switch e {
	case resolvConfOverwritten:
		return "resolvconf_overwritten"
	default:
		return fmt.Sprintf("%d", e)
	}
}

const (
	debuggerEventBaseKey = "dns"
)

// analytics provides an interface for sending DNS related debugger events
type analytics interface {
	emitResolvConfOverwrittenEvent()
}

type dnsAnalytics struct {
	debugPublisher events.Publisher[events.DebuggerEvent]
}

func newDNSAnalytics(publisher events.Publisher[events.DebuggerEvent]) *dnsAnalytics {
	return &dnsAnalytics{
		debugPublisher: publisher,
	}
}

func (d *dnsAnalytics) emitResolvConfOverwrittenEvent() {
	debuggerEvent := buildDebuggerEvent(resolvConfOverwritten)

	log.Printf("%s%s publishing event: %+v", internal.DebugPrefix, dnsPrefix, debuggerEvent)

	d.debugPublisher.Publish(*debuggerEvent)
}

// buildDebuggerEvent creates a debugger event for the provided eventType
func buildDebuggerEvent(eventType eventType) *events.DebuggerEvent {
	e := event{
		Event:            eventType.String(),
		MessageNamespace: internal.DebugEventMessageNamespace,
	}

	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.DebugPrefix, dnsPrefix, "failed to serialize event json for resovl.conf overwrite:", err)
		jsonData = []byte("{}")
	}

	debuggerEvent := events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(events.ContextValue{
			Path:  debuggerEventBaseKey + ".type",
			Value: resolvConfOverwritten.String()}).
		WithGlobalContextPaths("device.*",
			"application.nordvpnapp.*",
			"application.nordvpnapp.version",
			"application.nordvpnapp.platform")

	return debuggerEvent
}
