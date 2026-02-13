package dns

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

type analyticsMock struct {
	resolvConfEventEmitted atomic.Bool
	dnsConfiguredEmited    bool
	managementService      dnsManagementService
}

func (a *analyticsMock) setManagementService(managementService dnsManagementService) {
	a.managementService = managementService
}

func (a *analyticsMock) emitResolvConfOverwrittenEvent() {
	a.resolvConfEventEmitted.Store(true)
}

func (a *analyticsMock) getResolvConfEmitted() bool {
	return a.resolvConfEventEmitted.Load()
}

func (a *analyticsMock) emitDNSConfiguredEvent() {
	a.dnsConfiguredEmited = true
}

func newAnalyticsMock() analyticsMock {
	return analyticsMock{}
}

// checkLoop executes test in an interval until it returns true or a timeout is reached
func checkLoop(test func() bool, interval time.Duration, timeout time.Duration) bool {
	if test() {
		return true
	}
	resultChan := make(chan bool)
	ctx := context.Background()

	go func() {
		for {
			select {
			case <-time.After(interval):
				if test() {
					resultChan <- true
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-time.After(timeout):
			ctx.Done()
			return false
		case <-resultChan:
			return true
		}
	}
}

func Test_ResolvConfMonitoring(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		event      fsnotify.Op
		shouldEmit bool
	}{
		{
			name:       "Write event emits analytics",
			event:      fsnotify.Write,
			shouldEmit: true,
		},
		{
			name:       "Remove event emits analytics",
			event:      fsnotify.Remove,
			shouldEmit: true,
		},
		{
			name:       "Create event does not emit analytics",
			event:      fsnotify.Create,
			shouldEmit: false,
		},
		{
			name:       "Rename event does not emit analytics",
			event:      fsnotify.Rename,
			shouldEmit: false,
		},
		{
			name:       "Chmod event does not emit analytics",
			event:      fsnotify.Chmod,
			shouldEmit: false,
		},
		{
			name:       "Combined Chmod and Write event does emit analytics",
			event:      fsnotify.Chmod | fsnotify.Write,
			shouldEmit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventsChan := make(chan fsnotify.Event)
			errorChan := make(chan error)
			getMockWatcherFunc := func(...string) (*fsnotify.Watcher, error) {
				watcher, _ := fsnotify.NewWatcher()
				watcher.Events = eventsChan
				watcher.Errors = errorChan
				return watcher, nil
			}

			analyticsMock := newAnalyticsMock()

			resolvConfMonitor := resolvConfFileWatcherMonitor{
				analytics:      &analyticsMock,
				getWatcherFunc: getMockWatcherFunc,
			}

			resolvConfMonitor.start()
			eventsChan <- fsnotify.Event{Op: tt.event}
			checkResultFunc := func() bool {
				return analyticsMock.getResolvConfEmitted()
			}
			revolvConfEventEmitted := checkLoop(checkResultFunc, 10*time.Millisecond, 500*time.Millisecond)

			assert.Equal(t, tt.shouldEmit, revolvConfEventEmitted, "Event emission did not match expected behavior.")
		})
	}
}

func Test_ResolvConfMonitoringDoesNotDeadlock(t *testing.T) {
	category.Set(t, category.Unit)

	eventsChan := make(chan fsnotify.Event)
	errorChan := make(chan error)
	getMockWatcherFunc := func(...string) (*fsnotify.Watcher, error) {
		watcher, _ := fsnotify.NewWatcher()
		watcher.Events = eventsChan
		watcher.Errors = errorChan
		time.Sleep(time.Duration(time.Duration.Seconds(1)))
		return watcher, nil
	}

	analyticsMock := newAnalyticsMock()

	resolvConfMonitor := resolvConfFileWatcherMonitor{
		analytics:      &analyticsMock,
		getWatcherFunc: getMockWatcherFunc,
	}

	resolvConfMonitor.start()

	stoppedChan := make(chan any)
	go func() {
		resolvConfMonitor.stop()
		stoppedChan <- true
	}()

	select {
	case <-stoppedChan:
	case <-time.After(time.Second * 1):
		assert.Fail(t, "Timed out waiting for the resolvConf monitor to stop")
	}
}
