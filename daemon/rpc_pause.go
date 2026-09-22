package daemon

import (
	"context"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// PauseConnection disconnects and schedules a reconnection in a timespan provided in the pause request
func (r *RPC) PauseConnection(ctx context.Context, in *pb.PauseRequest) (*pb.Payload, error) {
	log.RPCPause.Tracef("seconds=%d vpnActive=%v", in.Seconds, r.netw.IsVPNActive())
	if !r.netw.IsVPNActive() {
		log.RPCPause.Trace("vpn not active, nothing to pause")
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	connectionStatus := r.connectionInfo.Status()
	if connectionStatus.State == pb.ConnectionState_PAUSED {
		log.RPCPause.Trace("already paused, nothing to do")
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if r.connectionInfo.Status().IsMeshnetPeer {
		log.RPCPause.Trace("connected to meshnet peer, pause not supported")
		return &pb.Payload{Type: internal.CodePauseAttemptWhenConnectedToMeshPeer}, nil
	}

	if in.Seconds == 0 {
		log.RPCPause.Trace("pause duration is zero, nothing to do")
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	pauseDuration := time.Duration(in.Seconds) * time.Second
	log.RPCPause.Tracef("scheduling reconnection in %v", pauseDuration)
	r.pauseManager.ScheduleReconnection(pauseDuration)

	_, err := r.DoPause(pauseDuration)
	if err != nil {
		r.pauseManager.CancelReconnection()
		log.RPCPause.Error("failed to disconnect when pausing the connection:", err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}

func (r *RPC) CancelPause() {
	r.pauseManager.CancelReconnection()
	//invariant: if the pause gets cancelled, the application state is then disconnected
	//send out status update here, so the fontend can update its state accordingly
	r.events.Service.Disconnect.Publish(events.DataDisconnect{})
}
