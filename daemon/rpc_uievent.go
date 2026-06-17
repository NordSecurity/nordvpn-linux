package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/uievent"
)

// ReportUIEvent receives a UIEvent from any client and publishes the
// corresponding analytics action to the Moose pipeline.
func (r *RPC) ReportUIEvent(
	_ context.Context, in *pb.UIEvent,
) (*pb.Payload, error) {
	if in != nil {
		action := uievent.ProtoToMooseStrings(in)
		r.events.Service.UiItemsClick.Publish(action)
	}
	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
