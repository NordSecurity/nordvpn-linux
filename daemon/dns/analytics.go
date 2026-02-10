package dns

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type event struct {
	Event             string `json:"event"`
	MessageNamespace  string `json:"namespace"`
	ManagementService string `json:"management_service"`
}

type eventType int

const (
	resolvConfOverwritten eventType = iota
	dnsConfigured
)

func (e eventType) String() string {
	switch e {
	case resolvConfOverwritten:
		return "resolvconf_overwritten"
	case dnsConfigured:
		return "dns_configured"
	default:
		return fmt.Sprintf("undefined, id: %d", e)
	}
}

const (
	debuggerEventBaseKey              = "dns"
	debuggerEventTypeKey              = "type"
	debuggerEventManagementServiceKey = "management_service"
)

// analytics provides an interface for sending DNS related debugger events
type analytics interface {
	setManagementService(dnsManagementService)
	emitResolvConfOverwrittenEvent()
	emitDNSConfiguredEvent()
}

type dnsAnalytics struct {
	mu                sync.Mutex
	debugPublisher    events.Publisher[events.DebuggerEvent]
	managementService dnsManagementService
}

func newDNSAnalytics(publisher events.Publisher[events.DebuggerEvent]) *dnsAnalytics {
	return &dnsAnalytics{
		debugPublisher:    publisher,
		managementService: unknown,
	}
}

// setManagementService sets management service to be used in the context of DNS related debugger events
func (d *dnsAnalytics) setManagementService(managementService dnsManagementService) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.managementService = managementService
}

func (d *dnsAnalytics) emitResolvConfOverwrittenEvent() {
	d.mu.Lock()
	defer d.mu.Unlock()

	debuggerEvent := buildDebuggerEvent(resolvConfOverwritten, d.managementService)

	log.Printf("%s%s publishing event: %+v", internal.DebugPrefix, dnsPrefix, debuggerEvent)

	d.debugPublisher.Publish(*debuggerEvent)
}

func (d *dnsAnalytics) emitDNSConfiguredEvent() {
	d.mu.Lock()
	defer d.mu.Unlock()

	debuggerEvent := buildDebuggerEvent(dnsConfigured, d.managementService)

	log.Printf("%s%s publishing event: %+v", internal.DebugPrefix, dnsPrefix, debuggerEvent)

	d.debugPublisher.Publish(*debuggerEvent)
}

// buildDebuggerEvent creates a debugger event for the provided eventType
func buildDebuggerEvent(eventType eventType, managementService dnsManagementService) *events.DebuggerEvent {
	e := event{
		Event:             eventType.String(),
		MessageNamespace:  internal.DebugEventMessageNamespace,
		ManagementService: managementService.String(),
	}

	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.DebugPrefix, dnsPrefix, "failed to serialize event json for resovl.conf overwrite:", err)
		jsonData = []byte("{}")
	}

	debuggerEvent := events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventTypeKey,
				Value: eventType.String()},
			events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventManagementServiceKey,
				Value: managementService.String(),
			},
		).
		WithGlobalContextPaths("device.*",
			"application.nordvpnapp.*",
			"application.nordvpnapp.version",
			"application.nordvpnapp.platform")

	return debuggerEvent
}
