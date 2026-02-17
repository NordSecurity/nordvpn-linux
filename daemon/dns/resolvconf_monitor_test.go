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

type mockErrorEvent struct {
	errorType errorType
	critical  bool
}

type analyticsMock struct {
	resolvConfEventEmitted atomic.Bool // only resolvConfEventEmitted needs to be stored in an atomic because it's the
	// only field accessed concurrently. resolv.conf monitoring tests need to be run in multiple goroutines because of
	// the nature of this use case.
	managementService   dnsManagementService
	dnsConfiguredEmited bool
	emittedErrors       []mockErrorEvent
}

func (a *analyticsMock) emitResolvConfOverwrittenEvent(dnsManagementService) {
	a.resolvConfEventEmitted.Store(true)
}

func (a *analyticsMock) getResolvConfEmitted() bool {
	return a.resolvConfEventEmitted.Load()
}

func (a *analyticsMock) emitDNSConfiguredEvent(managementService dnsManagementService) {
	a.dnsConfiguredEmited = true
	a.managementService = managementService
}

func (a *analyticsMock) emitDNSConfigurationErrorEvent(managementService dnsManagementService, errorType errorType) {
	a.emittedErrors = append(a.emittedErrors, mockErrorEvent{errorType: errorType, critical: false})
	a.managementService = managementService
}

func (a *analyticsMock) emitDNSConfigurationCriticalErrorEvent(managementService dnsManagementService,
	errorType errorType) {
	a.emittedErrors = append(a.emittedErrors, mockErrorEvent{errorType: errorType, critical: true})
	a.managementService = managementService
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
	eventsChan <- fsnotify.Event{Op: fsnotify.Write}
	checkResultFunc := func() bool {
		return analyticsMock.getResolvConfEmitted()
	}
	revolvConfEventEmitted := checkLoop(checkResultFunc, 10*time.Millisecond, 1*time.Second)

	assert.Equal(t, true, revolvConfEventEmitted, "Event was not emitted after resolv.conf change was detected.")
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
