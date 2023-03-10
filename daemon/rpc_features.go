package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// IsFeatureSupported reports whether user can use a specific feature
func (r *RPC) IsFeatureSupported(ctx context.Context, req *pb.FeatureRequest) (*pb.Bool, error) {
	if !r.ac.IsLoggedIn() {
		return &pb.Bool{Value: false}, nil
	}

	isSupported, err := r.supportChecker.IsSupported(req.Feature)
	if err != nil {
		log.Printf("error checking feature compatibility: %s", err)
		return &pb.Bool{Value: false}, nil
	}
	return &pb.Bool{Value: isSupported}, nil
}
