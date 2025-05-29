package state

import (
	"log"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/state/types"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

type InternalStateChangeNotif interface {
	NotifyChangeState(events.DataConnectChangeNotif) error
}

// ConnectionInfo stores data about currently active connection
// and provides notifications about changes for the internal listeners
// whenver an update of connection status happens
type ConnectionInfo struct {
	status        types.ConnectionStatus
	mu            sync.RWMutex
	internalNotif events.PublishSubcriber[events.DataConnectChangeNotif]
}

func NewConnectionInfo() *ConnectionInfo {
	return &ConnectionInfo{
		status:        types.ConnectionStatus{},
		internalNotif: &subs.Subject[events.DataConnectChangeNotif]{},
	}
}

// getTransferRatesForTunnel retrieves the upload (Tx) and download (Rx) transfer rates for the specified tunnel
// Returns:
//   - uint64: Upload transfer rate (Tx) in bytes per second, 0 in case of an error
//   - uint64: Download transfer rate (Rx) in bytes per second, 0 in case of an error
func (cs *ConnectionInfo) getTransferRatesForTunnel(tunnelName string) (uint64, uint64) {
	transferStats, err := tunnel.GetTransferRates(tunnelName)
	if err != nil {
		log.Println(internal.ErrorPrefix, "Failed to get transfer rates for tunnel:", err)
		return 0, 0
	}
	return transferStats.Tx, transferStats.Rx
}

// StatusWithTransferRates returns the current connection status with updated transfer rates
func (cs *ConnectionInfo) StatusWithTransferRates() types.ConnectionStatus {
	cs.mu.RLock()
	status := cs.status
	cs.mu.RUnlock()
	status.Tx, status.Rx = cs.getTransferRatesForTunnel(status.TunnelName)
	return status
}

func (cs *ConnectionInfo) Status() types.ConnectionStatus {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.status
}

func (cs *ConnectionInfo) setStatus(s types.ConnectionStatus) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.status = s
}

func (c *ConnectionInfo) ConnectionStatusNotifyConnect(e events.DataConnect) error {
	var startTime *time.Time = nil
	var Rx uint64 = 0
	var Tx uint64 = 0

	//invariant: for DataConnect possible values of EvenStatus are either connected or connecting
	connectionStatus := pb.ConnectionState_CONNECTING
	if e.EventStatus == events.StatusSuccess {
		connectionStatus = pb.ConnectionState_CONNECTED
		start := time.Now()
		startTime = &start
		Tx, Rx = c.getTransferRatesForTunnel(e.TunnelName)
	}

	status := types.ConnectionStatus{
		State:             connectionStatus,
		Technology:        e.Technology,
		Protocol:          e.Protocol,
		IP:                e.IP,
		Name:              e.Name,
		Hostname:          e.Hostname,
		Country:           e.TargetServerCountry,
		CountryCode:       e.TargetServerCountryCode,
		City:              e.TargetServerCity,
		StartTime:         startTime,
		IsVirtualLocation: e.IsVirtualLocation,
		IsPostQuantum:     e.IsPostQuantum,
		IsObfuscated:      e.IsObfuscated,
		TunnelName:        e.TunnelName,
		IsMeshnetPeer:     e.IsMeshnetPeer,
		Rx:                Rx,
		Tx:                Tx,
	}
	c.setStatus(status)
	c.internalNotif.Publish(events.DataConnectChangeNotif{Status: status})
	return nil
}

func (c *ConnectionInfo) ConnectionStatusNotifyDisconnect(events.DataDisconnect) error {
	status := types.ConnectionStatus{
		State:     pb.ConnectionState_DISCONNECTED,
		StartTime: nil,
	}
	c.setStatus(status)
	c.internalNotif.Publish(events.DataConnectChangeNotif{Status: status})
	return nil
}

func (c *ConnectionInfo) Subscribe(to InternalStateChangeNotif) {
	c.internalNotif.Subscribe(to.NotifyChangeState)
}
