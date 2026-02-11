package dns

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events"
	mockevents "github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"github.com/stretchr/testify/assert"
)

func Test_emitResolvConfOverwrittenEvent(t *testing.T) {
	mockPublisher := mockevents.MockPublisher[events.DebuggerEvent]{}
	analytics := newDNSAnalytics(&mockPublisher)
	analytics.emitResolvConfOverwrittenEvent()

	event, n, stackIsEmpty := mockPublisher.PopEvent()

	assert.True(t, stackIsEmpty, "Event not emitted.")
	assert.Equal(t, 0, n, "Unexpected number of events emitted.")
	assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
		Path:  debuggerEventBaseKey + "." + debuggerEventTypeKey,
		Value: resolvConfOverwritten.String()})
	assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
		Path:  debuggerEventBaseKey + "." + debuggerEventManagementServiceKey,
		Value: unknown.String()})
	assert.Equal(t,
		"{\"event\":\"resolvconf_overwritten\",\"namespace\":\"nordvpn-linux\",\"management_service\":\"unknown_service\"}",
		event.JsonData)
}

func Test_emitDNSConfiguredEvent(t *testing.T) {
	tests := []struct {
		name              string
		managementService dnsManagementService
	}{
		{
			name:              "systemd-resolved",
			managementService: systemdResolved,
		},
		{
			name:              "unmanaged",
			managementService: unmanaged,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPublisher := mockevents.MockPublisher[events.DebuggerEvent]{}
			analytics := newDNSAnalytics(&mockPublisher)
			analytics.setManagementService(test.managementService)
			analytics.emitDNSConfiguredEvent()

			event, n, stackIsEmpty := mockPublisher.PopEvent()

			assert.True(t, stackIsEmpty, "Event not emitted.")
			assert.Equal(t, 0, n, "Unexpected number of events emitted.")
			assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventTypeKey,
				Value: dnsConfigured.String()})
			assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventManagementServiceKey,
				Value: test.managementService.String()})

			expectedJson :=
				fmt.Sprintf("{\"event\":\"dns_configured\",\"namespace\":\"nordvpn-linux\",\"management_service\":\"%s\"}",
					test.managementService.String())
			assert.Equal(t,
				expectedJson,
				event.JsonData)
		})
	}
}
