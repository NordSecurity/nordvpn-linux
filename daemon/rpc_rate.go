package daemon

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) RateConnection(ctx context.Context, in *pb.RateRequest) (*pb.Payload, error) {
	if r.netw.LastServerName() != "" {
		r.events.Service.UiItemsClick.Publish(events.UiItemsAction{ItemName: "server_speed_rating", ItemType: "button", ItemValue: r.netw.LastServerName(), FormReference: fmt.Sprintf("%d", int(in.GetRating()))})
		return &pb.Payload{Type: internal.CodeSuccess}, nil
	}
	return &pb.Payload{Type: internal.CodeNothingToDo}, nil
}
