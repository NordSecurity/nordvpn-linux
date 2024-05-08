package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) ClaimOnlinePurchase(ctx context.Context, in *pb.Empty) (*pb.Empty, error) {
	isExpired, err := r.ac.IsVPNExpired()
	if err != nil {
		log.Println(internal.ErrorPrefix+" failed to determine if user is registered: ", err)
	}

	if isExpired {
		log.Println(internal.DebugPrefix + " user is expired when caliming online purchase.")
		return &pb.Empty{}, nil
	}

	log.Println(internal.DebugPrefix + " send user subscribed notification.")
	// notify state subscribers

	return &pb.Empty{}, nil
}
