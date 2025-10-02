package tray

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (ti *Instance) onDaemonStateEvent(item *pb.AppState) {
	changed := false
	switch st := item.GetState().(type) {
	case *pb.AppState_Error:
		changed = ti.handleErrorState()
	case *pb.AppState_ConnectionStatus:
		changed = ti.handleConnectionStatusState(st)
	case *pb.AppState_LoginEvent:
		changed = ti.handleLoginEventState()
	case *pb.AppState_SettingsChange:
		changed = ti.handleSettingsChangeState(st)
	case *pb.AppState_UpdateEvent:
		changed = ti.handleUpdateEventState(st)
	case *pb.AppState_AccountModification:
		changed = ti.handleAccountModificationState()
	case *pb.AppState_VersionHealth:
		changed = ti.handleVersionHealthState(st)
	default:
		log.Printf("%s %s Unknown state type: %T\n", logTag, internal.WarningPrefix, item)
	}

	ti.redraw(changed)
}

// handleErrorState handles the daemon error state.
func (ti *Instance) handleErrorState() bool {
	log.Printf("%s %s Received daemon error state\n", logTag, internal.ErrorPrefix)
	return ti.updateDaemonConnectionStatus(internal.ErrDaemonConnectionRefused.Error())
}

// handleConnectionStatusState handles the connection status state from the daemon.
func (ti *Instance) handleConnectionStatusState(st *pb.AppState_ConnectionStatus) bool {
	log.Printf("%s %s New connection status: %s\n",
		logTag,
		internal.InfoPrefix,
		st.ConnectionStatus.GetState(),
	)
	changed := ti.updateVpnStatus()
	changed = ti.updateRecentConnections() || changed
	return changed
}

// handleLoginEventState handles the login event state from the daemon.
func (ti *Instance) handleLoginEventState(st *pb.AppState_LoginEvent) bool {
	return ti.updateLoginStatus()
}

// handleSettingsChangeState handles the settings change state from the daemon.
func (ti *Instance) handleSettingsChangeState(st *pb.AppState_SettingsChange) bool {
	changed := ti.setSettings(st.SettingsChange)
	// identify whether we need to also update connections
	ti.connSensor.Set(connectionSettings{
		Obfuscated:      st.SettingsChange.Obfuscate,
		Protocol:        st.SettingsChange.Protocol,
		Technology:      st.SettingsChange.Technology,
		VirtualLocation: st.SettingsChange.VirtualLocation,
	})

	if ti.connSensor.ChangeDetected() {
		countryListChanged := ti.updateCountryList()
		specialtyServerListChanged := ti.updateSpecialtyServerList()
		recentsChanged := ti.updateRecentConnections()
		return changed || countryListChanged || specialtyServerListChanged || recentsChanged
	}

	return changed
}

// handleUpdateEventState handles the update event state from the daemon.
func (ti *Instance) handleUpdateEventState(st *pb.AppState_UpdateEvent) bool {
	switch st.UpdateEvent {
	case pb.UpdateEvent_SERVERS_LIST_UPDATE:
		countryListChanged := ti.updateCountryList()
		specialtyServerListChanged := ti.updateSpecialtyServerList()
		return countryListChanged || specialtyServerListChanged

	case pb.UpdateEvent_RECENTS_LIST_UPDATE:
		return ti.updateRecentConnections()
	}

	return false
}

// handleAccountModificationState handles the account modification state from the daemon.
func (ti *Instance) handleAccountModificationState() bool {
	return ti.updateAccountInfo()
}

// handleVersionHealthState handles the version health state from the daemon.
func (ti *Instance) handleVersionHealthState(st *pb.AppState_VersionHealth) bool {
	return ti.handleVersionHealthChange(st.VersionHealth)
}
