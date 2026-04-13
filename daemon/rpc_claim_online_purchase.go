package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) ClaimOnlinePurchase(ctx context.Context, in *pb.Empty) (*pb.ClaimOnlinePurchaseResponse, error) {
	isExpired, err := r.ac.IsVPNExpired()
	if err != nil {
		log.Error("failed to determine if user is registered: ", err)
		return &pb.ClaimOnlinePurchaseResponse{Success: false}, nil
	}

	if isExpired {
		log.Debug("user is expired when claiming online purchase.")
		return &pb.ClaimOnlinePurchaseResponse{Success: false}, nil
	}

	// notify state subscribers

	return &pb.ClaimOnlinePurchaseResponse{Success: true}, nil
}
