package ens

import (
	"context"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

const (
	logPrefix = "[ens]"
	evChSize  = 2
)

type ConnectCallback func(serverPublicKey string) error

// ErrorReporter publishes a VPN connection error code for telemetry.
type ErrorReporter func(code events.VPNConnectionError)

type Monitor struct {
	eventsCh    chan events.VPNConnectionErrorEvent
	ctx         context.Context
	netw        networker.Networker
	reconnectFn ConnectCallback
	reportFn    ErrorReporter
	rc          remote.ConfigGetter
}

func NewMonitor(
	ctx context.Context,
	netw networker.Networker,
	rc remote.ConfigGetter,
	connectCallback ConnectCallback,
	reportCallback ErrorReporter,
) *Monitor {
	if connectCallback == nil {
		log.Fatal(logPrefix, "connect callback is nil")
	}
	if reportCallback == nil {
		log.Fatal(logPrefix, "report callback is nil")
	}
	return &Monitor{
		eventsCh:    make(chan events.VPNConnectionErrorEvent, evChSize),
		ctx:         ctx,
		netw:        netw,
		rc:          rc,
		reconnectFn: connectCallback,
		reportFn:    reportCallback,
	}
}

func (m *Monitor) HandleENSNotification(e events.VPNConnectionErrorEvent) error {
	select {
	case m.eventsCh <- e:
	case <-m.ctx.Done():
		log.Debug(logPrefix, "ignore event because context is done", e)
	case <-time.After(10 * time.Millisecond):
		log.Warn(logPrefix, "channel is full dropping ENS event", e)
	}
	return nil
}

func (m *Monitor) Start() {
	go m.run()
}

func (m *Monitor) run() {
	log.Info(logPrefix, "start ENS monitoring")

	for {
		select {
		case e, ok := <-m.eventsCh:
			if !ok {
				log.Warn(logPrefix, "events channel closed, stopping ENS monitoring")
				return
			}

			if !m.rc.IsFeatureEnabled(remote.FeatureENS) {
				continue
			}

			log.Debug(logPrefix, "event received", e)
			m.reportFn(e.Code)

			if e.Code != events.VPNConnectionErrorServerMaintenance {
				log.Debug(logPrefix, "ignoring", e)
				continue
			}

			if !m.netw.IsVPNActive() {
				log.Debug(logPrefix, "ignoring because VPN is not connected", e)
				continue
			}

			currServer, _ := m.netw.GetConnectionParameters()
			eventIsForDifferentServer := currServer.NordLynxPublicKey != e.ServerPublicKey
			if eventIsForDifferentServer {
				log.Debug(logPrefix, "ignoring ENS event for non-current server", e)
				continue
			}

			if err := m.reconnectFn(e.ServerPublicKey); err != nil {
				log.Error(logPrefix, "failed to reconnect", err)
			}

		case <-m.ctx.Done():
			log.Info(logPrefix, "stop ENS monitoring")
			return
		}
	}
}
