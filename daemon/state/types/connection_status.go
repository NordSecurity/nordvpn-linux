package types

import (
	"net/netip"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// ConnectionStatus of a currently active connection
type ConnectionStatus struct {
	// State of the vpn. OpenVPN specific.
	State pb.ConnectionState
	// Technology, which may or may not match what's in the config
	Technology config.Technology
	// Protocol, which may or may not match what's in the config
	Protocol config.Protocol
	// IP of the other end of the connection
	IP netip.Addr
	// Name in a human readable form of the other end of the connection
	Name string
	// Hostname of the other end of the connection
	Hostname string
	// Country of the other end of the connection
	Country string
	// CountryCode of the other end of the connection
	CountryCode string
	// City of the other end of the connection
	City string
	// StartTime time of the connection start
	StartTime *time.Time
	// Is virtual server
	VirtualLocation bool
	// Is post quantum on
	PostQuantum bool
	// Is obfuscation on
	Obfuscated bool
	// Currently set tunnel name
	TunnelName string
	// Is meshnet peer on
	MeshnetPeer bool
	// Number of bytes transferred
	Rx uint64
	// Number of bytes uploaded
	Tx uint64
}
