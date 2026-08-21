package openvpn

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockevents "github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func contextValueByPath(paths []events.ContextValue, path string) any {
	for _, cv := range paths {
		if cv.Path == path {
			return cv.Value
		}
	}
	return nil
}

func TestDCOStatusEvent_ToDebuggerEvent(t *testing.T) {
	category.Set(t, category.Unit)

	event := &dcoStatusEvent{
		Namespace:       ovpnNamespace,
		Subscope:        ovpnSubscope,
		Event:           dcoStatusEventName,
		DCOActive:       true,
		LinkKind:        "ovpn-dco",
		ModuleAvailable: true,
		ModuleVersion:   "0.2.20260519",
	}
	debuggerEvent := event.ToDebuggerEvent()

	require.NotNil(t, debuggerEvent)
	require.NotEmpty(t, debuggerEvent.JsonData)

	var decoded dcoStatusEvent
	require.NoError(t, json.Unmarshal([]byte(debuggerEvent.JsonData), &decoded))
	assert.Equal(t, *event, decoded)

	assert.Equal(t, "nordvpn-linux", decoded.Namespace)
	assert.Equal(t, "openvpn", decoded.Subscope)
	assert.Equal(t, "openvpn_dco_status", decoded.Event)

	assert.Equal(t, true, contextValueByPath(debuggerEvent.KeyBasedContextPaths, "openvpn.dco_active"))
	assert.Equal(t, "ovpn-dco", contextValueByPath(debuggerEvent.KeyBasedContextPaths, "openvpn.link_kind"))
	assert.Equal(t, true, contextValueByPath(debuggerEvent.KeyBasedContextPaths, "openvpn.module_available"))
	assert.NotEmpty(t, debuggerEvent.GeneralContextPaths)
}

func TestNewDCOStatusEvent(t *testing.T) {
	category.Set(t, category.Unit)

	event := newDCOStatusEvent()

	assert.Equal(t, "nordvpn-linux", event.Namespace)
	assert.Equal(t, "openvpn", event.Subscope)
	assert.Equal(t, "openvpn_dco_status", event.Event)
	assert.NotEmpty(t, event.LinkKind)
}

func TestDCOAnalytics_NotifyConnect(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		event         events.DataConnect
		wantPublished bool
	}{
		{
			name: "successful openvpn connect publishes DCO status",
			event: events.DataConnect{
				EventStatus: events.StatusSuccess,
				Technology:  config.Technology_OPENVPN,
			},
			wantPublished: true,
		},
		{
			name: "successful nordlynx connect publishes nothing",
			event: events.DataConnect{
				EventStatus: events.StatusSuccess,
				Technology:  config.Technology_NORDLYNX,
			},
			wantPublished: false,
		},
		{
			name: "openvpn connect attempt publishes nothing",
			event: events.DataConnect{
				EventStatus: events.StatusAttempt,
				Technology:  config.Technology_OPENVPN,
			},
			wantPublished: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := &mockevents.MockPublisherSubscriber[events.DebuggerEvent]{}

			err := NewDCOAnalytics(publisher).NotifyConnect(tt.event)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantPublished, publisher.EventPublished)
			if tt.wantPublished {
				assert.Contains(t, publisher.Event.JsonData, `"event":"openvpn_dco_status"`)
			}
		})
	}
}
