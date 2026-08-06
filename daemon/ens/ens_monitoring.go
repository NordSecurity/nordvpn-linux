package ens

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

const evChSize = 2

var ErrConnectionLimitReached = errors.New("connection limit reached")

type ConnectCallback func(serverEndpoint string) error

type Monitor struct {
	eventsCh       chan events.VPNConnectionErrorEvent
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	netw           networker.Networker
	reconnectFn    ConnectCallback
	debuggerEvents events.Publisher[events.DebuggerEvent]
	rc             remote.ConfigGetter
}

func NewMonitor(
	netw networker.Networker,
	rc remote.ConfigGetter,
	connectCallback ConnectCallback,
	debuggerEvents events.Publisher[events.DebuggerEvent],
) *Monitor {
	if connectCallback == nil {
		log.ENS.Fatal("connect callback is nil")
	}
	if debuggerEvents == nil {
		log.ENS.Warn("debugger events publisher is nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Monitor{
		eventsCh:       make(chan events.VPNConnectionErrorEvent, evChSize),
		ctx:            ctx,
		cancel:         cancel,
		netw:           netw,
		rc:             rc,
		reconnectFn:    connectCallback,
		debuggerEvents: debuggerEvents,
	}
}

func (m *Monitor) HandleENSNotification(e events.VPNConnectionErrorEvent) error {
	select {
	case m.eventsCh <- e:
	case <-m.ctx.Done():
		log.ENS.Debug("ignore event because context is done", e)
	case <-time.After(10 * time.Millisecond):
		log.ENS.Warn("channel is full dropping ENS event", e)
	}
	return nil
}

func (m *Monitor) Start() {
	m.wg.Go(m.run)
}

func (m *Monitor) run() {
	log.ENS.Info("start ENS monitoring")

	for {
		select {
		case e, ok := <-m.eventsCh:
			if !ok {
				log.ENS.Warn("events channel closed, stopping ENS monitoring")
				return
			}

			if !m.rc.IsFeatureEnabled(remote.FeatureENS) {
				continue
			}

			log.ENS.Debug("event received", e)
			if m.debuggerEvents != nil {
				m.debuggerEvents.Publish(*newVPNConnectionErrorEvent(e.Code).ToDebuggerEvent())
			}

			//exhaustive:ignore
			switch e.Code {
			case events.VPNConnectionErrorServerMaintenance:
				m.serverMaintenanceEventProcessing(e)
			case events.VPNConnectionErrorConnectionLimitReached:
				m.connectionLimitReachedEventProcessing(e)
			default:
				log.ENS.Debug("ignoring", e)
			}

		case <-m.ctx.Done():
			log.ENS.Info("stop ENS monitoring")
			return
		}
	}
}

func (m *Monitor) Stop() {
	log.ENS.Info("stopping ENS monitoring")
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
}

func (m *Monitor) serverMaintenanceEventProcessing(e events.VPNConnectionErrorEvent) {
	if e.Code != events.VPNConnectionErrorServerMaintenance {
		return
	}

	if !m.netw.IsVPNActive() {
		log.ENS.Debug("ignoring because VPN is not connected", e)
		return
	}

	currServer, _ := m.netw.GetConnectionParameters()
	eventIsForDifferentServer := !currServer.EndpointEqual(e.ServerEndpoint)
	if eventIsForDifferentServer {
		log.ENS.Debug("ignoring ENS event for non-current server", e)
		return
	}

	if err := m.reconnectFn(e.ServerEndpoint); err != nil {
		log.ENS.Error("failed to reconnect", err)
	}
}

func (m *Monitor) connectionLimitReachedEventProcessing(e events.VPNConnectionErrorEvent) {
	if e.Code != events.VPNConnectionErrorConnectionLimitReached {
		return
	}
	log.ENS.Debug("connection limit reach received")

	if !m.netw.CancelConnecting(ErrConnectionLimitReached) {
		log.ENS.Info("connection limit reach ignored")
	}
}
