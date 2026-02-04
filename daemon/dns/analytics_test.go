package dns

import (
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
		Path:  debuggerEventBaseKey + ".type",
		Value: resolvConfOverwritten.String()})
	assert.Equal(t, "{\"event\":\"resolvconf_overwritten\",\"namespace\":\"nordvpn-linux\"}", event.JsonData)
}
