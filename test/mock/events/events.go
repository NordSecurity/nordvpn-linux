package events

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
