// Package vpn provides interface for vpn management.
package vpn

import (
	"context"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// VPN defines a set of operations that any type that wants to act as a vpn must implement.
type VPN interface {
	Start(context.Context, Credentials, ServerData) error
	Stop() error
	State() State // required because of OpenVPN
	IsActive() bool
	Tun() tunnel.T // required because of OpenVPN
	NetworkChanged() error
	// GetConnectionParameters returns ServerData of current connection and true if connection is established, or empty
	// ServerData and false if it isn't.
	GetConnectionParameters() (ServerData, bool)
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
	Name              string // status display only
	Country           string // status display only
	City              string // status display only
	Protocol          config.Protocol
	NordLynxPublicKey string
	Obfuscated        bool
	OpenVPNVersion    string
	VirtualLocation   bool
	PostQuantum       bool
	QuenchPort        int64
}
