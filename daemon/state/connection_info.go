package state

import (
	"net/netip"
	"sync"
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
}

// ConnectionInfo stores data about currently active connection
type ConnectionInfo struct {
	status ConnectionStatus
	mu     sync.RWMutex
}

func NewConnectionInfo() *ConnectionInfo {
	return &ConnectionInfo{status: ConnectionStatus{}}
}

func (cs *ConnectionInfo) Status() ConnectionStatus {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.status
}

func (cs *ConnectionInfo) SetStatus(s ConnectionStatus) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.status = s
}
