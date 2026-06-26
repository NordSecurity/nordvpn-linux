package ens

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"gotest.tools/v3/assert"
)

func TestENSMonitoring(t *testing.T) {
	category.Set(t, category.Unit)

	const serverKey = "current-server-key"

	ctx, cancelFn := context.WithCancel(context.Background())
	netw := &networker.Mock{
		VpnActive:        true,
		ActiveServerData: &vpn.ServerData{NordLynxPublicKey: serverKey},
	}
	rc := mock.NewRemoteConfigMock()
	rc.FeatureToggles[remote.FeatureENS] = true

	callbackCount := atomic.Int32{}
	ch := make(chan any)
	monitor := NewMonitor(ctx, netw, rc, func(_ string) error {
		callbackCount.Add(1)
		ch <- 1
		return nil
	}, func(events.VPNConnectionError) {})
	monitor.Start()

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:            events.VPNConnectionErrorServerMaintenance,
		ServerPublicKey: serverKey,
	}))
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 1, int(callbackCount.Load()))

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:            events.VPNConnectionErrorServerMaintenance,
		ServerPublicKey: serverKey,
	}))
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 2, int(callbackCount.Load()))

	cancelFn()

	// after stopping the monitoring, events are ignored
	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:            events.VPNConnectionErrorServerMaintenance,
		ServerPublicKey: serverKey,
	}))
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 2, int(callbackCount.Load()))
}

func TestENSMonitoringEventHandling(t *testing.T) {
	category.Set(t, category.Unit)

	const serverKey = "server-key"

	tests := []struct {
		name            string
		ensEnabled      bool
		serverKey       string
		event           events.VPNConnectionErrorEvent
		expectReport    bool
		expectReconnect bool
	}{
		{
			name:            "maintenance event for current server is reported and reconnects",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerPublicKey: serverKey},
			expectReport:    true,
			expectReconnect: true,
		},
		{
			name:            "maintenance event with stale server key is reported but does not reconnect",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerPublicKey: "stale-key"},
			expectReport:    true,
			expectReconnect: false,
		},
		{
			name:            "disabled ENS feature reports nothing and does not reconnect",
			ensEnabled:      false,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerPublicKey: serverKey},
			expectReport:    false,
			expectReconnect: false,
		},
		{
			name:            "superseded error is reported but does not reconnect",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorSuperseded, ServerPublicKey: serverKey},
			expectReport:    true,
			expectReconnect: false,
		},
		{
			name:            "unknown error is reported but does not reconnect",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorUnknown, ServerPublicKey: serverKey},
			expectReport:    true,
			expectReconnect: false,
		},
		{
			name:            "connection limit error is reported but does not reconnect",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorConnectionLimitReached, ServerPublicKey: serverKey},
			expectReport:    true,
			expectReconnect: false,
		},
		{
			name:            "unauthenticated error is reported but does not reconnect",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorUnauthenticated, ServerPublicKey: serverKey},
			expectReport:    true,
			expectReconnect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netw := &networker.Mock{
				VpnActive:        true,
				ActiveServerData: &vpn.ServerData{NordLynxPublicKey: tt.serverKey},
			}
			rc := mock.NewRemoteConfigMock()
			rc.FeatureToggles[remote.FeatureENS] = tt.ensEnabled

			reported := make(chan events.VPNConnectionError, 1)
			reconnected := make(chan struct{}, 1)
			monitor := NewMonitor(t.Context(), netw, rc,
				func(_ string) error {
					reconnected <- struct{}{}
					return nil
				},
				func(code events.VPNConnectionError) {
					reported <- code
				},
			)
			monitor.Start()

			assert.NilError(t, monitor.HandleENSNotification(tt.event))

			var reportedCode events.VPNConnectionError
			gotReport := false
			select {
			case reportedCode = <-reported:
				gotReport = true
			case <-time.After(time.Millisecond * 50):
			}
			assert.Equal(t, tt.expectReport, gotReport)
			if tt.expectReport {
				assert.Equal(t, tt.event.Code, reportedCode)
			}

			gotReconnect := false
			select {
			case <-reconnected:
				gotReconnect = true
			case <-time.After(time.Millisecond * 50):
			}
			assert.Equal(t, tt.expectReconnect, gotReconnect)
		})
	}
}
