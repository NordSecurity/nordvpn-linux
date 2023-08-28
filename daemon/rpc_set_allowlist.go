package daemon

import (
	"context"
	"log"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// containsPrivateNetwork returns true if subnets contains a private network
func containsPrivateNetwork(subnets []string) bool {
	for _, subnet := range subnets {
		if net, err := netip.ParsePrefix(subnet); err != nil {
			log.Println("Failed to parse subnet: ", err)
		} else if net.Addr().IsPrivate() || net.Addr().IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

func (r *RPC) SetAllowlist(ctx context.Context, in *pb.SetAllowlistRequest) (*pb.Payload, error) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	allowlist := config.NewAllowlist(
		in.GetAllowlist().GetPorts().GetUdp(),
		in.GetAllowlist().GetPorts().GetTcp(),
		in.GetAllowlist().GetSubnets(),
	)

	if cfg.LanDiscovery &&
		containsPrivateNetwork(in.GetAllowlist().GetSubnets()) {
		return &pb.Payload{
			Type: internal.CodePrivateSubnetLANDiscovery,
		}, nil
	}

	// If LAN discovery is enabled, we want to append LANs to the new allowlist and modify the
	// firewall. We do not want to add LANs to the configuration, so we have to create a copy.
	firewallAllowlist := allowlist
	if cfg.LanDiscovery {
		firewallAllowlist = addLANPermissions(firewallAllowlist)
	}

	if err := r.netw.SetAllowlist(firewallAllowlist); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeFailure,
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Allowlist = allowlist
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}
	r.events.Settings.Allowlist.Publish(events.DataAllowlist{
		TCPPorts: len(in.Allowlist.Ports.Tcp),
		UDPPorts: len(in.Allowlist.Ports.Udp),
		Subnets:  len(in.Allowlist.Subnets),
	})

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
