package dns

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockevents "github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"github.com/stretchr/testify/assert"
)

func Test_emitResolvConfOverwrittenEvent(t *testing.T) {
	category.Set(t, category.Unit)

	mockPublisher := mockevents.MockPublisher[events.DebuggerEvent]{}
	analytics := newDNSAnalytics(&mockPublisher)
	analytics.emitResolvConfOverwrittenEvent(unknownManagementService)

	event, n, stackIsEmpty := mockPublisher.PopEvent()

	assert.True(t, stackIsEmpty, "Event not emitted.")
	assert.Equal(t, 0, n, "Unexpected number of events emitted.")
	assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
		Path:  debuggerEventBaseKey + "." + debuggerEventTypeKey,
		Value: resolvConfOverwrittenEventType.String()})
	assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
		Path:  debuggerEventBaseKey + "." + debuggerEventManagementServiceKey,
		Value: unknownManagementService.String()})
	assert.Equal(t,
		"{\"event\":\"resolvconf_overwritten\",\"namespace\":\"nordvpn-linux\",\"management_service\":\"unknown_service\"}",
		event.JsonData)
}

func Test_emitDNSConfiguredEvent(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		managementService dnsManagementService
	}{
		{
			name:              "systemd-resolved",
			managementService: systemdResolvedManagementService,
		},
		{
			name:              "unmanaged",
			managementService: unmanagedManagementService,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPublisher := mockevents.MockPublisher[events.DebuggerEvent]{}
			analytics := newDNSAnalytics(&mockPublisher)
			analytics.emitDNSConfiguredEvent(test.managementService)

			event, n, stackIsEmpty := mockPublisher.PopEvent()

			assert.True(t, stackIsEmpty, "Event not emitted.")
			assert.Equal(t, 0, n, "Unexpected number of events emitted.")
			assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventTypeKey,
				Value: dnsConfiguredEventType.String()})
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

func Test_emitDNSConfigurationErrorEvent(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		managementService dnsManagementService
		errorType         errorType
		critical          bool
	}{
		{
			name:              "set failed for unmanaged, not critical",
			managementService: unmanagedManagementService,
			errorType:         setFailedErrorType,
			critical:          false,
		},
		{
			name:              "failed to read for unmanaged, not critical",
			managementService: unmanagedManagementService,
			errorType:         resolvConfReadFailedErrorType,
			critical:          false,
		},
		{
			name:              "set failed for unmanaged, critical",
			managementService: unmanagedManagementService,
			errorType:         setFailedErrorType,
			critical:          true,
		},
		{
			name:              "failed to read for unmanaged, critical",
			managementService: unmanagedManagementService,
			errorType:         resolvConfReadFailedErrorType,
			critical:          true,
		},
		{
			name:              "set failed for systemd-resolved, not critical",
			managementService: systemdResolvedManagementService,
			errorType:         setFailedErrorType,
			critical:          false,
		},
		{
			name:              "failed to read for systemd-resolved, not critical",
			managementService: systemdResolvedManagementService,
			errorType:         resolvConfReadFailedErrorType,
			critical:          false,
		},
		{
			name:              "set failed for systemd-resolved, critical",
			managementService: systemdResolvedManagementService,
			errorType:         setFailedErrorType,
			critical:          true,
		},
		{
			name:              "failed to read for systemd-resolved, critical",
			managementService: systemdResolvedManagementService,
			errorType:         resolvConfReadFailedErrorType,
			critical:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPublisher := mockevents.MockPublisher[events.DebuggerEvent]{}
			analytics := newDNSAnalytics(&mockPublisher)
			if test.critical {
				analytics.emitDNSConfigurationCriticalErrorEvent(test.managementService, test.errorType)
			} else {
				analytics.emitDNSConfigurationErrorEvent(test.managementService, test.errorType)
			}

			event, n, stackIsEmpty := mockPublisher.PopEvent()

			assert.True(t, stackIsEmpty, "Event not emitted.")
			assert.Equal(t, 0, n, "Unexpected number of events emitted.")
			assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventTypeKey,
				Value: dnsConfigurationErrorEventType.String()})
			assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventManagementServiceKey,
				Value: test.managementService.String()})
			assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventErrorTypeKey,
				Value: test.errorType.String(),
			})
			assert.Contains(t, event.KeyBasedContextPaths, events.ContextValue{
				Path:  debuggerEventBaseKey + "." + debuggerEventCriticalKey,
				Value: test.critical,
			})

			expectedJson :=
				fmt.Sprintf("{\"event\":\"dns_configuration_error\",\"namespace\":\"nordvpn-linux\",\"management_service\":\"%s\",\"error_type\":\"%s\",\"critical\":%s}",
					test.managementService.String(),
					test.errorType.String(),
					strconv.FormatBool(test.critical))
			assert.Equal(t,
				expectedJson,
				event.JsonData)
		})
	}
}
