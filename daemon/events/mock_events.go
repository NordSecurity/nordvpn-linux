package events

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
)

type MockPublisherSubscriber[T any] struct {
	EventPublished bool
	Event          T
}

func (mp *MockPublisherSubscriber[T]) Publish(message T) {
	mp.EventPublished = true
	mp.Event = message
}
func (*MockPublisherSubscriber[T]) Subscribe(handler events.Handler[T]) {}

func NewMockSettingsEmptyEvents() *SettingsEvents {
	return &SettingsEvents{
		Killswitch:           &MockPublisherSubscriber[bool]{},
		Autoconnect:          &MockPublisherSubscriber[bool]{},
		DNS:                  &MockPublisherSubscriber[events.DataDNS]{},
		ThreatProtectionLite: &MockPublisherSubscriber[bool]{},
		Protocol:             &MockPublisherSubscriber[config.Protocol]{},
		Allowlist:            &MockPublisherSubscriber[events.DataAllowlist]{},
		Technology:           &MockPublisherSubscriber[config.Technology]{},
		Obfuscate:            &MockPublisherSubscriber[bool]{},
		Firewall:             &MockPublisherSubscriber[bool]{},
		Routing:              &MockPublisherSubscriber[bool]{},
		Notify:               &MockPublisherSubscriber[bool]{},
		Meshnet:              &MockPublisherSubscriber[bool]{},
		Ipv6:                 &MockPublisherSubscriber[bool]{},
		Defaults:             &MockPublisherSubscriber[any]{},
		LANDiscovery:         &MockPublisherSubscriber[bool]{},
		VirtualLocation:      &MockPublisherSubscriber[bool]{},
		PostquantumVPN:       &MockPublisherSubscriber[bool]{},
	}
}
