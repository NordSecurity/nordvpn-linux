package daemon

import (
	"context"
	"log"
	"net"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// containsPrivateNetwork returns true if subnets contains a private network
func containsPrivateNetwork(subnet string) bool {
	if net, err := netip.ParsePrefix(subnet); err != nil {
		log.Println("Failed to parse subnet: ", err)
	} else if net.Addr().IsPrivate() || net.Addr().IsLinkLocalUnicast() {
		return true
	}
	return false
}

// isSubnetValid returns true if subnet is valid and false and appropriate error code when it's invalid.
func isSubnetValid(subnet string, currentSubnets config.Subnets, remove bool) (bool, int64) {
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, internal.CodeAllowlistInvalidSubnet
	}

	if _, ok := currentSubnets[subnet]; ok != remove {
		return false, internal.CodeAllowlistSubnetNoop
	}

	return true, 0
}

func arePortsValid(portRangeStart int64, portRangeEnd int64, currentPorts config.PortSet, remove bool) (bool, int64) {
	if portRangeStart < internal.AllowlistMinPort || portRangeStart > internal.AllowlistMaxPort {
		return false, internal.CodeAllowlistPortOutOfRange
	}

	if portRangeEnd != 0 &&
		((portRangeStart < internal.AllowlistMinPort || portRangeStart > internal.AllowlistMaxPort) ||
			portRangeEnd < portRangeStart) {
		return false, internal.CodeAllowlistPortOutOfRange
	}

	if remove {
		for port := portRangeStart; port <= portRangeEnd; port++ {
			if _, ok := currentPorts[port]; !ok {
				return false, internal.CodeAllowlistPortNoop
			}
		}
	}

	if portRangeEnd == 0 {
		if _, ok := currentPorts[portRangeStart]; ok != remove {
			return false, internal.CodeAllowlistPortNoop
		}
	}
	return true, 0
}

func (r *RPC) getNewAllowlist(req *pb.SetAllowlistRequest, remove bool) (config.Allowlist, int64) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, "reading config: %w", err)
		return config.Allowlist{}, internal.CodeConfigError
	}

	allowlist := cfg.AutoConnectData.Allowlist

	switch request := req.Request.(type) {
	case *pb.SetAllowlistRequest_SetAllowlistSubnetRequest:
		subnet := request.SetAllowlistSubnetRequest.GetSubnet()

		if cfg.LanDiscovery &&
			containsPrivateNetwork(subnet) {
			return config.Allowlist{}, internal.CodePrivateSubnetLANDiscovery
		}

		if valid, errorCode := isSubnetValid(subnet,
			cfg.AutoConnectData.Allowlist.Subnets,
			remove); !valid {
			return config.Allowlist{}, errorCode
		}

		allowlist.UpdateSubnets(subnet, remove)
	case *pb.SetAllowlistRequest_SetAllowlistPortsRequest:
		if request.SetAllowlistPortsRequest.IsUdp {
			portRange := request.SetAllowlistPortsRequest.GetPortRange()
			if valid, errorCode := arePortsValid(portRange.GetStartPort(),
				portRange.GetEndPort(),
				allowlist.Ports.UDP,
				remove); !valid {
				return config.Allowlist{}, errorCode
			}

			allowlist.UpdateUDPPorts(getPortsInARange(portRange.GetStartPort(), portRange.GetEndPort()), remove)
		}

		if request.SetAllowlistPortsRequest.IsTcp {
			portRange := request.SetAllowlistPortsRequest.GetPortRange()
			if valid, errorCode := arePortsValid(portRange.GetStartPort(),
				portRange.GetEndPort(),
				allowlist.Ports.TCP,
				remove); !valid {
				return config.Allowlist{}, errorCode
			}

			allowlist.UpdateTCPPorts(getPortsInARange(portRange.GetStartPort(), portRange.GetEndPort()), remove)
		}
	}

	return allowlist, internal.CodeSuccess
}

func (r *RPC) handleNewAllowlist(allowlist config.Allowlist) int64 {
	if err := r.netw.SetAllowlist(allowlist); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return internal.CodeFailure
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Allowlist = allowlist
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return internal.CodeConfigError
	}

	r.events.Settings.Allowlist.Publish(events.DataAllowlist{
		TCPPorts: allowlist.GetTCPPorts(),
		UDPPorts: allowlist.GetUDPPorts(),
		Subnets:  allowlist.GetSubnets(),
	})

	return internal.CodeSuccess
}

func getPortsInARange(start int64, stop int64) []int64 {
	ports := []int64{start}
	for port := start + 1; port <= stop; port++ {
		ports = append(ports, port)
	}
	return ports
}

func (r *RPC) SetAllowlist(ctx context.Context, in *pb.SetAllowlistRequest) (*pb.Payload, error) {
	allowlist, code := r.getNewAllowlist(in, false)
	if code != internal.CodeSuccess {
		return &pb.Payload{
			Type: code,
		}, nil
	}

	return &pb.Payload{
		Type: r.handleNewAllowlist(allowlist),
	}, nil
}

func (r *RPC) UnsetAllowlist(ctx context.Context, in *pb.SetAllowlistRequest) (*pb.Payload, error) {
	allowlist, code := r.getNewAllowlist(in, true)
	if code != internal.CodeSuccess {
		return &pb.Payload{
			Type: code,
		}, nil
	}

	return &pb.Payload{
		Type: r.handleNewAllowlist(allowlist),
	}, nil
}

func (r *RPC) UnsetAllAllowlist(ctx context.Context, in *pb.Empty) (*pb.Payload, error) {
	return &pb.Payload{
		Type: r.handleNewAllowlist(config.NewAllowlist([]int64{}, []int64{}, []string{})),
	}, nil
}
