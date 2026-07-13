package ens

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"gotest.tools/v3/assert"
)

func TestENSMonitoring(t *testing.T) {
	category.Set(t, category.Unit)

	const serverEndpoint = "192.168.1.1:51820"

	ctx, cancelFn := context.WithCancel(context.Background())
	netw := &networker.Mock{
		VpnActive:        true,
		ActiveServerData: &vpn.ServerData{Endpoint: serverEndpoint},
	}
	rc := mock.NewRemoteConfigMock()
	rc.FeatureToggles[remote.FeatureENS] = true

	callbackCount := atomic.Int32{}
	ch := make(chan any)
	monitor := NewMonitor(ctx, netw, rc, func(_ string) error {
		callbackCount.Add(1)
		ch <- 1
		return nil
	}, &subs.Subject[events.DebuggerEvent]{})
	monitor.Start()

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorServerMaintenance,
		ServerEndpoint: serverEndpoint,
	}))
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 1, int(callbackCount.Load()))

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorServerMaintenance,
		ServerEndpoint: serverEndpoint,
	}))
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 2, int(callbackCount.Load()))

	cancelFn()

	// after stopping the monitoring, events are ignored
	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorServerMaintenance,
		ServerEndpoint: serverEndpoint,
	}))
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 2, int(callbackCount.Load()))
}

func TestENSMonitoringEventHandling(t *testing.T) {
	category.Set(t, category.Unit)

	const serverEndpoint = "192.168.1.1:51820"

	tests := []struct {
		name            string
		ensEnabled      bool
		serverEndpoint  string
		event           events.VPNConnectionErrorEvent
		expectReport    bool
		expectReconnect bool
	}{
		{
			name:            "maintenance event for current server is reported and reconnects",
			ensEnabled:      true,
			serverEndpoint:  serverEndpoint,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerEndpoint: serverEndpoint},
			expectReport:    true,
			expectReconnect: true,
		},
		{
			name:            "maintenance event with stale server endpoint is reported but does not reconnect",
			ensEnabled:      true,
			serverEndpoint:  serverEndpoint,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerEndpoint: "10.0.0.1:51820"},
			expectReport:    true,
			expectReconnect: false,
		},
		{
			name:            "disabled ENS feature reports nothing and does not reconnect",
			ensEnabled:      false,
			serverEndpoint:  serverEndpoint,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerEndpoint: serverEndpoint},
			expectReport:    false,
			expectReconnect: false,
		},
		{
			name:            "superseded error is reported but does not reconnect",
			ensEnabled:      true,
			serverEndpoint:  serverEndpoint,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorSuperseded, ServerEndpoint: serverEndpoint},
			expectReport:    true,
			expectReconnect: false,
		},
		{
			name:            "unknown error is reported but does not reconnect",
			ensEnabled:      true,
			serverEndpoint:  serverEndpoint,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorUnknown, ServerEndpoint: serverEndpoint},
			expectReport:    true,
			expectReconnect: false,
		},
		{
			name:            "connection limit error is reported but does not reconnect",
			ensEnabled:      true,
			serverEndpoint:  serverEndpoint,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorConnectionLimitReached, ServerEndpoint: serverEndpoint},
			expectReport:    true,
			expectReconnect: false,
		},
		{
			name:            "unauthenticated error is reported but does not reconnect",
			ensEnabled:      true,
			serverEndpoint:  serverEndpoint,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorUnauthenticated, ServerEndpoint: serverEndpoint},
			expectReport:    true,
			expectReconnect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netw := &networker.Mock{
				VpnActive:        true,
				ActiveServerData: &vpn.ServerData{Endpoint: tt.serverEndpoint},
			}
			rc := mock.NewRemoteConfigMock()
			rc.FeatureToggles[remote.FeatureENS] = tt.ensEnabled

			reported := make(chan bool, 1)
			debuggerEvents := &subs.Subject[events.DebuggerEvent]{}
			debuggerEvents.Subscribe(func(events.DebuggerEvent) error {
				reported <- true
				return nil
			})
			reconnected := make(chan bool, 1)
			monitor := NewMonitor(t.Context(), netw, rc,
				func(_ string) error {
					reconnected <- true
					return nil
				},
				debuggerEvents,
			)
			monitor.Start()

			assert.NilError(t, monitor.HandleENSNotification(tt.event))
			assert.Equal(t, tt.expectReport, helpers.WaitWithTimeout(t, reported, time.Millisecond*50))
			assert.Equal(t, tt.expectReconnect, helpers.WaitWithTimeout(t, reconnected, time.Millisecond*50))
		})
	}
}
