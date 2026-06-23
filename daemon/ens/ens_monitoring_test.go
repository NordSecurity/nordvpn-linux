package ens

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"gotest.tools/v3/assert"
)

func TestENSMonitoring(t *testing.T) {
	category.Set(t, category.Unit)

	ctx, cancelFn := context.WithCancel(context.Background())
	netw := &networker.Mock{
		VpnActive: true,
	}
	rc := mock.NewRemoteConfigMock()
	rc.FeatureToggles[remote.FeatureENS] = true

	callbackCount := atomic.Int32{}
	ch := make(chan any)
	monitor := NewMonitor(
		ctx,
		netw,
		rc,
		func() error {
			callbackCount.Add(1)
			ch <- 1
			return nil
		},
	)

	monitor.Start()

	monitor.HandleENSNotification(events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance})

	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 1, int(callbackCount.Load()))

	// events except VPNConnectionErrorServerMaintenance are ignored
	for _, ev := range []events.VPNConnectionError{
		events.VPNConnectionErrorSuperseded,
		events.VPNConnectionErrorUnknown,
		events.VPNConnectionErrorConnectionLimitReached,
		events.VPNConnectionErrorUnauthenticated,
		events.VPNConnectionErrorSuperseded,
	} {
		monitor.HandleENSNotification(events.VPNConnectionErrorEvent{Code: ev})
		helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
		assert.Equal(t, 1, int(callbackCount.Load()))
	}

	monitor.HandleENSNotification(events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance})
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 2, int(callbackCount.Load()))

	cancelFn()

	// after stopping the monitoring, events are ignored
	monitor.HandleENSNotification(events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance})
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 2, int(callbackCount.Load()))
}

func TestENSMonitoringWhenENSIsDisabledFromRC(t *testing.T) {
	category.Set(t, category.Unit)

	netw := &networker.Mock{
		VpnActive: true,
	}
	rc := mock.NewRemoteConfigMock()
	rc.FeatureToggles[remote.FeatureENS] = false
	callbackCount := atomic.Int32{}
	ch := make(chan any)
	monitor := NewMonitor(
		context.Background(),
		netw,
		rc,
		func() error {
			callbackCount.Add(1)
			ch <- 1
			return nil
		},
	)

	monitor.Start()

	monitor.HandleENSNotification(events.VPNConnectionErrorEvent{Code: events.VPNConnectionErrorServerMaintenance})
	helpers.WaitWithTimeout(t, ch, time.Millisecond*10)
	assert.Equal(t, 0, int(callbackCount.Load()))
}
