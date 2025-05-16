package state

import (
	"net/netip"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
)

type InternalStateChangeNotif interface {
	NotifyChangeState(events.DataConnectChangeNotif) error
}

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
}

// ConnectionInfo stores data about currently active connection
type ConnectionInfo struct {
	status        ConnectionStatus
	mu            sync.RWMutex
	internalNotif events.PublishSubcriber[events.DataConnectChangeNotif]
}

func NewConnectionInfo() *ConnectionInfo {
	return &ConnectionInfo{
		status:        ConnectionStatus{},
		internalNotif: &subs.Subject[events.DataConnectChangeNotif]{},
	}
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

func (c *ConnectionInfo) ConnectionStatusNotifyConnect(e events.DataConnect) error {
	//invariant: for DataConnect possible values of EvenStatus are either connected or connecting
	connectionStatus := pb.ConnectionState_CONNECTED
	if e.EventStatus == events.StatusAttempt {
		connectionStatus = pb.ConnectionState_CONNECTING
	}
	c.SetStatus(ConnectionStatus{
		State:           connectionStatus,
		Technology:      e.Technology,
		Protocol:        e.Protocol,
		IP:              e.IP,
		Name:            e.Name,
		Hostname:        e.Hostname,
		Country:         e.TargetServerCountry,
		CountryCode:     e.TargetServerCountryCode,
		City:            e.TargetServerCity,
		StartTime:       e.StartTime,
		VirtualLocation: e.IsVirtualLocation,
		PostQuantum:     e.IsPostQuantum,
		Obfuscated:      e.IsObfuscated,
		TunnelName:      e.TunnelName,
		MeshnetPeer:     e.IsMeshnetPeer,
	})
	c.internalNotif.Publish(events.DataConnectChangeNotif{})
	return nil
}

func (c *ConnectionInfo) ConnectionStatusNotifyDisconnect(_ events.DataDisconnect) error {
	c.SetStatus(ConnectionStatus{
		State:     pb.ConnectionState_DISCONNECTED,
		StartTime: nil,
	})
	c.internalNotif.Publish(events.DataConnectChangeNotif{})
	return nil
}

func (c *ConnectionInfo) Subscribe(to InternalStateChangeNotif) {
	c.internalNotif.Subscribe(to.NotifyChangeState)
}
