package ens

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	netw "github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	firewallmock "github.com/NordSecurity/nordvpn-linux/test/mock/firewall"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"gotest.tools/v3/assert"
)

type callCounter struct {
	counter atomic.Int32
	ch      chan int32
}

func newCallCounter() *callCounter {
	return &callCounter{
		counter: atomic.Int32{},
		ch:      make(chan int32),
	}
}

func (c *callCounter) add(delta int32) {
	c.ch <- c.counter.Add(delta)
}

func (c *callCounter) waitAndCmp(t *testing.T, expected int) {
	t.Helper()
	helpers.WaitWithTimeout[int32](t, c.ch, time.Millisecond*10)
	assert.Equal(t, expected, int(c.counter.Load()))
}

func TestENSMonitoring(t *testing.T) {
	category.Set(t, category.Unit)

	const serverEndpoint = "192.168.1.1:51820"

	connectCallbackCounter := newCallCounter()
	cancelConnCounter := newCallCounter()

	netw := &networker.Mock{
		VpnActive:        true,
		ActiveServerData: &vpn.ServerData{Endpoint: serverEndpoint},
		CancelConnectingFn: func(err error) bool {
			assert.ErrorIs(t, err, ErrConnectionLimitReached)
			cancelConnCounter.add(1)
			return true
		},
	}
	rc := mock.NewRemoteConfigMock()
	rc.FeatureToggles[remote.FeatureENS] = true

	monitor := NewMonitor(netw, rc, func(_ string) error {
		connectCallbackCounter.add(1)
		return nil
	}, &subs.Subject[events.DebuggerEvent]{})
	monitor.Start()

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorServerMaintenance,
		ServerEndpoint: serverEndpoint,
	}))

	connectCallbackCounter.waitAndCmp(t, 1)
	cancelConnCounter.waitAndCmp(t, 0)

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorServerMaintenance,
		ServerEndpoint: serverEndpoint,
	}))
	connectCallbackCounter.waitAndCmp(t, 2)

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorConnectionLimitReached,
		ServerEndpoint: serverEndpoint,
	}))
	connectCallbackCounter.waitAndCmp(t, 2)
	cancelConnCounter.waitAndCmp(t, 1)

	monitor.Stop()

	// after stopping the monitoring, events are ignored
	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorServerMaintenance,
		ServerEndpoint: serverEndpoint,
	}))

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorConnectionLimitReached,
		ServerEndpoint: serverEndpoint,
	}))
	connectCallbackCounter.waitAndCmp(t, 2)
	cancelConnCounter.waitAndCmp(t, 1)
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
			monitor := NewMonitor(netw, rc,
				func(_ string) error {
					reconnected <- true
					return nil
				},
				debuggerEvents,
			)
			monitor.Start()
			defer monitor.Stop()

			assert.NilError(t, monitor.HandleENSNotification(tt.event))
			assert.Equal(t, tt.expectReport, helpers.WaitWithTimeout(t, reported, time.Millisecond*50))
			assert.Equal(t, tt.expectReconnect, helpers.WaitWithTimeout(t, reconnected, time.Millisecond*50))
		})
	}
}

func TestCombined_ENSConnectionsLimitReached(t *testing.T) {
	netw := netw.NewCombined(
		&mock.WorkingVPN{
			ConnectingFn: func(ctx context.Context, c vpn.Credentials, sd vpn.ServerData) error {
				<-ctx.Done()
				return context.Cause(ctx)
			},
		},
		nil,
		nil,
		&subs.Subject[string]{},
		nil,
		nil,
		firewallmock.NewFirewall(),
		nil,
		&mock.PolicyRouter{},
		nil,
		mock.Router{},
		nil,
		0,
		false,
		&mock.IpV6Blocker{},
		false,
		&mock.SysctlSetterMock{},
		config.Allowlist{},
		&mock.SysctlSetterMock{},
	)

	rc := mock.NewRemoteConfigMock()
	rc.FeatureToggles[remote.FeatureENS] = true
	monitor := NewMonitor(netw, rc, func(_ string) error {
		return nil
	}, &subs.Subject[events.DebuggerEvent]{})
	monitor.Start()

	assert.NilError(t, monitor.HandleENSNotification(events.VPNConnectionErrorEvent{
		Code:           events.VPNConnectionErrorConnectionLimitReached,
		ServerEndpoint: "",
	}))

	err := netw.Start(
		context.Background(),
		vpn.Credentials{},
		vpn.ServerData{},
		config.NewAllowlist(nil, nil, nil),
		[]string{"1.1.1.1"},
		true,
		func(startTime time.Time, err error) {
		},
	)

	assert.ErrorIs(t, err, ErrConnectionLimitReached)
}
