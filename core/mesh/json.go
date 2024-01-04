package mesh

import (
	"encoding/json"
	"net/netip"

	"github.com/google/uuid"
)

type MachineCreateRequest struct {
	// PublicKey is a WireGuard public key used to encrypt outgoing packets.
	PublicKey string `json:"public_key"`
	// HardwareID is used in combination with PublicKey to allow registering
	// multiple machines on the same hardware.
	HardwareID uuid.UUID `json:"hardware_identifier"`
	// OS is always 'linux'
	OS string `json:"os"`
	// Distro can be found in /etc/os-release
	Distro string `json:"os_version"`
	// Endpoints are publicly routable IP and port pairs.
	Endpoints       []netip.AddrPort `json:"endpoints"`
	SupportsRouting bool             `json:"traffic_routing_supported"`
}

type MachineCreateResponse struct {
	// Identifier is an API generated unique identifier for a peer
	Identifier uuid.UUID `json:"identifier"`
	// PublicKey is a WireGuard public key used to encrypt outgoing packets.
	PublicKey string `json:"public_key"`
	// Hostname is a fully qualified domain name used to reach machine.
	// Also used as a human readable identifier.
	Hostname string `json:"hostname"`
	// OS is always 'linux'
	OS string `json:"os"`
	// Distro can be found in /etc/os-release
	Distro string `json:"os_version"`
	// Endpoints are publicly routable IP and port pairs.
	Endpoints []netip.AddrPort `json:"endpoints"`
	// Addresses belonging to 100.64.0.0/10 subnet.
	Addresses       []netip.Addr `json:"ip_addresses"`
	SupportsRouting bool         `json:"traffic_routing_supported"`
	Nickname        string       `json:"nickname"`
}

// MachineUpdateRequest is used to update one's meshnet device.
type MachineUpdateRequest struct {
	// TODO: Endpoints doesn't exist in documentation, check if is needed
	Endpoints       []netip.AddrPort `json:"endpoints"`
	SupportsRouting bool             `json:"traffic_routing_supported"`
	Nickname        string           `json:"nickname"`
}

type MachinePeerResponse struct {
	ID        uuid.UUID        `json:"identifier"`
	PublicKey string           `json:"public_key"`
	Hostname  string           `json:"hostname"`
	OS        string           `json:"os"`
	Distro    string           `json:"os_version"`
	Endpoints []netip.AddrPort `json:"endpoints"`
	Addresses []netip.Addr     `json:"ip_addresses"`
	IsLocal   bool             `json:"is_local"`
	Email     string           `json:"user_email"`

	// MachinePeer settings
	DoesPeerAllowRouting      bool `json:"peer_allows_traffic_routing"`
	DoesPeerAllowInbound      bool `json:"allow_outgoing_connections"`
	DoesPeerAllowLocalNetwork bool `json:"peer_allows_local_network_access"`
	DoesPeerSupportRouting    bool `json:"traffic_routing_supported"`
	DoesPeerAllowFileshare    bool `json:"peer_allows_send_files"`

	// Machine settings
	DoIAllowInbound      bool   `json:"allow_incoming_connections"`
	DoIAllowRouting      bool   `json:"allow_peer_traffic_routing"`
	DoIAllowLocalNetwork bool   `json:"allow_peer_local_network_access"`
	DoIAllowFileshare    bool   `json:"allow_peer_send_files"`
	AlwaysAcceptFiles    bool   `json:"always_accept_files"`
	Nickname             string `json:"nickname"`
}

type MachineMapResponse struct {
	ID        uuid.UUID `json:"identifier"`
	PublicKey string    `json:"public_key"`
	// Hostname is a fully qualified domain name used to reach machine.
	// Also used as a human readable identifier.
	Hostname string `json:"hostname"`
	// OS is always 'linux'
	OS string `json:"os"`
	// Distro can be found in /etc/os-release
	Distro string `json:"os_version"`
	// Endpoints are publicly routable IP and port pairs.
	Endpoints []netip.AddrPort `json:"endpoints"`
	// Addresses belonging to 100.64.0.0/10 subnet.
	Addresses       []netip.Addr          `json:"ip_addresses"`
	SupportsRouting bool                  `json:"traffic_routing_supported"`
	DNS             DNS                   `json:"dns"`
	Peers           []MachinePeerResponse `json:"peers"`
	Nickname        string                `json:"nickname"`
}

// PeerUpdateRequest is used to update one's peer.
type PeerUpdateRequest struct {
	DoIAllowInbound      bool   `json:"allow_incoming_connections"`
	DoIAllowRouting      bool   `json:"allow_peer_traffic_routing"`
	DoIAllowLocalNetwork bool   `json:"allow_peer_local_network_access"`
	DoIAllowFileshare    bool   `json:"allow_peer_send_files"`
	AlwaysAcceptFiles    bool   `json:"always_accept_files"`
	Nickname             string `json:"nickname"`
}

func NewPeerUpdateRequest(peer MachinePeer) PeerUpdateRequest {
	return PeerUpdateRequest{
		DoIAllowInbound:      peer.DoIAllowInbound,
		DoIAllowRouting:      peer.DoIAllowRouting,
		DoIAllowLocalNetwork: peer.DoIAllowLocalNetwork,
		DoIAllowFileshare:    peer.DoIAllowFileshare,
		AlwaysAcceptFiles:    peer.AlwaysAcceptFiles,
		Nickname:             peer.Nickname,
	}
}

// Invitation to/from other user.
type Invitation struct {
	ID                uuid.UUID `json:"token"`
	Email             string    `json:"email"`
	OS                string    `json:"os"`
	AllowInbound      bool      `json:"allow_incoming_connections"`
	AllowRouting      bool      `json:"allow_peer_traffic_routing"`
	AllowLocalNetwork bool      `json:"allow_peer_local_network_access"`
}

type AcceptInvitationRequest struct {
	AllowInbound      bool `json:"allow_incoming_connections"`
	AllowRouting      bool `json:"allow_peer_traffic_routing"`
	AllowLocalNetwork bool `json:"allow_peer_local_network_access"`
	AllowFileshare    bool `json:"allow_peer_send_files"`
}

type SendInvitationRequest struct {
	Email             string `json:"email"`
	AllowInbound      bool   `json:"allow_incoming_connections"`
	AllowRouting      bool   `json:"allow_peer_traffic_routing"`
	AllowLocalNetwork bool   `json:"allow_peer_local_network_access"`
	AllowFileshare    bool   `json:"allow_peer_send_files"`
}

type NotificationNewTransactionRequest struct {
	ReceiverMachineIdentifier string `json:"receiver_machine_identifier"`
	FileName                  string `json:"file_name"`
	FileCount                 int    `json:"file_count"`
}

// DNS defines mapping between peer name and ip used by magic DNS within the mesh network.
type DNS struct {
	Hosts Hosts `json:"hosts"`
}

// Hosts defines mapping between hostname an IP address.
type Hosts map[string]netip.Addr

// UnmarshalJSON customizes json deserialization.
func (h Hosts) UnmarshalJSON(b []byte) error {
	var unparsed map[string]string
	if err := json.Unmarshal(b, &unparsed); err != nil {
		return err
	}

	h = make(map[string]netip.Addr, len(unparsed))
	for name, addr := range unparsed {
		ip, err := netip.ParseAddr(addr)
		if err != nil {
			return err
		}
		h[name] = ip
	}
	return nil
}
