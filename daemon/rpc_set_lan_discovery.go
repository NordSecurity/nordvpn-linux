package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

func (r *RPC) SetLANDiscovery(ctx context.Context, in *pb.SetLANDiscoveryRequest) (*pb.SetLANDiscoveryResponse, error) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		return &pb.SetLANDiscoveryResponse{
			Response: &pb.SetLANDiscoveryResponse_ErrorCode{
				ErrorCode: pb.SetErrorCode_CONFIG_ERROR,
			}}, nil
	}

	if cfg.LanDiscovery == in.GetEnabled() {
		return &pb.SetLANDiscoveryResponse{
			Response: &pb.SetLANDiscoveryResponse_ErrorCode{
				ErrorCode: pb.SetErrorCode_ALREADY_SET,
			}}, nil
	}

	whitelist := cfg.AutoConnectData.Allowlist
	if in.GetEnabled() {
		whitelist = addLANPermissions(cfg.AutoConnectData.Allowlist)
	}

	if r.netw.IsVPNActive() || cfg.KillSwitch {
		if err := r.netw.UnsetAllowlist(); err != nil {
			log.Printf("Failed to unset whitelist: %v", err)
			return &pb.SetLANDiscoveryResponse{
				Response: &pb.SetLANDiscoveryResponse_ErrorCode{
					ErrorCode: pb.SetErrorCode_FAILURE,
				},
			}, nil
		}

		if err := r.netw.SetAllowlist(whitelist); err != nil {
			log.Printf("Failed to set whitelist: %v", err)
			return &pb.SetLANDiscoveryResponse{
				Response: &pb.SetLANDiscoveryResponse_ErrorCode{
					ErrorCode: pb.SetErrorCode_FAILURE,
				},
			}, nil
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.LanDiscovery = in.GetEnabled()
		return c
	}); err != nil {
		return &pb.SetLANDiscoveryResponse{
			Response: &pb.SetLANDiscoveryResponse_ErrorCode{
				ErrorCode: pb.SetErrorCode_CONFIG_ERROR,
			}}, nil
	}

	return &pb.SetLANDiscoveryResponse{
		Response: &pb.SetLANDiscoveryResponse_SetLanDiscoveryStatus{
			SetLanDiscoveryStatus: pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED}}, nil
}
