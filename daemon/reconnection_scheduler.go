package daemon

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type ReconnectScheduler interface {
	ScheduleReconnection(duration time.Duration)
	CancelReconnection()
}

type connectFunc func(srv pb.Daemon_ConnectServer, source pb.ConnectionSource) error

type Scheduler struct {
	mu              sync.Mutex
	pauseCancelFunc context.CancelFunc
	connectFunc     connectFunc
	connectionInfo  *state.ConnectionInfo
}

func NewPauseManager(connectFunc connectFunc, connectionInfo *state.ConnectionInfo) Scheduler {
	return Scheduler{
		connectFunc:    connectFunc,
		connectionInfo: connectionInfo,
	}
}

// ScheduleReconnection schedules a reconnection. Reconnection attempt will be made in a duration of time
func (pm *Scheduler) ScheduleReconnection(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	ctx, cancelFunc := context.WithCancel(context.Background())
	pm.pauseCancelFunc = cancelFunc
	pm.connectionInfo.Pause(time.Now(), duration)
	go func() {
		log.Println(internal.DebugPrefix, "pausing connection for", duration.String())
		select {
		case <-time.After(duration):
			log.Println(internal.DebugPrefix, "resuming connection after a pause")
			pm.connectionInfo.Unpause()

			connServer := connectServer{}
			err := pm.connectFunc(&connServer, pb.ConnectionSource_AUTO)
			if err != nil || connServer.err != nil {
				log.Println(internal.ErrorPrefix,
					"failed to reconnect after a pause: connection error:", err, "server error:", connServer.err)
			}
		case <-ctx.Done():
			return
		}
	}()
}

// CancelReconnection cancels the reconnect goroutine if it was started
func (pm *Scheduler) CancelReconnection() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.pauseCancelFunc != nil {
		log.Println(internal.DebugPrefix, "cancelling the reconnection after a pause")
		pm.connectionInfo.Unpause()
		pm.pauseCancelFunc()
	}
}
