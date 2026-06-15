package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// SendUIEvent is a lightweight RPC endpoint for standalone UI analytics events.
// The actual event extraction and publishing is handled by the gRPC middleware
// (uievent.Middleware), which intercepts all incoming requests and publishes
// any UI event metadata to the Moose analytics pipeline.
func (r *RPC) SendUIEvent(_ context.Context, _ *pb.Empty) (*pb.Payload, error) {
	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
