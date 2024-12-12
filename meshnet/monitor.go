package meshnet

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/vishvananda/netlink"
)

type (
	PID     uint64
	SetupFn func() (MonitorChannels, error)
)

type MonitorChannels struct {
	EventCh chan netlink.ProcEvent
	DoneCh  chan struct{}
	ErrCh   <-chan error
}

type ProcEvent struct {
	PID PID
}

// EventHandler processes events of process started and stopped.
type EventHandler interface {
	OnProcessStarted(ProcEvent)
	OnProcessStopped(ProcEvent)
}

// NetlinkProcessMonitor monitors EXEC and EXIT events of processes and calls
// [EventHandler.OnProcessStarted] and [EventHandler.OnProcessStopped] accordingly.
type NetlinkProcessMonitor struct {
	handler   EventHandler
	setup     SetupFn
	isRunning atomic.Bool
}

func NewProcMonitor(handler EventHandler, setup SetupFn) NetlinkProcessMonitor {
	return NetlinkProcessMonitor{
		handler: handler,
		setup:   setup,
	}
}

// Start starts listening for process events using Process Events Connector netlink API.
//
// It recreates the source of the events by calling [SetupFn]
// every time [NetlinkProcessMonitor.Start] is called.
func (pm *NetlinkProcessMonitor) Start(ctx context.Context) error {
	if !pm.isRunning.CompareAndSwap(false, true) {
		return errors.New("monitoring already started for this instance")
	}

	var monitoringStarted sync.WaitGroup
	monitoringStarted.Add(1)
	err := pm.start(ctx, &monitoringStarted)
	if err != nil {
		pm.isRunning.CompareAndSwap(true, false)
		return err
	}
	monitoringStarted.Wait()
	return nil
}

func (pm *NetlinkProcessMonitor) start(ctx context.Context, monitoringStarted *sync.WaitGroup) error {
	channels, err := pm.setup()
	if err != nil {
		return err
	}

	go func() {
		monitoringStarted.Done()
		// check if the work was cancelled before the loop even started
		// and mark the monitor as not running
		select {
		case <-ctx.Done():
			pm.isRunning.CompareAndSwap(true, false)
			channels.DoneCh <- struct{}{}
			return
		default:
			// continue to main loop
		}

		for {
			select {
			case ev := <-channels.EventCh:
				pm.handleProcessEvent(&ev)
			case <-ctx.Done():
				pm.isRunning.CompareAndSwap(true, false)
				channels.DoneCh <- struct{}{}
				return
			case err := <-channels.ErrCh:
				log.Println(internal.ErrorPrefix, "error when listening to process events:", err)
			}
		}
	}()

	return nil
}

func (pm *NetlinkProcessMonitor) handleProcessEvent(ev *netlink.ProcEvent) {
	switch ev.What {
	case netlink.PROC_EVENT_EXEC:
		event := ProcEvent{PID: PID(ev.Msg.Pid())}
		pm.handler.OnProcessStarted(event)

	case netlink.PROC_EVENT_EXIT:
		event := ProcEvent{PID: PID(ev.Msg.Pid())}
		pm.handler.OnProcessStopped(event)
	}
}
