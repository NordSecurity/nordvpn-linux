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
	if in != nil &&
		in.FormReference != pb.UIEvent_FORM_REFERENCE_UNSPECIFIED &&
		in.ItemName != pb.UIEvent_ITEM_NAME_UNSPECIFIED &&
		in.ItemType != pb.UIEvent_ITEM_TYPE_UNSPECIFIED {
		// only valid event should be sent
		action := uievent.ProtoToMooseStrings(in)
		if in.GetItemType() == pb.UIEvent_SHOW {
			r.events.Service.UiItemsShow.Publish(action)
		} else {
			r.events.Service.UiItemsClick.Publish(action)
		}
	}
	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
