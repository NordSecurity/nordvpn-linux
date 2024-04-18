package state

import (
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/events"
)

type Event int

const (
	VPNConnected Event = iota
	VPNDisconnected
)

type subscriber struct {
	stateChan chan<- interface{}
	stopChan  <-chan struct{}
}

func newSubscriber(stateChan chan<- interface{}, stopChan <-chan struct{}) subscriber {
	return subscriber{
		stateChan: stateChan,
		stopChan:  stopChan,
	}
}

type StatePublisher struct {
	mu          sync.Mutex
	subscribers []subscriber
}

func NewState() StatePublisher {
	return StatePublisher{}
}

func (s *StatePublisher) notify(e interface{}) {
	newSubs := []subscriber{}
	for _, sub := range s.subscribers {
		select {
		case <-sub.stopChan:
			close(sub.stateChan)
		case sub.stateChan <- e:
			newSubs = append(newSubs, sub)
		default:
			newSubs = append(newSubs, sub)
		}
	}
	s.subscribers = newSubs
}

func (s *StatePublisher) NotifyConnect(e events.DataConnect) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notify(e)

	return nil
}

func (s *StatePublisher) NotifyDisconnect(e events.DataDisconnect) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notify(e)

	return nil
}

func (s *StatePublisher) AddSubscriber() (<-chan interface{}, chan<- struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newSubs := []subscriber{}
	for _, sub := range s.subscribers {
		select {
		case <-sub.stopChan:
			log.Println("[DEBUG] remove subscriber")
			close(sub.stateChan)
		default:
			newSubs = append(newSubs, sub)
		}
	}

	stateChan := make(chan interface{})
	stopChan := make(chan struct{})

	newSubs = append(newSubs, newSubscriber(stateChan, stopChan))
	s.subscribers = newSubs

	return stateChan, stopChan
}
