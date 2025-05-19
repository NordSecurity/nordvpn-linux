package tray

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

type TrayState struct {
	systrayRunning      bool
	daemonAvailable     bool
	loggedIn            bool
	vpnActive           bool
	notificationsStatus Status
	trayStatus          Status
	daemonError         string
	accountName         string
	vpnStatus           pb.ConnectionState
	vpnName             string
	vpnHostname         string
	vpnCity             string
	vpnCountry          string
	vpnVirtualLocation  bool
	wasConsentGiven     bool
	mu                  sync.RWMutex
}

// XXX: add docs
func (state *TrayState) UpdateConsent(wasConsentGiven bool) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.wasConsentGiven = wasConsentGiven
}

// Not thread safe. Lock mu before using
func (state *TrayState) serverName() string {
	vpnServerName := state.vpnName
	if vpnServerName == "" {
		vpnServerName = state.vpnHostname
	}
	if vpnServerName != "" {
		if state.vpnVirtualLocation {
			vpnServerName += " - Virtual"
		}
	}
	return vpnServerName
}
