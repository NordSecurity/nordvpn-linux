// Package subs is responsible for communication between modules and it facilitates dependency decoupling.
package subs

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Subject is a single topic, which supports multiple subscribers.
type Subject[T any] struct {
	subscribers []events.Handler[T]
}

// Subscribe registers a handler for event processing.
func (s *Subject[T]) Subscribe(handler events.Handler[T]) {
	s.subscribers = append(s.subscribers, handler)
}

// Publish message to all registered subscribers.
func (s *Subject[T]) Publish(message T) {
	for _, handler := range s.subscribers {
		if err := handler(message); err != nil {
			log.Printf(
				"%s error while notifying subscriber: %s\n",
				internal.WarningPrefix,
				err)
		}
	}
}
