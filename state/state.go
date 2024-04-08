package state

import (
	"log"
	"sync"
)

type Event int

const (
	VPNConnected Event = iota
	VPNDisconnected
)

type StateManager struct {
	mu sync.Mutex
}

func NewState() StateManager {
	return StateManager{}
}

func (s *StateManager) NotifyEvent(e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Received event notification: ", e)

	return nil
}
