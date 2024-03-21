package tray

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"

	"github.com/NordSecurity/systray"
	"github.com/fatih/color"
	"google.golang.org/grpc/status"
)

// The pattern is to return 'true' if something has changed and 'false' when no changes were detected

func (ti *Instance) ping() bool {
	changed := false
	daemonAvailable := false
	daemonError := ""

	resp, err := ti.Client.Ping(context.Background(), &pb.Empty{})
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
		if ti.NotifyEnabled {
			defer ti.notification("info", "Connected to NordVPN daemon")
		}
	} else if ti.state.daemonAvailable && !daemonAvailable {
		ti.state.daemonAvailable = false
		changed = true
		if ti.NotifyEnabled {
			defer ti.notification("info", "Disconnected from NordVPN daemon")
		}
	}

	if ti.state.daemonError != daemonError {
		ti.state.daemonError = daemonError
		changed = true
	}

	ti.state.mu.Unlock()
	return changed
}

func (ti *Instance) fetchLogged() bool {
	changed := false
	resp, err := ti.Client.IsLoggedIn(context.Background(), &pb.Empty{})
	loggedIn := err == nil && resp.GetValue()

	ti.state.mu.Lock()

	if !ti.state.loggedIn && loggedIn {
		ti.state.loggedIn = true
		changed = true
		if ti.NotifyEnabled {
			defer ti.notification("info", "Logged in")
		}
	} else if ti.state.loggedIn && !loggedIn {
		ti.state.loggedIn = false
		changed = true
		if ti.NotifyEnabled {
			defer ti.notification("info", "Logged out")
		}
	}

	ti.state.mu.Unlock()
	return changed
}

func (ti *Instance) fetchMeshnet() bool {
	changed := false
	meshResp, err := ti.MeshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	meshnetEnabled := err == nil && meshResp.GetValue()

	ti.state.mu.Lock()

	if !ti.state.meshnetEnabled && meshnetEnabled {
		ti.state.meshnetEnabled = true
		changed = true
		if ti.NotifyEnabled {
			defer ti.notification("info", "Meshnet enabled")
		}
	} else if ti.state.meshnetEnabled && !meshnetEnabled {
		ti.state.meshnetEnabled = false
		changed = true
		if ti.NotifyEnabled {
			defer ti.notification("info", "Meshnet disabled")
		}
	}

	ti.state.mu.Unlock()
	return changed
}

func (ti *Instance) fetchStatus() bool {
	changed := false
	vpnStatus := ""
	vpnHostname := ""
	vpnCity := ""
	vpnCountry := ""
	resp, err := ti.Client.Status(context.Background(), &pb.Empty{})
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
			if ti.NotifyEnabled {
				defer ti.notification("info", "Connected to VPN server: %s", vpnHostname)
			}
		} else {
			systray.SetIconName(ti.iconDisconnected)
			if ti.NotifyEnabled {
				defer ti.notification("info", "Disconnected from VPN server")
			}
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

func (ti *Instance) accountInfo() bool {
	changed := false
	loggedIn := false
	vpnActive := false
	accountName := ""

	payload, err := ti.Client.AccountInfo(context.Background(), &pb.Empty{})
	if err != nil {
		if status.Convert(err).Message() != internal.ErrNotLoggedIn.Error() {
			color.Red("Error retrieving account info: %s", err)
		}
	} else {
		switch payload.Type {
		case internal.CodeUnauthorized:
			color.Red(cli.AccountTokenUnauthorizedError)
		case internal.CodeExpiredRenewToken:
			color.Red("CodeExpiredRenewToken")
		case internal.CodeTokenRenewError:
			color.Red("CodeTokenRenewError")
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
			fullUpdate = ti.maybeRedraw(ti.fetchLogged(), fullUpdate)
			if ti.state.loggedIn {
				fullUpdate = ti.maybeRedraw(ti.fetchMeshnet(), fullUpdate)
				fullUpdate = ti.maybeRedraw(ti.fetchStatus(), fullUpdate)
				if fullUpdate {
					changed = ti.accountInfo()
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
		if ti.DebugMode {
			if fullUpdate {
				fmt.Println(time.Now().String(), "Full update")
			} else {
				fmt.Println(time.Now().String(), "Update")
			}
		}
	}
}
