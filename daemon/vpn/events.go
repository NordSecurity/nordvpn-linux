package vpn

import (
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
)

type InternalVPNPublisher interface {
	NotifyConnect(events.DataConnect) error
	NotifyDisconnect(events.DataDisconnect) error
}

type Events struct {
	Connected    events.PublishSubcriber[events.DataConnect]
	Disconnected events.PublishSubcriber[events.DataDisconnect]
}

func NewInternalVPNEvents() *Events {
	return &Events{
		Connected:    &subs.Subject[events.DataConnect]{},
		Disconnected: &subs.Subject[events.DataDisconnect]{},
	}
}

func (e *Events) Subscribe(to InternalVPNPublisher) {
	e.Connected.Subscribe(to.NotifyConnect)
	e.Disconnected.Subscribe(to.NotifyDisconnect)
}
