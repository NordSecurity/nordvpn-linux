package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetThreatProtectionLite(
	ctx context.Context,
	in *pb.SetThreatProtectionLiteRequest,
) (*pb.SetThreatProtectionLiteResponse, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	threatProtectionLite := in.GetThreatProtectionLite()

	if cfg.AutoConnectData.ThreatProtectionLite == threatProtectionLite {
		return &pb.SetThreatProtectionLiteResponse{
			Response: &pb.SetThreatProtectionLiteResponse_ErrorCode{ErrorCode: pb.SetErrorCode_ALREADY_SET},
		}, nil
	}

	nameservers := r.nameservers.Get(threatProtectionLite)

	if err := r.netw.SetDNS(nameservers); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.SetThreatProtectionLiteResponse{
			Response: &pb.SetThreatProtectionLiteResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.ThreatProtectionLite = threatProtectionLite
		c.AutoConnectData.DNS = nil
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.SetThreatProtectionLiteResponse{
			Response: &pb.SetThreatProtectionLiteResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
		}, nil
	}
	r.events.Settings.ThreatProtectionLite.Publish(in.GetThreatProtectionLite())

	if cfg.AutoConnectData.DNS != nil && threatProtectionLite {
		return &pb.SetThreatProtectionLiteResponse{
			Response: &pb.SetThreatProtectionLiteResponse_SetThreatProtectionLiteStatus{
				SetThreatProtectionLiteStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED_DNS_RESET},
		}, nil
	}

	return &pb.SetThreatProtectionLiteResponse{
		Response: &pb.SetThreatProtectionLiteResponse_SetThreatProtectionLiteStatus{
			SetThreatProtectionLiteStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED},
	}, nil
}
