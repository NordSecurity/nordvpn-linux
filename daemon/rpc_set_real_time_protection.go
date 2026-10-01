package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) SetRealTimeProtection(
	ctx context.Context,
	in *pb.SetRealTimeProtectionRequest,
) (*pb.SetRealTimeProtectionResponse, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error("loading config:", err)
	}

	realTimeProtection := in.GetRealTimeProtection()

	if cfg.AutoConnectData.RealTimeProtection == realTimeProtection {
		return &pb.SetRealTimeProtectionResponse{
			Response: &pb.SetRealTimeProtectionResponse_ErrorCode{ErrorCode: pb.SetErrorCode_ALREADY_SET},
		}, nil
	}

	nameservers := r.nameservers.Get(realTimeProtection)

	if err := r.netw.SetDNS(nameservers); err != nil {
		log.Error("applying DNS to networker:", err)
		return &pb.SetRealTimeProtectionResponse{
			Response: &pb.SetRealTimeProtectionResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.RealTimeProtection = realTimeProtection
		c.AutoConnectData.DNS = nil
		return c
	}); err != nil {
		log.Error("saving config:", err)
		return &pb.SetRealTimeProtectionResponse{
			Response: &pb.SetRealTimeProtectionResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
		}, nil
	}
	r.events.Settings.RealTimeProtection.Publish(in.GetRealTimeProtection())

	if cfg.AutoConnectData.DNS != nil && realTimeProtection {
		return &pb.SetRealTimeProtectionResponse{
			Response: &pb.SetRealTimeProtectionResponse_SetRealTimeProtectionStatus{
				SetRealTimeProtectionStatus: pb.SetRealTimeProtectionStatus_RTP_CONFIGURED_DNS_RESET},
		}, nil
	}

	return &pb.SetRealTimeProtectionResponse{
		Response: &pb.SetRealTimeProtectionResponse_SetRealTimeProtectionStatus{
			SetRealTimeProtectionStatus: pb.SetRealTimeProtectionStatus_RTP_CONFIGURED},
	}, nil
}
