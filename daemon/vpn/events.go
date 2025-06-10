package vpn

import (
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
)

type InternalVPNPublisher interface {
	ConnectionStatusNotifyInternalConnect(ConnectEvent) error
	ConnectionStatusNotifyInternalDisconnect(events.TypeEventStatus) error
}

type Events struct {
	Connected    events.PublishSubcriber[ConnectEvent]
	Disconnected events.PublishSubcriber[events.TypeEventStatus]
}

type ConnectEvent struct {
	Status     events.TypeEventStatus
	TunnelName string
}

func NewInternalVPNEvents() *Events {
	return &Events{
		Connected:    &subs.Subject[ConnectEvent]{},
		Disconnected: &subs.Subject[events.TypeEventStatus]{},
	}
}

func (e *Events) Subscribe(to InternalVPNPublisher) {
	e.Connected.Subscribe(to.ConnectionStatusNotifyInternalConnect)
	e.Disconnected.Subscribe(to.ConnectionStatusNotifyInternalDisconnect)
}
