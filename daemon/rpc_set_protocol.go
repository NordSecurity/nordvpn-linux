package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetProtocol(ctx context.Context, in *pb.SetProtocolRequest) (*pb.SetProtocolResponse, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.AutoConnectData.Protocol == in.Protocol {
		return &pb.SetProtocolResponse{
			Response: &pb.SetProtocolResponse_ErrorCode{
				ErrorCode: pb.SetErrorCode_ALREADY_SET,
			},
		}, nil
	}

	if cfg.Technology == config.Technology_NORDLYNX {
		return &pb.SetProtocolResponse{
			Response: &pb.SetProtocolResponse_SetProtocolStatus{
				SetProtocolStatus: pb.SetProtocolStatus_INVALID_TECHNOLOGY,
			},
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Protocol = in.GetProtocol()
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.SetProtocolResponse{
			Response: &pb.SetProtocolResponse_ErrorCode{
				ErrorCode: pb.SetErrorCode_CONFIG_ERROR,
			},
		}, nil
	}

	r.events.Settings.Protocol.Publish(in.GetProtocol())

	status := pb.SetProtocolStatus_PROTOCOL_CONFIGURED
	if r.netw.IsVPNActive() {
		status = pb.SetProtocolStatus_PROTOCOL_CONFIGURED_VPN_ON
	}

	return &pb.SetProtocolResponse{
		Response: &pb.SetProtocolResponse_SetProtocolStatus{
			SetProtocolStatus: status,
		},
	}, nil
}
