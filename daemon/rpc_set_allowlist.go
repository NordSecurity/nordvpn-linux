package daemon

import (
	"context"
	"log"
	"net"
	"net/netip"
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/allowlist"
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
func isSubnetValid(subnet string, currentSubnets []string, remove bool) (bool, int64) {
	parsedAddress, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, internal.CodeAllowlistInvalidSubnet
	}

	// Do not allow IPv6 subnets
	if parsedAddress.To4() == nil {
		return false, internal.CodeAllowlistInvalidSubnet
	}

	if slices.Contains(currentSubnets, subnet) != remove {
		return false, internal.CodeAllowlistSubnetNoop
	}

	return true, 0
}

func arePortsValid(portRangeStart int64, portRangeEnd int64, currentPorts config.PortSet, remove bool) (bool, int64) {
	if portRangeStart < internal.AllowlistMinPort || portRangeStart > internal.AllowlistMaxPort {
		return false, internal.CodeAllowlistPortOutOfRange
	}

	if portRangeEnd != 0 &&
		((portRangeEnd < internal.AllowlistMinPort || portRangeEnd > internal.AllowlistMaxPort) ||
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
		log.Println(internal.ErrorPrefix, "reading config:", err)
		return config.Allowlist{}, internal.CodeConfigError
	}

	allowlist := cfg.AutoConnectData.Allowlist

	switch request := req.Request.(type) {
	case *pb.SetAllowlistRequest_SetAllowlistSubnetRequest:
		subnet := request.SetAllowlistSubnetRequest.GetSubnet()

		if cfg.LanDiscovery && containsPrivateNetwork(subnet) {
			return config.Allowlist{}, internal.CodePrivateSubnetLANDiscovery
		}

		if valid, errorCode := isSubnetValid(subnet, cfg.AutoConnectData.Allowlist.Subnets, remove); !valid {
			return config.Allowlist{}, errorCode
		}

		allowlist.UpdateSubnets(subnet, remove)
	case *pb.SetAllowlistRequest_SetAllowlistPortsRequest:
		if request.SetAllowlistPortsRequest.IsUdp {
			portRange := request.SetAllowlistPortsRequest.GetPortRange()
			if valid, errorCode := arePortsValid(portRange.GetStartPort(), portRange.GetEndPort(), allowlist.Ports.UDP, remove); !valid {
				return config.Allowlist{}, errorCode
			}

			allowlist.UpdateUDPPorts(getPortsInARange(portRange.GetStartPort(), portRange.GetEndPort()), remove)
		}

		if request.SetAllowlistPortsRequest.IsTcp {
			portRange := request.SetAllowlistPortsRequest.GetPortRange()
			if valid, errorCode := arePortsValid(portRange.GetStartPort(), portRange.GetEndPort(), allowlist.Ports.TCP, remove); !valid {
				return config.Allowlist{}, errorCode
			}

			allowlist.UpdateTCPPorts(getPortsInARange(portRange.GetStartPort(), portRange.GetEndPort()), remove)
		}
	}

	return allowlist, internal.CodeSuccess
}

func (r *RPC) handleNewAllowlist(allowlist config.Allowlist) int64 {
	// TODO: if save fails revert
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
		Subnets:  allowlist.Subnets,
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

// emitAllowlistAnalytics publishes an allowlist operation event to the debugger events channel
func (r *RPC) emitAllowlistAnalytics(event *allowlist.OperationEvent) {
	r.events.Debugger.DebuggerEvents.Publish(*event.ToDebuggerEvent())
}

func (r *RPC) SetAllowlist(ctx context.Context, in *pb.SetAllowlistRequest) (*pb.Payload, error) {
	allowlistCfg, code := r.getNewAllowlist(in, false)
	if code != internal.CodeSuccess {
		// emit failure event
		if event := allowlist.NewOperationEventFromRequest(in, allowlist.OpAdd, false, code); event != nil {
			r.emitAllowlistAnalytics(event)
		}
		return &pb.Payload{Type: code}, nil
	}

	resultCode := r.handleNewAllowlist(allowlistCfg)
	success := resultCode == internal.CodeSuccess

	// emit analytics event
	if event := allowlist.NewOperationEventFromRequest(in, allowlist.OpAdd, success, resultCode); event != nil {
		r.emitAllowlistAnalytics(event)
	}

	return &pb.Payload{Type: resultCode}, nil
}

func (r *RPC) UnsetAllowlist(ctx context.Context, in *pb.SetAllowlistRequest) (*pb.Payload, error) {
	allowlistCfg, code := r.getNewAllowlist(in, true)
	if code != internal.CodeSuccess {
		// emit failure event
		if event := allowlist.NewOperationEventFromRequest(in, allowlist.OpRemove, false, code); event != nil {
			r.emitAllowlistAnalytics(event)
		}
		return &pb.Payload{Type: code}, nil
	}

	resultCode := r.handleNewAllowlist(allowlistCfg)
	success := resultCode == internal.CodeSuccess

	// emit analytics event
	if event := allowlist.NewOperationEventFromRequest(in, allowlist.OpRemove, success, resultCode); event != nil {
		r.emitAllowlistAnalytics(event)
	}

	return &pb.Payload{Type: resultCode}, nil
}

// emitAllowlistSnapshot emits a snapshot event with the current allowlist configuration.
// Called on daemon startup to capture initial state.
func (r *RPC) emitAllowlistSnapshot() {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.WarningPrefix, "failed to load config for allowlist snapshot:", err)
		return
	}

	snapshot := allowlist.NewSnapshot(allowlist.SnapshotConfig{
		TCPPorts: cfg.AutoConnectData.Allowlist.GetTCPPorts(),
		UDPPorts: cfg.AutoConnectData.Allowlist.GetUDPPorts(),
		Subnets:  cfg.AutoConnectData.Allowlist.Subnets,
	})

	r.events.Debugger.DebuggerEvents.Publish(*snapshot.ToDebuggerEvent())
}

func (r *RPC) UnsetAllAllowlist(ctx context.Context, in *pb.Empty) (*pb.Payload, error) {
	resultCode := r.handleNewAllowlist(config.NewAllowlist([]int64{}, []int64{}, []string{}))
	success := resultCode == internal.CodeSuccess

	event := allowlist.NewClearOperation(success, resultCode)
	r.emitAllowlistAnalytics(event)

	return &pb.Payload{Type: resultCode}, nil
}
