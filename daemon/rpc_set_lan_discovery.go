package daemon

import (
	"context"
	"log"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
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
	allowlist := cfg.AutoConnectData.Allowlist
	status := pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED

	r.netw.SetLanDiscovery(in.Enabled)

	if in.GetEnabled() {
		// Make a new list of allowlist of subnets based on the old allowlist, filter all of the
		// private networks as they will be allowed by lan-discovery.
		subnets = make([]string, 0)
		for _, subnet := range cfg.AutoConnectData.Allowlist.Subnets {
			if prefix, err := netip.ParsePrefix(subnet); err != nil {
				log.Println("Failed to parse subnet: ", err)
			} else if !prefix.Addr().IsPrivate() && !prefix.Addr().IsLinkLocalUnicast() {
				subnets = append(subnets, subnet)
			} else {
				status = pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED_ALLOWLIST_RESET
			}
		}

		cfg.AutoConnectData.Allowlist.Subnets = subnets
		allowlist = cfg.AutoConnectData.Allowlist
	}

	if err := r.netw.SetAllowlist(allowlist); err != nil {
		log.Printf("Failed to set allowlist: %v", err)
		return &pb.SetLANDiscoveryResponse{
			Response: &pb.SetLANDiscoveryResponse_ErrorCode{
				ErrorCode: pb.SetErrorCode_FAILURE,
			},
		}, nil
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

	r.events.Settings.LANDiscovery.Publish(in.GetEnabled())
	if in.GetEnabled() {
		// when LAN discovery is enabled notify that allow list changed
		r.events.Settings.Allowlist.Publish(events.DataAllowlist{
			TCPPorts: cfg.AutoConnectData.Allowlist.Ports.TCP.ToSlice(),
			UDPPorts: cfg.AutoConnectData.Allowlist.Ports.UDP.ToSlice(),
			Subnets:  subnets,
		})
	}

	return &pb.SetLANDiscoveryResponse{
		Response: &pb.SetLANDiscoveryResponse_SetLanDiscoveryStatus{
			SetLanDiscoveryStatus: status}}, nil
}
