package tray

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/NordSecurity/systray"
	"google.golang.org/grpc/status"
)

// The pattern is to return 'true' if something has changed and 'false' when no changes were detected

func (ti *Instance) ping() bool {
	changed := false
	daemonAvailable := false
	daemonError := ""

	resp, err := ti.client.Ping(context.Background(), &pb.Empty{})
	if err != nil {
		daemonError = internal.ErrDaemonConnectionRefused.Error()
		if strings.Contains(err.Error(), "no such file or directory") {
			daemonError = "nordvpnd is not running"
		}
		if strings.Contains(err.Error(), "permission denied") {
			daemonError = "add a user to the nordvpn group"
		}
		if snapErr := cli.RetrieveSnapConnsError(err); snapErr != nil {
			daemonError = cli.FormatSnapMissingConnsErr(snapErr)
		}
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

	if daemonError == "" {
		daemonAvailable = true
	}

	ti.state.mu.Lock()

	if !ti.state.daemonAvailable && daemonAvailable {
		ti.state.daemonAvailable = true
		changed = true
		defer ti.notify("Connected to NordVPN daemon")
	} else if ti.state.daemonAvailable && !daemonAvailable {
		ti.state.daemonAvailable = false
		changed = true
		defer ti.notify("Disconnected from NordVPN daemon")
	}

	if ti.state.daemonError != daemonError {
		ti.state.daemonError = daemonError
		changed = true
	}

	ti.state.mu.Unlock()
	return changed
}

func (ti *Instance) updateLoginStatus() bool {
	changed := false
	resp, err := ti.client.IsLoggedIn(context.Background(), &pb.Empty{})
	loggedIn := err == nil && resp.GetValue()

	ti.state.mu.Lock()

	if !ti.state.loggedIn && loggedIn {
		ti.state.loggedIn = true
		changed = true
		defer ti.notify("Logged in")
	} else if ti.state.loggedIn && !loggedIn {
		ti.state.loggedIn = false
		changed = true
		defer ti.notify("Logged out")
	}

	ti.state.mu.Unlock()
	return changed
}

func (ti *Instance) updateVpnStatus() bool {
	changed := false
	vpnStatus := ""
	vpnHostname := ""
	vpnCity := ""
	vpnCountry := ""
	resp, err := ti.client.Status(context.Background(), &pb.Empty{})
	if err == nil {
		vpnStatus = resp.State
		vpnHostname = resp.Hostname
		vpnCity = resp.City
		vpnCountry = resp.Country
	}

	ti.state.mu.Lock()

	if ti.state.vpnStatus != vpnStatus {
		if vpnStatus == "Connected" {
			systray.SetIconName(ti.iconConnected)
			defer ti.notify("Connected to VPN server: %s", vpnHostname)
		} else {
			systray.SetIconName(ti.iconDisconnected)
			defer ti.notify("Disconnected from VPN server")
		}
		ti.state.vpnStatus = vpnStatus
		changed = true
	}

	if ti.state.vpnHostname != vpnHostname {
		ti.state.vpnHostname = vpnHostname
		changed = true
	}

	ti.state.vpnCity = vpnCity
	ti.state.vpnCountry = vpnCountry

	ti.state.mu.Unlock()
	return changed
}

func (ti *Instance) updateSettings() bool {
	const errorRetrievingSettingsLog = "Error retrieving settings:"
	changed := false

	resp, err := ti.client.Settings(context.Background(), &pb.SettingsRequest{
		Uid: int64(os.Getuid()),
	})
	var settings *pb.Settings

	if err != nil {
		log.Println(internal.ErrorPrefix+errorRetrievingSettingsLog, err)
	} else {
		switch resp.Type {
		case internal.CodeConfigError:
			log.Println(internal.ErrorPrefix+errorRetrievingSettingsLog, client.ConfigMessage)
		case internal.CodeSuccess:
			settings = resp.GetData()
		default:
			log.Println(internal.ErrorPrefix+errorRetrievingSettingsLog, internal.ErrUnhandled)
		}
	}

	if settings == nil {
		return false
	}

	ti.state.mu.Lock()
	if !ti.state.notifyEnabled && settings.Notify {
		ti.state.notifyEnabled = true
		changed = true
		defer ti.notify("Notifications enabled")
	}
	if ti.state.notifyEnabled && !settings.Notify {
		ti.state.notifyEnabled = false
		changed = true
		defer log.Println(internal.InfoPrefix + " Notifications disabled")
	}
	ti.state.mu.Unlock()

	return changed
}

func (ti *Instance) updateAccountInfo() bool {
	changed := false
	loggedIn := ti.state.loggedIn
	vpnActive := ti.state.vpnActive
	accountName := ti.state.accountName

	payload, err := ti.accountInfo.getAccountInfo(ti.client)
	if err != nil {
		if status.Convert(err).Message() != internal.ErrNotLoggedIn.Error() {
			log.Println(internal.ErrorPrefix+" Error retrieving account info: ", err)
			return false
		}
		loggedIn = false
	} else {
		switch payload.Type {
		case internal.CodeUnauthorized:
			log.Println(internal.ErrorPrefix + " " + cli.AccountTokenUnauthorizedError)
		case internal.CodeExpiredRenewToken:
			log.Println(internal.ErrorPrefix + " CodeExpiredRenewToken")
		case internal.CodeTokenRenewError:
			log.Println(internal.ErrorPrefix + " CodeTokenRenewError")
		default:
			loggedIn = true
		}

		if payload.Username != "" {
			accountName = payload.Username
		} else {
			accountName = payload.Email
		}

		switch payload.Type {
		case internal.CodeSuccess:
			vpnActive = true
		case internal.CodeNoVPNService:
			vpnActive = false
		}
	}

	ti.state.mu.Lock()

	if ti.state.loggedIn != loggedIn {
		ti.state.loggedIn = loggedIn
		changed = true
	}

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

func (ti *Instance) maybeRedraw(result bool, previous bool) bool {
	if result {
		ti.redrawChan <- struct{}{}
	}
	return result || previous
}

func (ti *Instance) pollingMonitor() {
	ticker := time.NewTicker(PollingUpdateInterval)
	defer ticker.Stop()

	fullUpdate := true
	fullUpdateLast := time.Time{}
	for {
		changed := false
		fullUpdate = ti.maybeRedraw(ti.ping(), fullUpdate)
		if ti.state.daemonAvailable {
			fullUpdate = ti.maybeRedraw(ti.updateLoginStatus(), fullUpdate)
			if ti.state.loggedIn {
				fullUpdate = ti.maybeRedraw(ti.updateVpnStatus(), fullUpdate)
				if fullUpdate {
					changed = ti.updateAccountInfo()
					changed = ti.updateSettings() || changed
					fullUpdateLast = time.Now()
				}
			}
		}

		if changed {
			ti.redrawChan <- struct{}{}
		}
		select {
		case fullUpdate = <-ti.updateChan:
		case <-systray.TrayOpenedCh:
			fullUpdate = true
		case ts := <-ticker.C:
			if ts.Sub(fullUpdateLast) >= PollingFullUpdateInterval {
				fullUpdate = true
			} else {
				fullUpdate = false
			}
		}
		if ti.debugMode {
			if fullUpdate {
				fmt.Println(time.Now().String(), "Full update")
			} else {
				fmt.Println(time.Now().String(), "Update")
			}
		}
	}
}
