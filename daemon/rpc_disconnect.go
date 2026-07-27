package daemon

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) Disconnect(_ *pb.Empty, srv pb.Daemon_DisconnectServer) error {
	connectionStateBeforeDisconnect, err := r.DoDisconnect()
	if err != nil {
		log.Error(err)
		return internal.ErrUnhandled
	}
	if connectionStateBeforeDisconnect == pb.ConnectionState_DISCONNECTED {
		return srv.Send(&pb.Payload{
			Type: internal.CodeVPNNotRunning,
		})
	}
	if connectionStateBeforeDisconnect == pb.ConnectionState_PAUSED {
		return srv.Send(&pb.Payload{Type: internal.CodePauseInterrupted})
	}
	return srv.Send(&pb.Payload{Type: internal.CodeDisconnected})
}

// DoDisconnect is the non-gRPC function for Disconnect to be used directly.
func (r *RPC) DoDisconnect() (pb.ConnectionState, error) {
	connectionStateBeforeDisconnect := pb.ConnectionState_DISCONNECTED
	if r.connectionInfo.IsPaused() {
		connectionStateBeforeDisconnect = pb.ConnectionState_PAUSED
		r.CancelPause()
	}
	wasConnected, err := r.doDisconnect(0)
	if wasConnected {
		connectionStateBeforeDisconnect = pb.ConnectionState_CONNECTED
	}

	return connectionStateBeforeDisconnect, err
}

func (r *RPC) DoPause(interval time.Duration) (bool, error) {
	return r.doDisconnect(interval)
}

func (r *RPC) doDisconnect(pauseInterval time.Duration) (bool, error) {
	var recommendationUUID string
	// Not sure if it can be nil in the real scenarios
	if r.connectionInfo != nil {
		recommendationUUID = r.connectionInfo.Status().RecommendationUUID
	} else {
		log.Warn("connection info is nil and it shouldn't be")
	}

	wasConnected, err := access.Disconnect(access.DisconnectInput{
		Networker:     r.netw,
		ConfigManager: r.cm,
		PublishDisconnectEventFunc: func(data events.DataDisconnect) {
			data.PauseInterval = pauseInterval
			r.events.Service.Disconnect.Publish(data)
		},
		RecommendationUUID: recommendationUUID,
	})

	if wasConnected {
		storePendingRecentConnection(r.recentVPNConnStore)
	}

	return wasConnected, err
}
