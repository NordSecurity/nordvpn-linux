package daemon

import (
	"context"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/log"
)

type ReconnectScheduler interface {
	ScheduleReconnection(duration time.Duration)
	CancelReconnection() time.Duration
}

type connectFunc func(srv pb.Daemon_ConnectServer, source pb.ConnectionSource, pauseDuration time.Duration) error

type ReconnectSchedulerImpl struct {
	mu                  sync.Mutex
	reconnectCancelFunc context.CancelFunc
	// reconnectionScheduledChan will be closed once reconnection is cancelled or after the reconection wait period
	// finishes
	reconnectionScheduledChan <-chan any
	connectFunc               connectFunc
	connectionInfo            *state.ConnectionInfo
}

func NewReconnectScheduler(connectFunc connectFunc, connectionInfo *state.ConnectionInfo) ReconnectScheduler {
	return &ReconnectSchedulerImpl{
		connectFunc:    connectFunc,
		connectionInfo: connectionInfo,
	}
}

// ScheduleReconnection schedules a reconnection. Reconnection attempt will be made in a duration of time
func (s *ReconnectSchedulerImpl) ScheduleReconnection(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reconnectCancelFunc != nil {
		log.Debug("cancelling previous reconnection before initiating a new one")
		s.reconnectCancelFunc()
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	s.reconnectCancelFunc = cancelFunc
	reconnectionScheduledChan := make(chan any)
	s.reconnectionScheduledChan = reconnectionScheduledChan
	s.connectionInfo.Pause(time.Now(), duration)
	go func() {
		defer close(reconnectionScheduledChan)
		log.Debug("pausing connection for", duration.String())
		select {
		case <-time.After(duration):
			log.Debug("resuming connection after a pause")
			pauseDuration := s.connectionInfo.Unpause()

			connServer := connectServer{}
			err := s.connectFunc(&connServer, pb.ConnectionSource_AUTO, pauseDuration)
			if err != nil || connServer.err != nil {
				log.Error(
					"failed to reconnect after a pause: connection error:", err,
					"server error:", connServer.err,
				)
			}
		case <-ctx.Done():
			return
		}
	}()
}

func (s *ReconnectSchedulerImpl) isReconnectionScheduled() bool {
	if s.reconnectionScheduledChan == nil {
		return false
	}

	select {
	case <-s.reconnectionScheduledChan:
		return false
	default:
		return true
	}
}

// CancelReconnection cancels the reconnect goroutine if it was started.
func (s *ReconnectSchedulerImpl) CancelReconnection() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reconnectCancelFunc != nil && s.isReconnectionScheduled() {
		log.Debug("cancelling the reconnection after a pause")
		s.reconnectCancelFunc()
		s.reconnectCancelFunc = nil
		return s.connectionInfo.CancelPause()
	}

	return 0
}
