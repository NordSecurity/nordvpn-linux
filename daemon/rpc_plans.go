package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) Plans(ctx context.Context, in *pb.Empty) (*pb.PlansResponse, error) {
	resp, err := r.api.Plans()
	if err != nil {
		log.Println(internal.ErrorPrefix, "retrieving plans:", err)
		return &pb.PlansResponse{
			Type: internal.CodeFailure,
		}, nil
	}

	var ret []*pb.Plan
	for _, plan := range *resp {
		p := &pb.Plan{
			Id:       plan.Identifier,
			Title:    plan.Title,
			Cost:     plan.Cost,
			Currency: plan.Currency,
		}
		ret = append(ret, p)
	}

	return &pb.PlansResponse{
		Type:  internal.CodeSuccess,
		Plans: ret,
	}, nil
}
