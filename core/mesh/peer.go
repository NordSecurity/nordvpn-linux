package mesh

import (
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"

	"github.com/google/uuid"
)

// MachineMap is used to refresh mesh.
type MachineMap struct {
	Machine
	Peers MachinePeers
	// Hosts maps hostnames to IP addresses
	Hosts
	// Raw is unprocessed API response passed directly to libtelio
	Raw []byte
}

// OperatingSystem defines an operating system in use.
type OperatingSystem struct {
	// Name is always 'linux' in our case.
	Name string
	// Distro can be found under the NAME key in /etc/os-release
	Distro string
}

// Machine represents current device in meshnet.
type Machine struct {
	// ID uniquely identifies the peer
	ID uuid.UUID
	// HardwareID uniquely identifies the physical machine,
	// without it, only one account can be used per machine.
	HardwareID uuid.UUID
	// Hostname to ping the peer by.
	Hostname string
	// OS where the peer is running on.
	OS OperatingSystem
	// PublicKey is a base64 encoded string.
	PublicKey string
	// Endpoints are used to reach the peer within meshnet.
	Endpoints []netip.AddrPort
	// Address is a meshnet IP address of a peer
	Address         netip.Addr
	SupportsRouting bool
}

func (s Machine) ToProtobuf() *pb.Peer {
	ip := ""
	if s.Address.IsValid() {
		ip = s.Address.String()
	}
	return &pb.Peer{
		Identifier: s.ID.String(),
		Pubkey:     s.PublicKey,
		Ip:         ip,
		Endpoints:  s.EndpointsString(),
		Os:         s.OS.Name,
		Distro:     s.OS.Distro,
		Hostname:   s.Hostname,
	}
}

type Machines []Machine

type MachinePeers []MachinePeer

// MachinePeer represents someone else's device in meshnet.
type MachinePeer struct {
	// ID uniquely identifies the peer
	ID uuid.UUID
	// Hostname to ping the peer by.
	Hostname string
	// OS where the peer is running on.
	OS OperatingSystem
	// PublicKey is a base64 encoded string.
	PublicKey string
	// Endpoints are used to reach the peer within meshnet.
	Endpoints []netip.AddrPort
	// Address is used to reach the peer outside meshnet.
	Address netip.Addr
	// Email of the owner.
	Email string
	// IsLocal to this meshnet. If not, it means invited.
	// Another way to represent this would be: enum Origin { Local, Invited }
	IsLocal bool
	// DoesPeerAllowRouting through it?
	DoesPeerAllowRouting bool
	// DoesPeerAllowInbound traffic to it?
	DoesPeerAllowInbound bool
	// DoesPeerAllowLocalNetwork access when routing through it?
	DoesPeerAllowLocalNetwork bool
	DoesPeerAllowFileshare    bool
	DoesPeerSupportRouting    bool
	// DoIAllowInbound traffic to this peer?
	DoIAllowInbound bool
	// DoIAllowRouting through this peer?
	DoIAllowRouting bool
	// DoIAllowLocalNetwork access when routing through me?
	DoIAllowLocalNetwork bool
	DoIAllowFileshare    bool
	AlwaysAcceptFiles    bool
	PeerNickname         string `json:"peer_nickname"`
}

func (p MachinePeer) ToProtobuf() *pb.Peer {
	ip := ""
	if p.Address.IsValid() {
		ip = p.Address.String()
	}
	return &pb.Peer{
		Identifier:            p.ID.String(),
		Pubkey:                p.PublicKey,
		Endpoints:             p.EndpointsString(),
		Ip:                    ip,
		Os:                    p.OS.Name,
		Distro:                p.OS.Distro,
		Hostname:              p.Hostname,
		Email:                 p.Email,
		IsInboundAllowed:      p.DoesPeerAllowInbound,
		IsRoutable:            p.DoesPeerAllowRouting,
		IsLocalNetworkAllowed: p.DoesPeerAllowLocalNetwork,
		IsFileshareAllowed:    p.DoesPeerAllowFileshare,
		DoIAllowInbound:       p.DoIAllowInbound,
		DoIAllowRouting:       p.DoIAllowRouting,
		DoIAllowLocalNetwork:  p.DoIAllowLocalNetwork,
		DoIAllowFileshare:     p.DoIAllowFileshare,
		AlwaysAcceptFiles:     p.AlwaysAcceptFiles,
	}
}

// EndpointsString could be replaced with
// slices.Map(p.Endpoints, func(s Stringer) string { return s.String() })
// once we upgrade to Go 1.18
func (s Machine) EndpointsString() []string {
	var endpoints []string
	for _, ep := range s.Endpoints {
		endpoints = append(endpoints, ep.String())
	}
	return endpoints
}

// EndpointsString could be replaced with
// slices.Map(p.Endpoints, func(s Stringer) string { return s.String() })
// once we upgrade to Go 1.18
func (p MachinePeer) EndpointsString() []string {
	var endpoints []string
	for _, ep := range p.Endpoints {
		endpoints = append(endpoints, ep.String())
	}
	return endpoints
}
