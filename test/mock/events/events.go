package events

import "github.com/NordSecurity/nordvpn-linux/events"

type MockPublisher[T any] struct {
	publishedEvents []T
}

// PopEvent returns last received event and number of remaining events. Returns false if event stack is empty.
func (m *MockPublisher[T]) PopEvent() (T, int, bool) {
	if len(m.publishedEvents) == 0 {
		var empty T
		return empty, 0, false
	}

	event := m.publishedEvents[len(m.publishedEvents)-1:]
	m.publishedEvents = m.publishedEvents[:len(m.publishedEvents)-1]
	return event[0], len(m.publishedEvents), true
}

func (m *MockPublisher[T]) Publish(message T) {
	m.publishedEvents = append(m.publishedEvents, message)
}

func NewMockPublisherSubscriber[T any]() *MockPublisherSubscriber[T] {
	return &MockPublisherSubscriber[T]{}
}

type MockPublisherSubscriber[T any] struct {
	Handler        events.Handler[T]
	EventPublished bool
	Event          T
}

func (mp *MockPublisherSubscriber[T]) Publish(message T) {
	mp.EventPublished = true
	mp.Event = message
	if mp.Handler != nil {
		_ = mp.Handler(message)
	}
}

func (mp *MockPublisherSubscriber[T]) Subscribe(handler events.Handler[T]) {
	mp.Handler = handler
}
