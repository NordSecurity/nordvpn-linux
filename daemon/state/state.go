package state

import (
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
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

func NewState() *StatePublisher {
	return &StatePublisher{}
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

	log.Printf(internal.DebugPrefix+" notifying about connect event: %+v", e)
	s.notify(e)

	return nil
}

func (s *StatePublisher) NotifyDisconnect(e events.DataDisconnect) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf(internal.DebugPrefix+" notifying about disconnect event: %+v", e)
	s.notify(e)

	return nil
}

func (s *StatePublisher) NotifyLogin(e events.DataAuthorization) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.DebugPrefix, "notifying about login event")
	s.notify(pb.LoginEventType_LOGIN)

	return nil
}

func (s *StatePublisher) NotifyLogout(e events.DataAuthorization) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.DebugPrefix, "notifying about logout event")
	s.notify(pb.LoginEventType_LOGOUT)

	return nil
}

func (s *StatePublisher) NotifyMFA(bool) error {
	return nil
}

func (s *StatePublisher) NotifyConfigChanged(e *config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.DebugPrefix, "notifying about config change")
	s.notify(e)

	return nil
}

func (s *StatePublisher) NotifyServersListUpdate(any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.DebugPrefix, "notifying about servers list update")
	s.notify(pb.UpdateEvent_SERVERS_LIST_UPDATE)

	return nil
}

func (s *StatePublisher) AddSubscriber() (<-chan interface{}, chan<- struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newSubs := []subscriber{}
	for _, sub := range s.subscribers {
		select {
		case <-sub.stopChan:
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
