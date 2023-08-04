package daemon

import (
	"context"
	"log"
	"net/netip"

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

	subnets := cfg.AutoConnectData.Allowlist.Subnets
	whitelist := cfg.AutoConnectData.Allowlist
	status := pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED

	if cfg.Mesh {
		token := cfg.TokensData[cfg.AutoConnectData.ID].Token
		if peers, err := r.meshRegistry.List(token, cfg.MeshDevice.ID); err == nil {
			r.netw.SetLanDiscoveryAndResetMesh(in.Enabled, peers)
		} else {
			log.Printf("Failed to fetch peers from the API when setting LAN discovery: %v", err)
		}
	} else {
		r.netw.SetLanDiscovery(in.Enabled)
	}

	if in.GetEnabled() {
		// Make a new list of allowlist of subnets based on the old allowlist, filter all of the
		// private networks as they will be allowed by lan-discovery.
		subnets = make(config.Subnets)
		for subnet := range cfg.AutoConnectData.Allowlist.Subnets {
			if prefix, err := netip.ParsePrefix(subnet); err != nil {
				log.Println("Failed to parse subnet: ", err)
			} else if !prefix.Addr().IsPrivate() {
				subnets[subnet] = true
			} else {
				status = pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED_ALLOWLIST_RESET
			}
		}

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
		c.AutoConnectData.Allowlist.Subnets = subnets
		return c
	}); err != nil {
		return &pb.SetLANDiscoveryResponse{
			Response: &pb.SetLANDiscoveryResponse_ErrorCode{
				ErrorCode: pb.SetErrorCode_CONFIG_ERROR,
			}}, nil
	}

	return &pb.SetLANDiscoveryResponse{
		Response: &pb.SetLANDiscoveryResponse_SetLanDiscoveryStatus{
			SetLanDiscoveryStatus: status}}, nil
}
