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

type ConnectCallback func() error

type Monitor struct {
	eventsCh    chan events.VPNConnectionErrorEvent
	ctx         context.Context
	netw        networker.Networker
	reconnectFn ConnectCallback
	rc          remote.ConfigGetter
}

func NewMonitor(ctx context.Context, netw networker.Networker, rc remote.ConfigGetter, connectCallback ConnectCallback) *Monitor {
	return &Monitor{
		eventsCh:    make(chan events.VPNConnectionErrorEvent, evChSize),
		ctx:         ctx,
		netw:        netw,
		rc:          rc,
		reconnectFn: connectCallback,
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
	if m.reconnectFn == nil {
		log.Fatal(logPrefix, "connect callback is nil")
	}
	go m.run()
}

func (m *Monitor) run() {
	log.Info(logPrefix, "start ENS monitoring")

	for {
		select {
		case e, ok := <-m.eventsCh:
			if !ok {
				continue
			}

			if !m.rc.IsFeatureEnabled(remote.FeatureENS) {
				continue
			}

			log.Debug(logPrefix, "event received", e)
			if e.Code != events.VPNConnectionErrorServerMaintenance {
				log.Debug(logPrefix, "ignoring", e)
				continue
			}

			if !m.netw.IsVPNActive() {
				log.Debug(logPrefix, "ignoring because VPN is not connected", e)
				continue
			}
			// TODO: check if current server == event.server and check r.RequestedConnParams
			if err := m.reconnectFn(); err != nil {
				log.Error(logPrefix, "failed to reconnect", err)
			}

		case <-m.ctx.Done():
			log.Info(logPrefix, "stop ENS monitoring")
			return
		}
	}
}
