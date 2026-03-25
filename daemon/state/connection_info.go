package state

import (
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/state/types"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

type remainingDurationFunc func(startTime time.Time, duration time.Duration) int32

func getRemainingDuration(startTime time.Time, duration time.Duration) int32 {
	remainingDuration := time.Until(startTime.Add(duration))
	return int32(remainingDuration.Seconds())
}

type InternalStateChangeNotif interface {
	NotifyChangeState(events.DataConnectChangeNotif) error
}

type pauseData struct {
	pausedAt      time.Time
	pauseDuration time.Duration
}

// ConnectionInfo stores data about currently active connection
// and provides notifications about changes for the internal listeners
// whenver an update of connection status happens
type ConnectionInfo struct {
	status                types.ConnectionStatus
	fullyConnected        bool
	pauseData             *pauseData
	remainingDurationFunc remainingDurationFunc
	mu                    sync.RWMutex
	internalNotif         events.PublishSubcriber[events.DataConnectChangeNotif]
}

func NewConnectionInfo() *ConnectionInfo {
	return &ConnectionInfo{
		status:                types.ConnectionStatus{},
		internalNotif:         &subs.Subject[events.DataConnectChangeNotif]{},
		remainingDurationFunc: getRemainingDuration,
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

// Status returns the current connection status with updated transfer rates if internal state is connected
// and tunnel name is set
func (cs *ConnectionInfo) Status() types.ConnectionStatus {
	// we keep read lock in here, because the Tx/Rx rates are
	// a) merely point-in-time values
	// b) purely informative values
	// c) previous values are never used in any case
	// thus it is OK to not synchronize them each time
	cs.mu.RLock()
	status := cs.addPauseInfo(cs.status)
	cs.mu.RUnlock()
	if status.State == pb.ConnectionState_CONNECTED && status.TunnelName != "" {
		status.Tx, status.Rx = cs.getTransferRatesForTunnel(status.TunnelName)
	}
	return status
}

func (cs *ConnectionInfo) setStatus(s types.ConnectionStatus, fullyConnected bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if s.State != pb.ConnectionState_DISCONNECTED {
		// Don't override tunnel name as it comes from internal events
		s.TunnelName = cs.status.TunnelName
	}
	cs.status = cs.addPauseInfo(s)
	cs.fullyConnected = fullyConnected
}

// SetInitialConnecting should be executed as soon as connection started, even when no target server
// is known yet.
func (c *ConnectionInfo) SetInitialConnecting() {
	status := types.ConnectionStatus{State: pb.ConnectionState_CONNECTING}
	c.setStatus(status, false)
	c.internalNotif.Publish(events.DataConnectChangeNotif{Status: status})
}

func (c *ConnectionInfo) ConnectionStatusNotifyConnect(e events.DataConnect) error {
	var startTime *time.Time = nil

	fullyConnected := false
	var connectionStatus pb.ConnectionState
	switch e.EventStatus {
	case events.StatusAttempt:
		connectionStatus = pb.ConnectionState_CONNECTING
	case events.StatusCanceled, events.StatusFailure:
		connectionStatus = pb.ConnectionState_DISCONNECTED
	case events.StatusSuccess:
		connectionStatus = pb.ConnectionState_CONNECTED
		start := time.Now()
		startTime = &start
		fullyConnected = true
	}

	status := types.ConnectionStatus{
		State:              connectionStatus,
		Technology:         e.Technology,
		Protocol:           e.Protocol,
		IP:                 e.TargetServerIP,
		Name:               e.TargetServerName,
		Hostname:           e.TargetServerDomain,
		Country:            e.TargetServerCountry,
		CountryCode:        e.TargetServerCountryCode,
		City:               e.TargetServerCity,
		StartTime:          startTime,
		IsVirtualLocation:  e.IsVirtualLocation,
		IsPostQuantum:      e.IsPostQuantum,
		IsObfuscated:       e.IsObfuscated,
		IsMeshnetPeer:      e.IsMeshnetPeer,
		RecommendationUUID: e.RecommendationUUID,
	}

	c.setStatus(status, fullyConnected)
	c.internalNotif.Publish(events.DataConnectChangeNotif{Status: status})
	return nil
}

func (c *ConnectionInfo) ConnectionStatusNotifyDisconnect(e events.DataDisconnect) error {
	if e.IsRefresh {
		return nil
	}
	status := types.ConnectionStatus{
		State:              pb.ConnectionState_DISCONNECTED,
		TunnelName:         "",
		StartTime:          nil,
		RecommendationUUID: "",
	}

	c.setStatus(status, false)
	c.internalNotif.Publish(events.DataConnectChangeNotif{Status: c.status})
	return nil
}

func (c *ConnectionInfo) ConnectionStatusNotifyInternalConnect(
	e vpn.ConnectEvent,
) error {
	state := pb.ConnectionState_CONNECTED

	if e.Status != events.StatusSuccess {
		state = pb.ConnectionState_CONNECTING
	}
	return c.notifyInternalState(state, e.TunnelName)
}

func (c *ConnectionInfo) ConnectionStatusNotifyInternalDisconnect(
	status events.TypeEventStatus,
) error {
	// Currently only StatusSuccess is being reported in case disconnect fails internally
	return c.notifyInternalState(pb.ConnectionState_DISCONNECTED, "")
}

func (c *ConnectionInfo) notifyInternalState(
	state pb.ConnectionState,
	tunnelName string,
) error {
	c.mu.Lock()
	// Always set tunnelName as internal event may be shot before the real event
	c.status.TunnelName = tunnelName
	if !c.fullyConnected {
		c.mu.Unlock()
		return nil
	}
	c.status.State = state
	c.status = c.addPauseInfo(c.status)
	c.mu.Unlock()
	c.internalNotif.Publish(events.DataConnectChangeNotif{Status: c.status})
	return nil
}

func (c *ConnectionInfo) Subscribe(to InternalStateChangeNotif) {
	c.internalNotif.Subscribe(to.NotifyChangeState)
}

// Pause sets the pause data. All the subsequent Disconnect events will be sent out with State set to
// pb.ConnectionState_PAUSED until Unpause is called.
func (c *ConnectionInfo) Pause(pausedAt time.Time, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pauseData = &pauseData{
		pausedAt:      pausedAt,
		pauseDuration: duration,
	}
}

// Unpause unsets the pause data.
func (c *ConnectionInfo) Unpause() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pauseData = nil
}

// addPauseInfo adds pause info to the status if pause data is set
func (cs *ConnectionInfo) addPauseInfo(status types.ConnectionStatus) types.ConnectionStatus {
	if cs.pauseData == nil {
		return status
	}

	status.State = pb.ConnectionState_PAUSED
	status.PausedAt = cs.pauseData.pausedAt.Unix()
	status.PauseRemainingTimeSec = cs.remainingDurationFunc(cs.pauseData.pausedAt, cs.pauseData.pauseDuration)

	return status
}
