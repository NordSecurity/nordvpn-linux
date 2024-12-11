package meshnet

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
	"golang.org/x/exp/rand"
	"golang.org/x/net/context"
)

func TestNetlinkProcessMonitor_Start(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		setupFn SetupFn
		isError bool
	}{
		{
			name:    "fails when the setup function returns error",
			setupFn: failingSetup,
			isError: true,
		},
		{
			name:    "succeeds",
			setupFn: workingSetup,
			isError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProcMonitor(eventHandlerDummy{}, tt.setupFn)
			ctx, _ := context.WithCancel(context.Background())
			err := monitor.Start(ctx)

			assert.Equal(t, err != nil, tt.isError)
		})
	}
}

func TestNetlinkProcessMonitor_EventHandler(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		eh            *eventHandlerSpy
		correctEvents []eventType
		wrongEvents   []eventType
		signalChan    func(*eventHandlerSpy) chan ProcEvent
		callCount     func(*eventHandlerSpy) int
	}{
		{
			name:          "OnProcessStarted is called only for EXEC event",
			eh:            newEventHandlerSpy(),
			correctEvents: []eventType{netlink.PROC_EVENT_EXEC},
			wrongEvents: []eventType{
				netlink.PROC_EVENT_NONE,
				netlink.PROC_EVENT_FORK,
				netlink.PROC_EVENT_UID,
				netlink.PROC_EVENT_GID,
				netlink.PROC_EVENT_SID,
				netlink.PROC_EVENT_PTRACE,
				netlink.PROC_EVENT_COMM,
				netlink.PROC_EVENT_COREDUMP,
				netlink.PROC_EVENT_EXIT,
			},
			signalChan: func(eh *eventHandlerSpy) chan ProcEvent {
				return eh.startedCalledSignal
			},
			callCount: func(eh *eventHandlerSpy) int {
				return eh.onStartedCallCount
			},
		},
		{
			name:          "OnProcessStopped is called only for EXIT event",
			eh:            newEventHandlerSpy(),
			correctEvents: []eventType{netlink.PROC_EVENT_EXIT},
			wrongEvents: []eventType{
				netlink.PROC_EVENT_NONE,
				netlink.PROC_EVENT_FORK,
				netlink.PROC_EVENT_UID,
				netlink.PROC_EVENT_GID,
				netlink.PROC_EVENT_SID,
				netlink.PROC_EVENT_PTRACE,
				netlink.PROC_EVENT_COMM,
				netlink.PROC_EVENT_COREDUMP,
				netlink.PROC_EVENT_EXEC,
			},
			signalChan: func(eh *eventHandlerSpy) chan ProcEvent {
				return eh.stoppedCalledSignal
			},
			callCount: func(eh *eventHandlerSpy) int {
				return eh.onStoppedCallCount
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channels, setupFn := openChannelsMonitorSetup()
			monitor := NewProcMonitor(tt.eh, setupFn)
			ctx, _ := context.WithCancel(context.Background())
			monitor.Start(ctx)
			assert.Zero(t, tt.eh.onStartedCallCount)
			assert.Zero(t, len(tt.eh.startedCalledSignal))

			// send all correct + wrong events
			for _, eventType := range interleaveRandomly(tt.correctEvents, tt.wrongEvents) {
				channels.EventCh <- mkEvent(eventType, 1337)
			}

			// receive only correct events
			for range tt.correctEvents {
				ev := <-tt.signalChan(tt.eh)
				assert.Equal(t, ev.PID, PID(1337))
			}

			// no more events in the channel
			assert.Zero(t, len(tt.signalChan(tt.eh)))
			assert.Equal(t, tt.callCount(tt.eh), len(tt.correctEvents))
		})
	}
}

func TestMonitorCancellation(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		setupFn SetupFn
		isError bool
	}{
		{
			name:    "stops monitoring",
			setupFn: workingSetup,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channels, setupFn := openChannelsMonitorSetup()
			spy := newEventHandlerSpy()
			monitor := NewProcMonitor(spy, setupFn)
			ctx, cancel := context.WithCancel(context.Background())
			monitor.Start(ctx)
			// event handler is running
			channels.EventCh <- mkEvent(netlink.PROC_EVENT_EXEC, 1337)
			<-spy.startedCalledSignal

			cancel()

			select {
			case <-channels.DoneCh:
				// monitor was stopped
			case <-time.After(500 * time.Millisecond):
				t.Fatal("monitor stopper failed to stop the monitor")
			}

			// calling cancel multiple times has no effect
			cancel()
		})
	}
}

// event handler dummy
type eventHandlerDummy struct{}

func (eventHandlerDummy) OnProcessStarted(ProcEvent) {}
func (eventHandlerDummy) OnProcessStopped(ProcEvent) {}

// event handler spy
type eventHandlerSpy struct {
	onStartedCallCount  int
	startedCalledSignal chan ProcEvent

	onStoppedCallCount  int
	stoppedCalledSignal chan ProcEvent
}

func newEventHandlerSpy() *eventHandlerSpy {
	return &eventHandlerSpy{
		startedCalledSignal: make(chan ProcEvent, 16),
		stoppedCalledSignal: make(chan ProcEvent, 16),
	}
}

func (eh *eventHandlerSpy) OnProcessStarted(ev ProcEvent) {
	eh.onStartedCallCount += 1
	eh.startedCalledSignal <- ev
}

func (eh *eventHandlerSpy) OnProcessStopped(ev ProcEvent) {
	eh.onStoppedCallCount += 1
	eh.stoppedCalledSignal <- ev
}

func failingSetup() (MonitorChannels, error) {
	return MonitorChannels{}, errors.New("test error")
}

func workingSetup() (MonitorChannels, error) {
	return MonitorChannels{
		EventCh: make(chan netlink.ProcEvent),
		DoneCh:  make(chan struct{}),
		ErrCh:   make(chan error),
	}, nil
}

// msg dummy
type msgStub struct {
	PID uint32
}

func (m msgStub) Pid() uint32 {
	return m.PID
}

func (msgStub) Tgid() uint32 {
	return 0 // not important
}

type eventType = uint32

func openChannelsMonitorSetup() (MonitorChannels, SetupFn) {
	channels := MonitorChannels{
		EventCh: make(chan netlink.ProcEvent),
		DoneCh:  make(chan struct{}),
		ErrCh:   make(chan error),
	}

	return channels, func() (MonitorChannels, error) {
		return channels, nil
	}
}

func interleaveRandomly(arr1, arr2 []uint32) []uint32 {
	combined := append(arr1, arr2...)

	rand.Seed(uint64(time.Now().UnixNano()))

	rand.Shuffle(len(combined), func(i, j int) {
		combined[i], combined[j] = combined[j], combined[i]
	})

	return combined
}

func mkEvent(what uint32, PID uint32) netlink.ProcEvent {
	return netlink.ProcEvent{
		ProcEventHeader: netlink.ProcEventHeader{
			What: what,
		},
		Msg: msgStub{
			PID: PID,
		},
	}
}
