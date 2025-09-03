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

const (
	// Notification messages
	labelLoginSuccess       = "You've successfully logged in"
	labelLogout             = "You've logged out"
	labelNotificationsOn    = "Notifications for NordVPN turned on"
	labelNotificationsOff   = "Notifications for NordVPN turned off"
	labelDaemonReconnected  = "Reconnected to NordVPN's background service"
	labelDaemonDisconnected = "Couldn't connect to NordVPN's background service. Please ensure the service is running."
	labelConnectedFormat    = "Connected to %s"
	labelDisconnectedFormat = "Disconnected from %s"
)

// The pattern is to return 'true' if something has changed and 'false' when no changes were detected
func (ti *Instance) ping() error {
	daemonError := ""

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := ti.client.Ping(ctx, &pb.Empty{})
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

	ti.redraw(ti.updateDaemonConnectionStatus(daemonError))
	if daemonError != "" {
		return fmt.Errorf(daemonError)
	}

	return nil
}

func (ti *Instance) update() {
	ti.updateSettings()
	ti.updateVpnStatus()
	ti.updateCountryList()
	ti.updateAccountInfo()
	ti.updateLoginStatus()
}

func (ti *Instance) updateLoginStatus() {
	changed := false
	resp, err := ti.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil {
		ti.redraw(ti.updateDaemonConnectionStatus(messageForDaemonError(err)))
		return
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
		defer ti.notify(labelLoginSuccess)
	} else if ti.state.loggedIn && !loggedIn {
		ti.state.loggedIn = false
		ti.accountInfo.reset()
		ti.state.accountName = ""
		changed = true
		defer ti.notify(labelLogout)
	}

	ti.state.mu.Unlock()
	ti.redraw(changed)
}

func (ti *Instance) updateVpnStatus() {
	changed := false
	resp, err := ti.client.Status(context.Background(), &pb.Empty{})
	if err != nil {
		ti.redraw(ti.updateDaemonConnectionStatus(messageForDaemonError(err)))
		return
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
		ti.updateAccountInfo()
		ti.updateSettings()
	}

	ti.redraw(ti.setVpnStatus(vpnStatus, vpnName, vpnHostname, vpnCity, vpnCountry, resp.VirtualLocation) || changed)
}

func (ti *Instance) updateCountryList() {
	ti.state.mu.Lock()
	oldCountryList := append([]string(nil), ti.state.connSelector.countries...)
	ti.state.mu.Unlock()

	newList, err := ti.state.connSelector.listCountries(ti.client)
	if err != nil {
		log.Println(internal.ErrorPrefix, "Error retrieving available country list:", err)
		return
	}

	if len(oldCountryList) != len(newList) {
		ti.redraw(true)
		return
	}

	for i := range oldCountryList {
		if oldCountryList[i] != newList[i] {
			ti.redraw(true)
			return
		}
	}
}

func (ti *Instance) setSettings(settings *pb.Settings) {
	if settings == nil {
		return
	}

	userSettings := settings.UserSettings
	if userSettings == nil {
		return
	}

	changed := false
	ti.state.mu.Lock()

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
			defer ti.notifyForce(labelNotificationsOn)
			defer log.Println(internal.InfoPrefix, "Notifications for NordVPN turned on")
		} else {
			defer ti.notifyForce(labelNotificationsOff)
			defer log.Println(internal.InfoPrefix, "Notifications for NordVPN turned off")
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
			defer log.Println(internal.InfoPrefix, "Tray enabled")
		} else {
			defer log.Println(internal.InfoPrefix, "Tray disabled")
		}
	}

	ti.state.mu.Unlock()
	ti.redraw(changed)
}

func (ti *Instance) updateSettings() {
	const errorRetrievingSettingsLog = "Error retrieving settings:"

	resp, err := ti.client.Settings(context.Background(), &pb.Empty{})
	if err != nil {
		log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, err)
	} else {
		switch resp.Type {
		case internal.CodeConfigError:
			log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, client.ConfigMessage)
		case internal.CodeSuccess:
			ti.setSettings(resp.Data)
		default:
			log.Println(internal.ErrorPrefix, errorRetrievingSettingsLog, internal.ErrUnhandled)
		}
	}
}

func (ti *Instance) updateAccountInfo() {
	payload, err := ti.accountInfo.getAccountInfo(ti.client)
	if err != nil {
		if errMessage := messageForDaemonError(err); errMessage != internal.ErrDaemonConnectionRefused.Error() {
			ti.updateDaemonConnectionStatus(errMessage)
		}
		log.Println(internal.ErrorPrefix, "Error retrieving account info:", err)
		ti.redraw(true)
		return
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
	ti.redraw(changed)
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
			defer ti.notify(labelDaemonReconnected)
		} else {
			defer ti.notify(labelDaemonDisconnected)
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
		ti.notify(labelConnectedFormat, ti.state.serverName())
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
				defer ti.notify(labelDisconnectedFormat, ti.state.serverName())
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
