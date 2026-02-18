package dns

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var globalPaths = []string{
	"device.*",
	"application.nordvpnapp.*",
	"application.nordvpnapp.version",
	"application.nordvpnapp.platform",
}

const (
	debuggerEventBaseKey              = "dns"
	debuggerEventTypeKey              = "type"
	debuggerEventManagementServiceKey = "management_service"
	debuggerEventErrorTypeKey         = "error_type"
	debuggerEventCriticalKey          = "critical"
)

type event struct {
	Event             string `json:"event"`
	MessageNamespace  string `json:"namespace"`
	ManagementService string `json:"management_service"`
}

func newEvent(eventType eventType, managementService dnsManagementService) event {
	return event{
		Event:             eventType.String(),
		MessageNamespace:  internal.DebugEventMessageNamespace,
		ManagementService: managementService.String(),
	}
}

func (e event) toContextPaths() []events.ContextValue {
	return []events.ContextValue{
		{
			Path:  debuggerEventBaseKey + "." + debuggerEventTypeKey,
			Value: e.Event,
		},
		{
			Path:  debuggerEventBaseKey + "." + debuggerEventManagementServiceKey,
			Value: e.ManagementService,
		},
	}
}

func (e event) toDebuggerEvent() *events.DebuggerEvent {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.DebugPrefix, dnsPrefix, "failed to serialize event json for resolv.conf overwrite:", err)
		jsonData = []byte("{}")
	}

	debuggerEvent := events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(e.toContextPaths()...).
		WithGlobalContextPaths(globalPaths...)

	return debuggerEvent
}

type errorEvent struct {
	event
	ErrorType string `json:"error_type"`
	// Critical should be set to true if the given error prevents DNS configuration
	Critical bool `json:"critical"`
}

func newErrorEvent(eventType eventType,
	managementService dnsManagementService,
	errorType errorType,
	critical bool) errorEvent {
	return errorEvent{
		event:     newEvent(eventType, managementService),
		ErrorType: errorType.String(),
		Critical:  critical,
	}
}

func (e errorEvent) toContextPaths() []events.ContextValue {
	contextPaths := []events.ContextValue{
		{
			Path:  debuggerEventBaseKey + "." + debuggerEventErrorTypeKey,
			Value: e.ErrorType,
		},
		{
			Path:  debuggerEventBaseKey + "." + debuggerEventCriticalKey,
			Value: e.Critical,
		},
	}
	contextPaths = append(contextPaths, e.event.toContextPaths()...)
	return contextPaths
}

func (e errorEvent) toDebuggerEvent() *events.DebuggerEvent {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Println(internal.WarningPrefix,
			dnsPrefix,
			"failed to serialize error event json for resolv.conf overwrite:", err)
		jsonData = []byte("{}")
	}

	debuggerEvent := events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(e.toContextPaths()...).
		WithGlobalContextPaths(globalPaths...)

	return debuggerEvent
}

type eventType int

const (
	resolvConfOverwrittenEventType eventType = iota
	dnsConfiguredEventType
	dnsConfigurationErrorEventType
)

func (e eventType) String() string {
	switch e {
	case resolvConfOverwrittenEventType:
		return "resolvconf_overwritten"
	case dnsConfiguredEventType:
		return "dns_configured"
	case dnsConfigurationErrorEventType:
		return "dns_configuration_error"
	default:
		return fmt.Sprintf("%d", e)
	}
}

type errorType int

const (
	unknonErrorType errorType = iota
	setFailedErrorType
	unsetFailedErrorType
	binaryNotFoundSetErrorType
	resolvConfReadFailedErrorType
	resolvConfStatFailedErrorType
)

func (e errorType) String() string {
	switch e {
	case setFailedErrorType:
		return "set_failed"
	case unsetFailedErrorType:
		return "unset_failed"
	case binaryNotFoundSetErrorType:
		return "binary_not_found"
	case resolvConfReadFailedErrorType:
		return "resolv_conf_read_failed"
	case resolvConfStatFailedErrorType:
		return "resolv_conf_stat_failed"
	case unknonErrorType:
		return "unknown"
	default:
		return fmt.Sprintf("%d", e)
	}
}

type analytics interface {
	emitResolvConfOverwrittenEvent(managementService dnsManagementService)
	emitDNSConfiguredEvent(managementService dnsManagementService)
	emitDNSConfigurationErrorEvent(managementService dnsManagementService, errorType errorType)
	emitDNSConfigurationCriticalErrorEvent(managementService dnsManagementService, errorType errorType)
}

type dnsAnalytics struct {
	debugPublisher events.Publisher[events.DebuggerEvent]
}

func newDNSAnalytics(publisher events.Publisher[events.DebuggerEvent]) *dnsAnalytics {
	return &dnsAnalytics{
		debugPublisher: publisher,
	}
}

func (d *dnsAnalytics) emitResolvConfOverwrittenEvent(managementService dnsManagementService) {
	debuggerEvent := newEvent(resolvConfOverwrittenEventType,
		managementService).toDebuggerEvent()

	log.Printf("%s%s publishing event: %+v", internal.DebugPrefix, dnsPrefix, debuggerEvent)

	d.debugPublisher.Publish(*debuggerEvent)
}

func (d *dnsAnalytics) emitDNSConfiguredEvent(managementService dnsManagementService) {
	debuggerEvent := newEvent(dnsConfiguredEventType,
		managementService).toDebuggerEvent()

	log.Printf("%s%s publishing event: %+v", internal.DebugPrefix, dnsPrefix, debuggerEvent)

	d.debugPublisher.Publish(*debuggerEvent)
}

func (d *dnsAnalytics) emitDNSConfigurationErrorEvent(managementService dnsManagementService, errorType errorType) {
	debuggerEvent := newErrorEvent(dnsConfigurationErrorEventType,
		managementService,
		errorType,
		false).toDebuggerEvent()

	log.Printf("%s%s publishing event: %+v", internal.DebugPrefix, dnsPrefix, debuggerEvent)

	d.debugPublisher.Publish(*debuggerEvent)
}

func (d *dnsAnalytics) emitDNSConfigurationCriticalErrorEvent(managementService dnsManagementService,
	errorType errorType) {
	debuggerEvent := newErrorEvent(dnsConfigurationErrorEventType,
		managementService,
		errorType,
		true).toDebuggerEvent()

	log.Printf("%s%s publishing event: %+v", internal.DebugPrefix, dnsPrefix, debuggerEvent)

	d.debugPublisher.Publish(*debuggerEvent)
}
