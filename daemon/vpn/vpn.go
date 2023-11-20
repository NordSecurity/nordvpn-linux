// Package vpn provides interface for vpn management.
package vpn

import (
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// VPN defines a set of operations that any type that wants to act as a vpn must implement.
type VPN interface {
	Start(Credentials, ServerData) error
	Stop() error
	State() State
	IsActive() bool
	Tun() tunnel.T // required because of OpenVPN
	StateChanged() <-chan State
}

// NetworkChanger allows refreshing VPN connection without the need for full start/stop.
type NetworkChanger interface {
	NetworkChange() error
}

// Credentials define a possible set of credentials required to
// connect to the VPN server
type Credentials struct {
	OpenVPNUsername    string
	OpenVPNPassword    string
	NordLynxPrivateKey string
}

// IsOpenVPNDefined returns true if both username and password are
// defined
func (c Credentials) IsOpenVPNDefined() bool {
	return c.OpenVPNUsername != "" && c.OpenVPNPassword != ""
}

// ServerData required to connect to VPN server.
type ServerData struct {
	IP                netip.Addr
	Hostname          string // used in openvpn server certificate validation
	Country           string // status display only
	City              string // status display only
	Protocol          config.Protocol
	NordLynxPublicKey string
	Obfuscated        bool
	OpenVPNVersion    string
}
