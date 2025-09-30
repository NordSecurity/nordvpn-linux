package tray

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/snapconf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
)

const (
	// Notification messages
	labelLoginSuccess       = "You've successfully logged in"
	labelLogout             = "You've logged out"
	labelNotificationsOn    = "Notifications for NordVPN turned on"
	labelNotificationsOff   = "Notifications for NordVPN turned off"
	labelTrayOn             = "Tray for NordVPN turned on"
	labelTrayOff            = "Tray for NordVPN turned off"
	labelDaemonReconnected  = "Reconnected to NordVPN's background service"
	labelDaemonDisconnected = "Couldn't connect to NordVPN's background service. Please ensure the service is running."
	labelConnectedFormat    = "Connected to %s"
	labelDisconnectedFormat = "Disconnected from %s"
)

func (ti *Instance) handleVersionHealthChange(health *pb.VersionHealthStatus) bool {
	daemonError := ""
	switch int64(health.StatusCode) {
	case internal.CodeOffline:
		daemonError = cli.ErrInternetConnection.Error()
	case internal.CodeOutdated:
		daemonError = cli.ErrUpdateAvailable.Error()
	case internal.CodeSuccess:
		daemonError = ""
	default:
		// For unknown status codes, assume success (no error)
		daemonError = ""
	}

	changed := ti.updateDaemonConnectionStatus(daemonError)
	return changed
}

func (ti *Instance) MonitorConnection(ctx context.Context, conn *grpc.ClientConn) {
	log.Println(logTag, internal.InfoPrefix, "Starting to monitor daemon connection state")
	state := conn.GetState()
	// check if connection is already in ready state
	if state == connectivity.Ready {
		changed := ti.updateDaemonConnectionStatus("")
		ti.redraw(changed)
	}

	cancelContext, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	for {
		connExpired := !conn.WaitForStateChange(cancelContext, state)
		if connExpired {
			log.Println(logTag, internal.InfoPrefix, "Daemon connection state changed to: EXPIRED")
			return // ctx cancelled
		}

		switch conn.GetState() {
		case connectivity.Connecting:
		case connectivity.Idle:
		case connectivity.Ready:
			log.Println(logTag, internal.InfoPrefix, "Daemon connection state changed to: READY")
			// conn ready
			changed := ti.updateDaemonConnectionStatus("")
			ti.redraw(changed)
		case connectivity.Shutdown:
			log.Println(logTag, internal.InfoPrefix, "Daemon connection state changed to: SHUTDOWN")
			// conn terminated/shutdown
			return
		case connectivity.TransientFailure:
			log.Println(logTag, internal.InfoPrefix, "Daemon connection state changed to: TRANSIENT_FAILURE")
			// server likely down
			err := internal.ErrDaemonConnectionRefused.Error()
			changed := ti.updateDaemonConnectionStatus(err)
			ti.redraw(changed)
		}
	}
}

func (ti *Instance) update() {
	needsRedraw := false

	changed := ti.updateSettings()
	needsRedraw = changed

	changed = ti.updateCountryList()
	needsRedraw = needsRedraw || changed

	changed = ti.updateVpnStatus()
	needsRedraw = needsRedraw || changed

	changed = ti.updateSpecialtyServerList()
	needsRedraw = needsRedraw || changed

	changed = ti.updateLoginStatus()
	needsRedraw = needsRedraw || changed

	changed = ti.updateRecentConnections()
	needsRedraw = needsRedraw || changed

	changed = ti.updateAccountInfo()
	needsRedraw = needsRedraw || changed

	ti.redraw(needsRedraw)
}

func (ti *Instance) updateLoginStatus() bool {
	changed := false
	resp, err := ti.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil {
		changed = ti.updateDaemonConnectionStatus(messageForDaemonError(err))
		return changed
	}

	loggedIn := resp.GetIsLoggedIn()

	if !loggedIn && ti.state.loggedIn && ti.state.vpnStatus == pb.ConnectionState_CONNECTED {
		// reset the VPN info if the user logs out while connected to VPN
		changedVpn := ti.setVpnStatus(pb.ConnectionState_DISCONNECTED, "", "", "", "", false)
		if changedVpn {
			changed = true
		}
	}

	var notificationText string
	func() {
		ti.state.mu.Lock()
		defer ti.state.mu.Unlock()

		if !ti.state.loggedIn && loggedIn {
			ti.state.loggedIn = true
			changed = true
			notificationText = labelLoginSuccess
		} else if ti.state.loggedIn && !loggedIn {
			ti.state.loggedIn = false
			ti.accountInfo.reset()
			ti.state.accountName = ""
			changed = true
			notificationText = labelLogout
		}
	}()

	if notificationText != "" {
		ti.notify(NoForce, notificationText)
	}
	return changed
}

func (ti *Instance) updateVpnStatus() bool {
	log.Println(logTag, "Updating VPN status")
	resp, err := ti.client.Status(context.Background(), &pb.Empty{})
	if err != nil {
		log.Println(logTag, "Error getting VPN status:", err)
		changed := ti.updateDaemonConnectionStatus(messageForDaemonError(err))
		return changed
	}

	vpnStatus := resp.State
	vpnHostname := resp.Hostname
	vpnCity := resp.City
	vpnCountry := resp.Country
	vpnName := resp.Name
	if vpnName == "" {
		vpnName = vpnHostname
	}

	changed := ti.setVpnStatus(vpnStatus, vpnName, vpnHostname, vpnCity, vpnCountry, resp.VirtualLocation)
	return changed
}

func (ti *Instance) updateCountryList() bool {
	ti.state.connSelector.mu.RLock()
	oldCountryList := slices.Clone(ti.state.connSelector.countries)
	ti.state.connSelector.mu.RUnlock()

	newList, err := ti.state.connSelector.fetchCountries(ti.client)
	if err != nil {
		log.Println(logTag, internal.ErrorPrefix, "Error retrieving available country list:", err)
		return false
	}

	return !slices.Equal(oldCountryList, newList)
}

func (ti *Instance) updateSpecialtyServerList() bool {
	ti.state.mu.Lock()
	oldList := slices.Clone(ti.state.connSelector.specialtyServers)
	ti.state.mu.Unlock()

	newList, err := ti.state.connSelector.fetchSpecialtyServers(ti.client)
	if err != nil {
		log.Println(logTag, internal.ErrorPrefix, "Error retrieving available specialty server list:", err)
		return false
	}

	return !slices.Equal(oldList, newList)
}

func (ti *Instance) updateRecentConnections() bool {
	oldConnectionsList := slices.Clone(ti.recentConnections.GetRecentConnections())

	err := ti.recentConnections.UpdateRecentConnections()
	if err != nil {
		log.Println(logTag, internal.ErrorPrefix, "Error retrieving recent connections:", err)
		return false
	}

	newConnectionsList := ti.recentConnections.GetRecentConnections()
	return !slices.Equal(oldConnectionsList, newConnectionsList)
}

func (ti *Instance) setSettings(settings *pb.Settings) bool {
	if settings == nil {
		return false
	}

	userSettings := settings.UserSettings
	if userSettings == nil {
		return false
	}

	changed := false
	var notificationsText, trayText string
	var forceNotifications, forceTray bool
	func() {
		ti.state.mu.Lock()
		defer ti.state.mu.Unlock()

		var newNotificationsStatus Status
		if userSettings.Notify {
			newNotificationsStatus = Enabled
		} else {
			newNotificationsStatus = Disabled
		}

		if ti.state.notificationsStatus == Invalid {
			changed = true
			ti.state.notificationsStatus = newNotificationsStatus
		} else if ti.state.notificationsStatus != newNotificationsStatus {
			changed = true
			ti.state.notificationsStatus = newNotificationsStatus

			if newNotificationsStatus == Enabled {
				log.Println(internal.InfoPrefix, "Notifications for NordVPN turned on")
				notificationsText = labelNotificationsOn
				forceNotifications = true
			} else {
				log.Println(internal.InfoPrefix, "Notifications for NordVPN turned off")
				notificationsText = labelNotificationsOff
				forceNotifications = true
			}
		}

		var newTrayStatus Status
		if userSettings.Tray {
			newTrayStatus = Enabled
		} else {
			newTrayStatus = Disabled
		}

		if ti.state.trayStatus == Invalid {
			changed = true
			ti.state.trayStatus = newTrayStatus
		} else if ti.state.trayStatus != newTrayStatus {
			changed = true
			ti.state.trayStatus = newTrayStatus

			if newTrayStatus == Enabled {
				log.Println(internal.InfoPrefix, "Tray enabled")
				trayText = labelTrayOn
				forceTray = true
			} else {
				log.Println(internal.InfoPrefix, "Tray disabled")
				trayText = labelTrayOff
				forceTray = true
			}
		}
	}()

	if notificationsText != "" {
		notificationType := NoForce
		if forceNotifications {
			notificationType = Force
		}
		ti.notify(notificationType, notificationsText)
	}
	if trayText != "" {
		notificationType := NoForce
		if forceTray {
			notificationType = Force
		}
		ti.notify(notificationType, trayText)
	}

	return changed
}

func (ti *Instance) updateSettings() bool {
	const errorRetrievingSettingsLog = "Error retrieving settings:"

	resp, err := ti.client.Settings(context.Background(), &pb.Empty{})
	if err != nil {
		log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, err)
		return false
	}

	switch resp.Type {
	case internal.CodeConfigError:
		log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, client.ConfigMessage)
	case internal.CodeSuccess:
		return ti.setSettings(resp.Data)
	default:
		log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, internal.ErrUnhandled)
	}
	return false
}

func (ti *Instance) updateAccountInfo() bool {
	payload, err := ti.accountInfo.getAccountInfo(ti.client)
	if err != nil {
		if errMessage := messageForDaemonError(err); errMessage != internal.ErrDaemonConnectionRefused.Error() {
			ti.updateDaemonConnectionStatus(errMessage)
		}
		log.Println(internal.ErrorPrefix, "Error retrieving account info:", err)
		return true
	}

	changed := false
	vpnActive := ti.state.vpnActive
	var accountName string

	switch payload.Type {
	case internal.CodeUnauthorized:
		log.Println(internal.ErrorPrefix, cli.AccountTokenUnauthorizedError)
	case internal.CodeExpiredRenewToken:
		log.Println(internal.ErrorPrefix, "CodeExpiredRenewToken")
	case internal.CodeTokenRenewError:
		log.Println(internal.ErrorPrefix, "CodeTokenRenewError")
	}

	if payload.Username != "" {
		accountName = payload.Username
	} else {
		accountName = payload.Email
	}

	switch payload.Type {
	case internal.CodeSuccess:
		vpnActive = true
	case internal.CodeNoService:
		vpnActive = false
	}

	ti.state.mu.Lock()

	if ti.state.vpnActive != vpnActive {
		ti.state.vpnActive = vpnActive
		changed = true
	}

	if ti.state.accountName != accountName {
		ti.state.accountName = accountName
		changed = true
	}

	ti.state.mu.Unlock()
	return changed
}

func (ti *Instance) redraw(result bool) {
	if result {
		select {
		case ti.renderChan <- struct{}{}:
		default:
		}
	}
}

func messageForDaemonError(err error) string {
	if err == nil {
		return ""
	}

	statusError, ok := status.FromError(err)
	if !ok {
		return ""
	}
	errorMessage := statusError.Message()

	if errorMessage == internal.ErrNotLoggedIn.Error() {
		// no error needs to be displayed in this case
		// the user is not logged in and will be handled when fetching the login status
		return ""
	}

	if strings.Contains(errorMessage, "no such file or directory") {
		message := "NordVPN daemon is not running\n\n"
		if snapconf.IsUnderSnap() {
			message += "sudo snap start nordvpn"
		} else {
			message += "sudo systemctl enable --now nordvpnd"
		}
		return message
	}

	if strings.Contains(errorMessage, "permission denied") || strings.Contains(errorMessage, "connection reset by peer") {
		return "Add the user to the nordvpn group and reboot the system\n\nsudo usermod -aG nordvpn $USER"
	}

	if snapconf.IsUnderSnap() {
		if snapErr := cli.RetrieveSnapConnsError(err); snapErr != nil {
			return fmt.Sprintf(cli.MsgSnapPermissionsErrorForTray, cli.JoinSnapMissingPermissions(snapErr))
		}
	}

	return internal.ErrDaemonConnectionRefused.Error()
}

func (ti *Instance) updateDaemonConnectionStatus(errorMessage string) bool {
	daemonAvailable := false

	if (errorMessage == "") || (errorMessage == cli.ErrUpdateAvailable.Error()) {
		daemonAvailable = true
	}

	var notificationText string
	ti.state.mu.Lock()
	changed := false
	if ti.state.daemonAvailable != daemonAvailable {
		changed = true
		ti.state.daemonAvailable = daemonAvailable
		if daemonAvailable {
			notificationText = labelDaemonReconnected
		} else {
			notificationText = labelDaemonDisconnected
		}
	}

	if ti.state.daemonError != errorMessage {
		ti.state.daemonError = errorMessage
		changed = true
	}
	ti.state.mu.Unlock()

	if notificationText != "" {
		ti.notify(NoForce, notificationText)
	}
	return changed
}

func (ti *Instance) setVpnStatus(
	vpnStatus pb.ConnectionState,
	vpnName string,
	vpnHostname string,
	vpnCity string,
	vpnCountry string,
	virtualLocation bool,
) bool {
	changed := false
	var notificationText, notificationArg string
	func() {
		ti.state.mu.Lock()
		defer ti.state.mu.Unlock()

		oldVpnStatus := ti.state.vpnStatus
		oldVpnHostname := ti.state.vpnHostname
		oldServerName := ti.state.serverName()

		ti.state.vpnName = vpnName
		ti.state.vpnCity = vpnCity
		ti.state.vpnCountry = vpnCountry
		ti.state.vpnVirtualLocation = virtualLocation
		ti.state.vpnHostname = vpnHostname
		ti.state.vpnStatus = vpnStatus

		statusChanged := oldVpnStatus != vpnStatus
		hostnameChanged := oldVpnHostname != vpnHostname
		changed = statusChanged || hostnameChanged

		if statusChanged {
			log.Printf("%s VPN status changed from %s to %s\n", logTag, oldVpnStatus, vpnStatus)
			if vpnStatus == pb.ConnectionState_CONNECTED {
				notificationText = labelConnectedFormat
				notificationArg = ti.state.serverName()
			} else if vpnStatus == pb.ConnectionState_DISCONNECTED {
				if oldServerName != "" {
					notificationText = labelDisconnectedFormat
					notificationArg = oldServerName
				}
			}
		} else if hostnameChanged {
			log.Printf("%s VPN hostname changed from %s to %s\n", logTag, oldVpnHostname, vpnHostname)
			if vpnHostname != "" && oldVpnHostname != "" && vpnStatus == pb.ConnectionState_CONNECTED {
				notificationText = labelConnectedFormat
				notificationArg = ti.state.serverName()
			}
		}
	}()

	if notificationText != "" {
		ti.notify(NoForce, notificationText, notificationArg)
	}

	return changed
}
