package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/uievent"
)

func (r *RPC) RateConnection(ctx context.Context, in *pb.RateRequest) (*pb.Payload, error) {
	if ok, _ := r.ac.IsLoggedIn(); !ok {
		return nil, internal.ErrNotLoggedIn
	}

	event := &pb.UIEvent{
		FormReference: pb.UIEvent_CLI,
		ItemName:      pb.UIEvent_RATE_CONNECTION,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     ratingToUIEventValue(in.Rating),
	}
	action := uievent.ProtoToMooseStrings(event)
	r.events.Service.UiItemsClick.Publish(action)
	r.disconnectCounter.Swap(0)

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}

func ratingToUIEventValue(rating int64) pb.UIEvent_ItemValue {
	switch rating {
	case 1:
		return pb.UIEvent_CSAT_SCORE_1
	case 2:
		return pb.UIEvent_CSAT_SCORE_2
	case 3:
		return pb.UIEvent_CSAT_SCORE_3
	case 4:
		return pb.UIEvent_CSAT_SCORE_4
	case 5:
		return pb.UIEvent_CSAT_SCORE_5
	default:
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}
}
