package firewall

import (
	"fmt"
	"net"
	"strings"
)

// FileshareRule rule allows tcp traffic to port 49111 from the given peerIP.
//
// -A INPUT -s <peerIP>/32 -p tcp -m tcp --dport 49111 -m comment --comment nordvpn -j ACCEPT
type FileshareRule struct {
	PeerIP string
}

func (f FileshareRule) ToArgs() []Args {
	command := fmt.Sprintf("-A INPUT -s %s/32 -p tcp -m tcp --dport 49111 -m comment --comment nordvpn -j ACCEPT", f.PeerIP)
	return []Args{strings.Split(command, "")}
}

func (f FileshareRule) ToUndoArgs() []Args {
	command := fmt.Sprintf("-D INPUT -s %s/32 -p tcp -m tcp --dport 49111 -m comment --comment nordvpn -j ACCEPT", f.PeerIP)
	return []Args{strings.Split(command, "")}
}

// AllowIncomingRule allows all incoming traffic from the given peerIP.
//
// -A INPUT -s <peerIP>/32 -m comment --comment nordvpn -j ACCEPT
type AllowIncomingRule struct {
	PeerIP string
}

func (a AllowIncomingRule) ToArgs() []Args {
	command := fmt.Sprintf("-A INPUT -s %s/32 -m comment --comment nordvpn -j ACCEPT", a.PeerIP)
	return []Args{strings.Split(command, "")}
}

func (a AllowIncomingRule) ToUndoArgs() []Args {
	command := fmt.Sprintf("-D INPUT -s %s/32 -m comment --comment nordvpn -j ACCEPT", a.PeerIP)
	return []Args{strings.Split(command, "")}
}

// DenyIncomingLanRule denies all incoming traffic form given peerIP to LAN subnets:
//
// - 169.254.0.0/16
//
// - 192.168.0.0/16
//
// - 172.16.0.0/12
//
// - 10.0.0.0/8
//
// -A INPUT -s <peerIP>/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP
//
// -A INPUT -s <peerIP>/32 -d 192.168.0.0/16 -m comment --comment nordvpn -j DROP
//
// -A INPUT -s <peerIP>/32 -d 172.16.0.0/12 -m comment --comment nordvpn -j DROP
//
// -A INPUT -s <peerIP>/32 -d 10.0.0.0/8 -m comment --comment nordvpn -j DROP
type DenyLocalIncomingRule struct {
	PeerIP string
}

func (d DenyLocalIncomingRule) ToArgs() []Args {
	return []Args{
		strings.Split(fmt.Sprintf("-A INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", d.PeerIP), " "),
		strings.Split(fmt.Sprintf("-A INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", d.PeerIP), " "),
		strings.Split(fmt.Sprintf("-A INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", d.PeerIP), " "),
		strings.Split(fmt.Sprintf("-A INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", d.PeerIP), " "),
	}
}

func (d DenyLocalIncomingRule) ToUndoArgs() []Args {
	return []Args{
		strings.Split(fmt.Sprintf("-D INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", d.PeerIP), " "),
		strings.Split(fmt.Sprintf("-D INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", d.PeerIP), " "),
		strings.Split(fmt.Sprintf("-D INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", d.PeerIP), " "),
		strings.Split(fmt.Sprintf("-D INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", d.PeerIP), " "),
	}
}

// BlockMeshRule blocks traffic form meshnet subnet, allowing only RELATED and ESTABLISHED
// traffic to the hosts deviceAddress.
//
// -A INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc <deviceAddress> -m comment --comment nordvpn -j ACCEPT
//
// -A INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP
type BlockMeshRule struct {
	DeviceAddress string
}

func (b BlockMeshRule) ToArgs() []Args {
	return []Args{
		strings.Split(fmt.Sprintf("-A INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc %s -m comment --comment nordvpn -j ACCEPT", b.DeviceAddress), " "),
		strings.Split("-A INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP", " "),
	}
}

func (b BlockMeshRule) ToUndoArgs() []Args {
	return []Args{
		strings.Split(fmt.Sprintf("-D INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc %s -m comment --comment nordvpn -j ACCEPT", b.DeviceAddress), " "),
		strings.Split("-D INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP", " "),
	}
}

// AllowlistSubnetRule allows incoming and outgoing traffic from subnet to iface.
//
// -A INPUT -s <subnet> -i <interface> -m comment --comment nordvpn -j ACCEPT
//
// -A OUTPUT -d <subnet> -o <interface> -m comment --comment nordvpn -j ACCEPT
type AllowlistSubnetRule struct {
	Subnet string
	Iface  string
}

func (a AllowlistSubnetRule) ToArgs() []Args {
	return []Args{
		strings.Split(fmt.Sprintf("-A INPUT -s %s -i %s -m comment --comment nordvpn -j ACCEPT", a.Subnet, a.Iface), " "),
		strings.Split(fmt.Sprintf("-A INPUT -d %s -o %s -m comment --comment nordvpn -j ACCEPT", a.Subnet, a.Iface), " "),
	}
}

func (a AllowlistSubnetRule) ToUndoArgs() []Args {
	return []Args{
		strings.Split(fmt.Sprintf("-D INPUT -s %s -i %s -m comment --comment nordvpn -j ACCEPT", a.Subnet, a.Iface), " "),
		strings.Split(fmt.Sprintf("-D INPUT -d %s -o %s -m comment --comment nordvpn -j ACCEPT", a.Subnet, a.Iface), " "),
	}
}

// AllowlistPortsRule allows incoming and outgoing traffic for ports from portRangeStart to portRangeEnd for the given
// protocol for the given iface.
//
// -A INPUT -i <interface> -p <protocol> -m <protocol> --dport <port> -m comment --comment nordvpn -j ACCEPT
//
// -A INPUT -i <interface> -p <protocol> -m <protocol> --sport <port> -m comment --comment nordvpn -j ACCEPT
//
// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --sport <port> -m comment --comment nordvpn -j ACCEPT
//
// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --dport <port> -m comment --comment nordvpn -j ACCEPT
type AllowlistPortsRule struct {
	PortRangeStart string
	PortRangeEnd   string
	Protocol       string
	Iface          string
}

func (a AllowlistPortsRule) ToArgs() []Args {
	portRangeArg := a.PortRangeStart

	if a.PortRangeStart != a.PortRangeEnd {
		portRangeArg = fmt.Sprintf("%s:%s", a.PortRangeStart, a.PortRangeEnd)
	}

	return []Args{
		strings.Split(fmt.Sprintf("-A INPUT -i %s -p %s -m %s --dport %s -m comment --comment nordvpn -j ACCEPT", a.Iface, a.Protocol, a.Protocol, portRangeArg), " "),
		strings.Split(fmt.Sprintf("-A INPUT -i %s -p %s -m %s --sport %s -m comment --comment nordvpn -j ACCEPT", a.Iface, a.Protocol, a.Protocol, portRangeArg), " "),
		strings.Split(fmt.Sprintf("-A OUTPUT -o %s -p %s -m %s --sport %s -m comment --comment nordvpn -j ACCEPT", a.Iface, a.Protocol, a.Protocol, portRangeArg), " "),
		strings.Split(fmt.Sprintf("-A OUTPUT -o %s -p %s -m %s --dport %s -m comment --comment nordvpn -j ACCEPT", a.Iface, a.Protocol, a.Protocol, portRangeArg), " "),
	}
}

func (a AllowlistPortsRule) ToUndoArgs() []Args {
	portRangeArg := a.PortRangeStart

	if a.PortRangeStart != a.PortRangeEnd {
		portRangeArg = fmt.Sprintf("%s:%s", a.PortRangeStart, a.PortRangeEnd)
	}

	return []Args{
		strings.Split(fmt.Sprintf("-D INPUT -i %s -p %s -m %s --dport %s -m comment --comment nordvpn -j ACCEPT", a.Iface, a.Protocol, a.Protocol, portRangeArg), " "),
		strings.Split(fmt.Sprintf("-D INPUT -i %s -p %s -m %s --sport %s -m comment --comment nordvpn -j ACCEPT", a.Iface, a.Protocol, a.Protocol, portRangeArg), " "),
		strings.Split(fmt.Sprintf("-D OUTPUT -o %s -p %s -m %s --sport %s -m comment --comment nordvpn -j ACCEPT", a.Iface, a.Protocol, a.Protocol, portRangeArg), " "),
		strings.Split(fmt.Sprintf("-D OUTPUT -o %s -p %s -m %s --dport %s -m comment --comment nordvpn -j ACCEPT", a.Iface, a.Protocol, a.Protocol, portRangeArg), " "),
	}
}

// ApiAllowlistRule allows any traffic from given Ifaces for traffic with the given Connmark.
//
// -A INPUT -i <iface> -m connmark --mark <connmark> -m comment --comment nordvpn -j ACCEPT
//
// -A OUTPUT -o <iface> -m mark --mark <connmark> -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
//
// -A OUTPUT -o <iface> -m connmark --mark <connmark> -m comment --comment nordvpn -j ACCEPT
type ApiAllowlistRule struct {
	Ifaces   []net.Interface
	Connmark uint32
}

func (a ApiAllowlistRule) ToArgs() []Args {
	args := []Args{}
	for _, iface := range a.Ifaces {
		args = append(args, strings.Split(fmt.Sprintf("-I INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, a.Connmark), " "))
		args = append(args, strings.Split(fmt.Sprintf("-I OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff", iface.Name, a.Connmark), " "))
		args = append(args, strings.Split(fmt.Sprintf("-I OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, a.Connmark), " "))
	}
	return args
}

func (a ApiAllowlistRule) ToUndoArgs() []Args {
	args := []Args{}
	for _, iface := range a.Ifaces {
		args = append(args, strings.Split(fmt.Sprintf("-D INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, a.Connmark), " "))
		args = append(args, strings.Split(fmt.Sprintf("-D OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff", iface.Name, a.Connmark), " "))
		args = append(args, strings.Split(fmt.Sprintf("-D OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, a.Connmark), " "))
	}
	return args
}

// BlockTrafficRule blocks all incoming/outgoing traffic for the given iface.
//
// -A INPUT -i <iface> -m comment --comment nordvpn -j DROP
// -A OUTPUT -o <iface> -m comment --comment nordvpn -j DROP
type BlockTrafficRule struct {
	Iface string
}

func (b *BlockTrafficRule) ToArgs() []Args {
	return []Args{
		strings.Split(fmt.Sprintf("-A INPUT -i %s -m comment --comment nordvpn -j DROP", b.Iface), " "),
		strings.Split(fmt.Sprintf("-A OUTPUT -o %s -m comment --comment nordvpn -j DROP", b.Iface), " "),
	}
}

func (b *BlockTrafficRule) ToUndoArgs() []Args {
	return []Args{
		strings.Split(fmt.Sprintf("-D INPUT -i %s -m comment --comment nordvpn -j DROP", b.Iface), " "),
		strings.Split(fmt.Sprintf("-D OUTPUT -o %s -m comment --comment nordvpn -j DROP", b.Iface), " "),
	}
}
