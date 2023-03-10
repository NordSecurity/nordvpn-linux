package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.Payload, error) {
	_, err := r.api.CreateUser(in.GetEmail(), in.GetPassword())
	if err != nil {
		log.Println(internal.ErrorPrefix, "registering user:", err)
		switch err {
		case core.ErrBadRequest:
			return &pb.Payload{
				Type: internal.CodeBadRequest,
			}, nil
		case core.ErrConflict:
			return &pb.Payload{
				Type: internal.CodeConflict,
			}, nil
		case core.ErrServerInternal:
			return &pb.Payload{
				Type: internal.CodeInternalError,
			}, nil
		default:
			return &pb.Payload{
				Type: internal.CodeFailure,
			}, nil
		}
	}

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
