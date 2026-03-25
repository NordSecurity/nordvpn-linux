package daemon

import (
	"context"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) PauseConnection(ctx context.Context, in *pb.PauseRequest) (*pb.Payload, error) {
	if !r.netw.IsVPNActive() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if r.connectionInfo.Status().IsMeshnetPeer {
		return &pb.Payload{Type: internal.CodePauseAttemptWhenConnectedToMeshPeer}, nil
	}

	pauseDuration := time.Duration(in.Seconds * int64(time.Second))
	r.pauseManager.ScheduleReconnection(pauseDuration)

	_, err := r.DoDisconnect()
	if err != nil {
		r.pauseManager.CancelReconnection()
		log.Println(internal.ErrorPrefix, "failed to disconnect when pausing the connection:", err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
