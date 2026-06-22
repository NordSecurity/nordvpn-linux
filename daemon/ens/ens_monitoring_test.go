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
		VpnActive:               true,
		ActiveServerDataPresent: true,
		ActiveServerData:        vpn.ServerData{NordLynxPublicKey: serverKey},
	}
	rc := mock.NewRemoteConfigMock()
	rc.FeatureToggles[remote.FeatureENS] = true

	callbackCount := atomic.Int32{}
	ch := make(chan any)
	monitor := NewMonitor(ctx, netw, rc, func(_ string) error {
		callbackCount.Add(1)
		ch <- 1
		return nil
	})
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

func TestENSMonitoringEventFiltering(t *testing.T) {
	category.Set(t, category.Unit)

	const serverKey = "server-key"

	tests := []struct {
		name            string
		ensEnabled      bool
		serverKey       string
		event           events.VPNConnectionErrorEvent
		expectReconnect bool
	}{
		{
			name:            "maintenance event for current server triggers reconnect",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerPublicKey: serverKey},
			expectReconnect: true,
		},
		{
			name:            "maintenance event with stale server key is ignored",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerPublicKey: "stale-key"},
			expectReconnect: false,
		},
		{
			name:            "ENS feature disabled ignores maintenance event",
			ensEnabled:      false,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance, ServerPublicKey: serverKey},
			expectReconnect: false,
		},
		{
			name:            "superseded error is ignored",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorSuperseded, ServerPublicKey: serverKey},
			expectReconnect: false,
		},
		{
			name:            "unknown error is ignored",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorUnknown, ServerPublicKey: serverKey},
			expectReconnect: false,
		},
		{
			name:            "connection limit error is ignored",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorConnectionLimitReached, ServerPublicKey: serverKey},
			expectReconnect: false,
		},
		{
			name:            "unauthenticated error is ignored",
			ensEnabled:      true,
			serverKey:       serverKey,
			event:           events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorUnauthenticated, ServerPublicKey: serverKey},
			expectReconnect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netw := &networker.Mock{
				VpnActive:               true,
				ActiveServerDataPresent: true,
				ActiveServerData:        vpn.ServerData{NordLynxPublicKey: tt.serverKey},
			}
			rc := mock.NewRemoteConfigMock()
			rc.FeatureToggles[remote.FeatureENS] = tt.ensEnabled

			callbackCalled := atomic.Bool{}
			ch := make(chan any, 1)
			monitor := NewMonitor(t.Context(), netw, rc, func(_ string) error {
				callbackCalled.Store(true)
				ch <- 1
				return nil
			})
			monitor.Start()

			assert.NilError(t, monitor.HandleENSNotification(tt.event))
			helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
			assert.Equal(t, tt.expectReconnect, callbackCalled.Load())
		})
	}
}
