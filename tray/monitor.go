package tray

import (
	"context"
	"fmt"
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
		defer ti.notify(pInfo, "Connected to NordVPN daemon")
	} else if ti.state.daemonAvailable && !daemonAvailable {
		ti.state.daemonAvailable = false
		changed = true
		defer ti.notify(pInfo, "Disconnected from NordVPN daemon")
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
		defer ti.notify(pInfo, "Logged in")
	} else if ti.state.loggedIn && !loggedIn {
		ti.state.loggedIn = false
		changed = true
		defer ti.notify(pInfo, "Logged out")
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
			defer ti.notify(pInfo, "Connected to VPN server: %s", vpnHostname)
		} else {
			systray.SetIconName(ti.iconDisconnected)
			defer ti.notify(pInfo, "Disconnected from VPN server")
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
	changed := false

	resp, err := ti.client.Settings(context.Background(), &pb.SettingsRequest{
		Uid: int64(os.Getuid()),
	})
	var settings *pb.Settings

	if err != nil {
		log(pError, "Error retrieving settings: %s", err)
	} else {
		switch resp.Type {
		case internal.CodeConfigError:
			log(pError, "Error retrieving settings: %s", client.ConfigMessage)
		case internal.CodeSuccess:
			settings = resp.GetData()
		default:
			log(pError, "Error retrieving settings: %s", internal.ErrUnhandled)
		}
	}

	if settings == nil {
		return false
	}

	ti.state.mu.Lock()
	if !ti.state.notifyEnabled && settings.Notify {
		ti.state.notifyEnabled = true
		changed = true
		defer ti.notify(pInfo, "Notifications enabled")
	}
	if ti.state.notifyEnabled && !settings.Notify {
		ti.state.notifyEnabled = false
		changed = true
		defer log(pInfo, "Notifications disabled")
	}
	ti.state.mu.Unlock()

	return changed
}

func (ti *Instance) updateAccountInfo() bool {
	changed := false
	loggedIn := false
	vpnActive := false
	accountName := ""

	payload, err := ti.client.AccountInfo(context.Background(), &pb.Empty{})
	if err != nil {
		if status.Convert(err).Message() != internal.ErrNotLoggedIn.Error() {
			log(pError, "Error retrieving account info: %s", err)
		}
	} else {
		switch payload.Type {
		case internal.CodeUnauthorized:
			log(pError, cli.AccountTokenUnauthorizedError)
		case internal.CodeExpiredRenewToken:
			log(pError, "CodeExpiredRenewToken")
		case internal.CodeTokenRenewError:
			log(pError, "CodeTokenRenewError")
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

func (ti *Instance) pollingMonitor(ticker <-chan time.Time) {
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
		case ts := <-ticker:
			if ts.Sub(fullUpdateLast) > PollingFullUpdateInterval {
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
