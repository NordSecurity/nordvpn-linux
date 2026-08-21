package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

func (r *RPC) GetRestrictedLogStrings(ctx context.Context, _ *pb.Empty) (*pb.LogSanitizationEvent, error) {
	restrictedLogs := r.logSanitizer.GetRestrictedLogStrings()

	return &pb.LogSanitizationEvent{
		RestrictedStrings: restrictedLogs,
	}, nil
}
