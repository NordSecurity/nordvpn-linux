package tray

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"

	"github.com/NordSecurity/systray"
	"github.com/fatih/color"
	"google.golang.org/grpc/status"
)

type stateType struct {
	daemonAvailable bool
	loggedIn        bool
	vpnActive       bool
	meshnetEnabled  bool
	daemonError     string
	accountName     string
	vpnStatus       string
	vpnHostname     string
	vpnCity         string
	vpnCountry      string
	mu              sync.RWMutex
}

var state = stateType{vpnStatus: "Disconnected"}

// The pattern is to return 'true' if something has changed and 'false' when no changes were detected

func ping(client pb.DaemonClient) bool {
	changed := false
	daemonAvailable := false
	daemonError := ""

	resp, err := client.Ping(context.Background(), &pb.Empty{})
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

	state.mu.Lock()

	if !state.daemonAvailable && daemonAvailable {
		state.daemonAvailable = true
		changed = true
		if NotifyEnabled {
			defer notification("info", "Connected to NordVPN daemon")
		}
	} else if state.daemonAvailable && !daemonAvailable {
		state.daemonAvailable = false
		changed = true
		if NotifyEnabled {
			defer notification("info", "Disconnected from NordVPN daemon")
		}
	}

	if state.daemonError != daemonError {
		state.daemonError = daemonError
		changed = true
	}

	state.mu.Unlock()
	return changed
}

func fetchLogged(client pb.DaemonClient) bool {
	changed := false
	resp, err := client.IsLoggedIn(context.Background(), &pb.Empty{})
	loggedIn := err == nil && resp.GetValue()

	state.mu.Lock()

	if !state.loggedIn && loggedIn {
		state.loggedIn = true
		changed = true
		if NotifyEnabled {
			defer notification("info", "Logged in")
		}
	} else if state.loggedIn && !loggedIn {
		state.loggedIn = false
		changed = true
		if NotifyEnabled {
			defer notification("info", "Logged out")
		}
	}

	state.mu.Unlock()
	return changed
}

func fetchMeshnet(meshClient meshpb.MeshnetClient) bool {
	changed := false
	meshResp, err := meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	meshnetEnabled := err == nil && meshResp.GetValue()

	state.mu.Lock()

	if !state.meshnetEnabled && meshnetEnabled {
		state.meshnetEnabled = true
		changed = true
		if NotifyEnabled {
			defer notification("info", "Meshnet enabled")
		}
	} else if state.meshnetEnabled && !meshnetEnabled {
		state.meshnetEnabled = false
		changed = true
		if NotifyEnabled {
			defer notification("info", "Meshnet disabled")
		}
	}

	state.mu.Unlock()
	return changed
}

func fetchStatus(client pb.DaemonClient) bool {
	changed := false
	vpnStatus := ""
	vpnHostname := ""
	vpnCity := ""
	vpnCountry := ""
	resp, err := client.Status(context.Background(), &pb.Empty{})
	if err == nil {
		vpnStatus = resp.State
		vpnHostname = resp.Hostname
		vpnCity = resp.City
		vpnCountry = resp.Country
	}

	state.mu.Lock()

	if state.vpnStatus != vpnStatus {
		if vpnStatus == "Connected" {
			systray.SetIconName(iconConnected)
			if NotifyEnabled {
				defer notification("info", "Connected to VPN server: %s", vpnHostname)
			}
		} else {
			systray.SetIconName(iconDisconnected)
			if NotifyEnabled {
				defer notification("info", "Disconnected from VPN server")
			}
		}
		state.vpnStatus = vpnStatus
		changed = true
	}

	if state.vpnHostname != vpnHostname {
		state.vpnHostname = vpnHostname
		changed = true
	}

	state.vpnCity = vpnCity
	state.vpnCountry = vpnCountry

	state.mu.Unlock()
	return changed
}

func accountInfo(client pb.DaemonClient) bool {
	changed := false
	loggedIn := false
	vpnActive := false
	accountName := ""

	payload, err := client.AccountInfo(context.Background(), &pb.Empty{})
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

	state.mu.Lock()

	if state.loggedIn != loggedIn {
		state.loggedIn = loggedIn
		changed = true
	}

	if state.vpnActive != vpnActive {
		state.vpnActive = vpnActive
		changed = true
	}

	if state.accountName != accountName {
		state.accountName = accountName
		changed = true
	}

	state.mu.Unlock()
	return changed
}

func maybeRedraw(result bool, previous bool) bool {
	if result {
		redrawChan <- struct{}{}
	}
	return result || previous
}

func pollingMonitor(client pb.DaemonClient, meshClient meshpb.MeshnetClient, update <-chan bool, ticker <-chan time.Time) {
	fullUpdate := true
	fullUpdateLast := time.Time{}
	for {
		changed := false
		fullUpdate = maybeRedraw(ping(client), fullUpdate)
		if state.daemonAvailable {
			fullUpdate = maybeRedraw(fetchLogged(client), fullUpdate)
			if state.loggedIn {
				fullUpdate = maybeRedraw(fetchMeshnet(meshClient), fullUpdate)
				fullUpdate = maybeRedraw(fetchStatus(client), fullUpdate)
				if fullUpdate {
					changed = accountInfo(client)
					fullUpdateLast = time.Now()
				}
			}
		}

		if changed {
			redrawChan <- struct{}{}
		}
		select {
		case fullUpdate = <-update:
		case <-systray.TrayOpenedCh:
			fullUpdate = true
		case ts := <-ticker:
			if ts.Sub(fullUpdateLast) > PollingFullUpdateInterval {
				fullUpdate = true
			} else {
				fullUpdate = false
			}
		}
		if DebugMode {
			if fullUpdate {
				fmt.Println(time.Now().String(), "Full update")
			} else {
				fmt.Println(time.Now().String(), "Update")
			}
		}
	}
}
