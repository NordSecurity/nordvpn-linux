package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) ClaimOnlinePurchase(ctx context.Context, in *pb.Empty) (*pb.ClaimOnlinePurchaseResponse, error) {
	isExpired, err := r.ac.IsVPNExpired()
	if err != nil {
		log.Println(internal.ErrorPrefix+" failed to determine if user is registered: ", err)
		return &pb.ClaimOnlinePurchaseResponse{Success: false}, nil
	}

	if isExpired {
		log.Println(internal.DebugPrefix + " user is expired when claiming online purchase.")
		return &pb.ClaimOnlinePurchaseResponse{Success: false}, nil
	}

	// notify state subscribers

	return &pb.ClaimOnlinePurchaseResponse{Success: true}, nil
}
