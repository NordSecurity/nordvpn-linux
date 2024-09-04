package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// ConnectCancel cancels an active call for connect to VPN server or meshnet exit node and returns
// response code immediately without waiting for an actual cancel.
func (r *RPC) ConnectCancel(_ context.Context, _ *pb.Empty) (*pb.Payload, error) {
	t := internal.CodeNothingToDo
	cancel, _ := r.connectContext.CancelFunc()
	if cancel != nil {
		t = internal.CodeSuccess
		cancel()
	}
	return &pb.Payload{
		Type: t,
	}, nil
}
