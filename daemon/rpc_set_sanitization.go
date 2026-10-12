package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) SetSanitization(ctx context.Context, req *pb.SetGenericRequest) (*pb.Payload, error) {
	cred, err := internal.UcredFromContext(ctx)
	if err != nil {
		log.Error("Getting ucred to disable log sanitization:", err)
		return &pb.Payload{Type: internal.CodeInternalError}, nil
	}

	if cred.Uid != 0 {
		return &pb.Payload{Type: internal.CodeDisableSanitizationNonRoot}, nil
	}

	if log.SetSanitization(req.Enabled) == req.Enabled {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
