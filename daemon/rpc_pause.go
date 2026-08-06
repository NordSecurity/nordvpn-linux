package daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func pauseIntervalToDuration(interval pb.PauseInverval) (time.Duration, error) {
	switch interval {
	case pb.PauseInverval_PAUSE_5_MIN:
		return 5 * time.Minute, nil
	case pb.PauseInverval_PAUSE_15_MIN:
		return 15 * time.Minute, nil
	case pb.PauseInverval_PAUSE_30_MIN:
		return 30 * time.Minute, nil
	case pb.PauseInverval_PAUSE_1_HOUR:
		return 1 * time.Hour, nil
	case pb.PauseInverval_PAUSE_24_HOURS:
		return 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown pause interval")
	}
}

// PauseConnection disconnects and schedules a reconnection in a timespan provided in the pause request
func (r *RPC) PauseConnection(ctx context.Context, in *pb.PauseRequest) (*pb.Payload, error) {
	if !r.netw.IsVPNActive() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	connectionStatus := r.connectionInfo.Status()
	if connectionStatus.State == pb.ConnectionState_PAUSED {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if r.connectionInfo.Status().IsMeshnetPeer {
		return &pb.Payload{Type: internal.CodePauseAttemptWhenConnectedToMeshPeer}, nil
	}

	pauseDuration, err := pauseIntervalToDuration(in.Interval)
	if err != nil {
		log.Error("failed to convert pause interval to duration:", err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}
	r.pauseManager.ScheduleReconnection(pauseDuration)

	_, err = r.DoPause(pauseDuration)
	if err != nil {
		r.pauseManager.CancelReconnection()
		log.Error("failed to disconnect when pausing the connection:", err)
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
