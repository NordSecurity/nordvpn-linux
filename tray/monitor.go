package tray

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/snapconf"

	"github.com/NordSecurity/systray"
	"google.golang.org/grpc/status"
)

// The pattern is to return 'true' if something has changed and 'false' when no changes were detected
func (ti *Instance) ping() bool {
	daemonError := ""

	resp, err := ti.client.Ping(context.Background(), &pb.Empty{})
	if err != nil {
		daemonError = messageForDaemonError(err)
	} else {
		switch resp.Type {
		case internal.CodeOffline:
			daemonError = cli.ErrInternetConnection.Error()
		case internal.CodeDaemonOffline:
			daemonError = internal.ErrDaemonConnectionRefused.Error()
		case internal.CodeOutdated:
			daemonError = cli.ErrUpdateAvailable.Error()
		}
	}

	return ti.updateDaemonConnectionStatus(daemonError)
}

func (ti *Instance) updateLoginStatus() bool {
	changed := false
	resp, err := ti.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil {
		return ti.updateDaemonConnectionStatus(messageForDaemonError(err))
	}

	loggedIn := resp.GetIsLoggedIn()

	if !loggedIn && ti.state.loggedIn && ti.state.vpnStatus == pb.ConnectionState_CONNECTED {
		// reset the VPN info if the user logs out while connected to VPN
		ti.setVpnStatus(pb.ConnectionState_DISCONNECTED, "", "", "", "", false)
	}

	ti.state.mu.Lock()

	if !ti.state.loggedIn && loggedIn {
		ti.state.loggedIn = true
		changed = true
		defer ti.notify("You've successfully logged in")
	} else if ti.state.loggedIn && !loggedIn {
		ti.state.loggedIn = false
		ti.accountInfo.reset()
		ti.state.accountName = ""
		changed = true
		defer ti.notify("You've logged out")
	}

	ti.state.mu.Unlock()
	return changed
}

func (ti *Instance) updateVpnStatus() bool {
	changed := false
	resp, err := ti.client.Status(context.Background(), &pb.Empty{})
	if err != nil {
		return ti.updateDaemonConnectionStatus(messageForDaemonError(err))
	}

	vpnStatus := resp.State
	vpnHostname := resp.Hostname
	vpnCity := resp.City
	vpnCountry := resp.Country
	vpnName := resp.Name
	if vpnName == "" {
		vpnName = vpnHostname
	}

	shouldDisplayNotification := (ti.state.vpnStatus != vpnStatus) || (ti.state.vpnHostname != vpnHostname)

	if shouldDisplayNotification {
		// update daemon settings before notifications are shown
		changed = ti.updateAccountInfo()
		changed = ti.updateSettings() || changed
	}

	return ti.setVpnStatus(vpnStatus, vpnName, vpnHostname, vpnCity, vpnCountry, resp.VirtualLocation) || changed
}

func (ti *Instance) updateSettings() bool {
	const errorRetrievingSettingsLog = "Error retrieving settings:"
	changed := false

	resp, err := ti.client.Settings(context.Background(), &pb.Empty{})
	var settings *pb.UserSpecificSettings

	if err != nil {
		log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, err)
	} else {
		switch resp.Type {
		case internal.CodeConfigError:
			log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, client.ConfigMessage)
		case internal.CodeSuccess:
			settings = resp.Data.UserSettings
		default:
			log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, internal.ErrUnhandled)
		}
	}

	if settings == nil {
		return false
	}

	ti.state.mu.Lock()

	var newNotificationsStatus Status
	if settings.Notify {
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
			defer ti.notifyForce("Notifications for NordVPN turned on")
			defer log.Println(internal.InfoPrefix, "Notifications for NordVPN turned on")
		} else {
			defer ti.notifyForce("Notifications for NordVPN turned off")
			defer log.Println(internal.InfoPrefix, "Notifications for NordVPN turned off")
		}
	}

	var newTrayStatus Status
	if settings.Tray {
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
			defer log.Println(internal.InfoPrefix, "Tray enabled")
		} else {
			defer log.Println(internal.InfoPrefix, "Tray disabled")
		}
	}

	ti.state.mu.Unlock()

	return changed
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
		case ti.redrawChan <- struct{}{}:
		default:
		}
	}
}

func (ti *Instance) pollingMonitor() {
	initialChan := ti.initialChan
	ticker := time.NewTicker(PollingUpdateInterval)
	defer ticker.Stop()

	fullUpdate := true
	fullUpdateLast := time.Time{}
	for {
		ti.redraw(ti.ping())
		if ti.state.daemonAvailable {
			ti.redraw(ti.updateLoginStatus())
			ti.redraw(ti.updateSettings())
			if ti.state.loggedIn {
				if fullUpdate {
					ti.redraw(ti.updateAccountInfo())
				}
				ti.redraw(ti.updateVpnStatus())
				if fullUpdate {
					fullUpdateLast = time.Now()
				}
			}
		}

		// while the settings were not fetch don't unblock the tray loop
		if ti.state.trayStatus != Invalid && initialChan != nil {
			initialChan <- struct{}{}
			close(initialChan)
			initialChan = nil
			if ti.debugMode {
				log.Println(internal.DebugPrefix, "Initial retrieve")
			}
		}

		select {
		case fullUpdate = <-ti.updateChan:
		case <-systray.TrayOpenedCh:
			fullUpdate = true
		case ts := <-ticker.C:
			fullUpdate = ts.Sub(fullUpdateLast) >= PollingFullUpdateInterval
		}
		if ti.debugMode {
			if fullUpdate {
				log.Println(internal.DebugPrefix, "Full update")
			} else {
				log.Println(internal.DebugPrefix, "Update")
			}
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
	changed := false
	daemonAvailable := false

	if (errorMessage == "") || (errorMessage == cli.ErrUpdateAvailable.Error()) {
		daemonAvailable = true
	}

	ti.state.mu.Lock()

	if ti.state.daemonAvailable != daemonAvailable {
		changed = true
		ti.state.daemonAvailable = daemonAvailable
		if daemonAvailable {
			defer ti.notify("Reconnected to NordVPN's background service")
		} else {
			defer ti.notify("Couldn't connect to NordVPN's background service. Please ensure the service is running.")
		}
	}

	if ti.state.daemonError != errorMessage {
		ti.state.daemonError = errorMessage
		changed = true
	}

	ti.state.mu.Unlock()
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
	ti.state.mu.Lock()

	notifyConnected := func() {
		// use this helper function to ensure that the connected notification is displaying the latest info from ti.state on defer
		ti.notify("Connected to %s", ti.state.serverName())
	}

	if ti.state.vpnStatus != vpnStatus {
		//exhaustive:ignore
		switch vpnStatus {
		case pb.ConnectionState_CONNECTED:
			if ti.state.systrayRunning {
				systray.SetIconName(ti.iconConnected)
			}
			defer notifyConnected()
		case pb.ConnectionState_DISCONNECTED:
			if ti.state.systrayRunning {
				systray.SetIconName(ti.iconDisconnected)
			}
			// when connection attempt is cancelled, we end up in "Disconnected"
			// state, but we were not connected to anything at this point,so
			// ignore the notification
			if ti.state.serverName() != "" {
				defer ti.notify(fmt.Sprintf("Disconnected from %s", ti.state.serverName()))
			}
		}
		ti.state.vpnStatus = vpnStatus
		changed = true
	}

	if ti.state.vpnHostname != vpnHostname {
		if ti.state.vpnHostname != "" && vpnHostname != "" {
			defer notifyConnected()
		}
		ti.state.vpnHostname = vpnHostname
		changed = true
	}

	ti.state.vpnName = vpnName
	ti.state.vpnCity = vpnCity
	ti.state.vpnCountry = vpnCountry
	ti.state.vpnVirtualLocation = virtualLocation

	ti.state.mu.Unlock()
	return changed
}
