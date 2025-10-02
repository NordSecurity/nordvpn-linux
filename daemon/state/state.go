package state

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const notifyTimeout = 100 * time.Millisecond

type subscriber struct {
	stateChan chan<- any
	stopChan  <-chan struct{}
}

func newSubscriber(stateChan chan<- any, stopChan <-chan struct{}) subscriber {
	return subscriber{
		stateChan: stateChan,
		stopChan:  stopChan,
	}
}

type StatePublisher struct {
	subscribers []subscriber
	mu          sync.Mutex
}

func NewState() *StatePublisher {
	return &StatePublisher{}
}

func (s *StatePublisher) notify(e any) {
	newSubs := []subscriber{}
	for _, sub := range s.subscribers {
		timeout, cancel := context.WithTimeout(context.Background(), notifyTimeout)
		defer cancel()
		select {
		case <-sub.stopChan:
			close(sub.stateChan)
		case sub.stateChan <- e:
			newSubs = append(newSubs, sub)
		case <-timeout.Done():
			newSubs = append(newSubs, sub)
			log.Println(internal.WarningPrefix, "could not notify state subscriber, event dropped")
		}
	}
	s.subscribers = newSubs
}

func (s *StatePublisher) NotifyChangeState(e events.DataConnectChangeNotif) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf(internal.InfoPrefix+" notifying about data connect change event: %+v", e)
	s.notify(e)
	return nil
}

func (s *StatePublisher) NotifyRecentsChanged(e events.DataRecentsChanged) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.InfoPrefix, "notifying about recents change")
	s.notify(pb.UpdateEvent_RECENTS_LIST_UPDATE)
	return nil
}

func (s *StatePublisher) notifyLoginLogout(status events.TypeEventStatus, eventType pb.LoginEventType) {
	// skip any event types other than success, as subscribers(GUI) do not care about them
	if status != events.StatusSuccess {
		return
	}

	s.notify(eventType)
}

func (s *StatePublisher) NotifyLogin(e events.DataAuthorization) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.InfoPrefix, "notifying about login event")
	s.notifyLoginLogout(e.EventStatus, pb.LoginEventType_LOGIN)

	return nil
}

func (s *StatePublisher) NotifyLogout(e events.DataAuthorization) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.InfoPrefix, "notifying about logout event")
	s.notifyLoginLogout(e.EventStatus, pb.LoginEventType_LOGOUT)
	return nil
}

func (s *StatePublisher) NotifyMFA(bool) error {
	return nil
}

func (s *StatePublisher) NotifyConfigChanged(e config.DataConfigChange) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.InfoPrefix, "notifying about config change:", e.Caller)

	s.notify(e.Config)

	return nil
}

func (s *StatePublisher) NotifyServersListUpdate(any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.InfoPrefix, "notifying about servers list update")
	s.notify(pb.UpdateEvent_SERVERS_LIST_UPDATE)

	return nil
}

func (s *StatePublisher) NotifySubscriptionChanged(e *pb.AccountModification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println(internal.InfoPrefix, "notifying about subscription update")
	s.notify(e)

	return nil
}

func (s *StatePublisher) NotifyVersionHealth(healthStatus *pb.VersionHealthStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf(internal.InfoPrefix+"notifying about version health change: status_code=%d\n", healthStatus.StatusCode)
	s.notify(healthStatus)

	return nil
}

func (s *StatePublisher) AddSubscriber() (<-chan any, chan<- struct{}) {
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

	stateChan := make(chan any)
	stopChan := make(chan struct{})

	newSubs = append(newSubs, newSubscriber(stateChan, stopChan))
	s.subscribers = newSubs

	return stateChan, stopChan
}
