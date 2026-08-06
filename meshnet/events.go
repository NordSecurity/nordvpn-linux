package meshnet

import (
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
)

// Publisher defines receiver methods for meshnet related notifications
type Publisher interface {
	NotifyPeerUpdate([]string) error
}

// Events allow for publishing and subscribing to meshnet related notifications
type Events struct {
	PeerUpdate  events.PublishSubcriber[[]string]
	SelfRemoved events.PublishSubcriber[any]
}

func NewEventsEmpty() *Events {
	return NewEvents(
		&subs.Subject[[]string]{},
		&subs.Subject[any]{},
	)
}

func NewEvents(
	peerUpdate events.PublishSubcriber[[]string],
	selfRemoved events.PublishSubcriber[any],
) *Events {
	return &Events{PeerUpdate: peerUpdate, SelfRemoved: selfRemoved}
}
