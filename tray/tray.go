package tray

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"

	"github.com/NordSecurity/systray"
)

const (
	NotifierStartDelay        = 3 * time.Second
	PollingUpdateInterval     = 1 * time.Second
	PollingFullUpdateInterval = 60 * time.Second
)

type Instance struct {
	Client           pb.DaemonClient
	DebugMode        bool
	notifier         dbusNotifier
	redrawChan       chan struct{}
	updateChan       chan bool
	iconConnected    string
	iconDisconnected string
	state            trayState
}

type trayState struct {
	daemonAvailable bool
	loggedIn        bool
	vpnActive       bool
	notifyEnabled   bool
	daemonError     string
	accountName     string
	vpnStatus       string
	vpnHostname     string
	vpnCity         string
	vpnCountry      string
	mu              sync.RWMutex
}

func OnReady(ti *Instance) {
	if os.Getenv("NORDVPN_TRAY_DEBUG") == "1" {
		ti.DebugMode = true
	} else {
		ti.DebugMode = false
	}

	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")

	ti.iconConnected = "nordvpn-tray-blue"
	ti.iconDisconnected = "nordvpn-tray-white"

	// TODO: Detect running DE and set iconDisconnected to "nordvpn-tray-black" on KDE/Plasma,
	// and to "nordvpn-tray-gray" on before-Gnome Ubuntu versions

	systray.SetIconName(ti.iconDisconnected)

	ti.state.vpnStatus = "Disconnected"
	ti.state.notifyEnabled = true
	ti.redrawChan = make(chan struct{})
	ti.updateChan = make(chan bool)

	time.AfterFunc(NotifierStartDelay, func() { ti.notifier.start() })

	ticker := time.Tick(PollingUpdateInterval)
	go ti.pollingMonitor(ticker)

	go func() {
		for {
			ti.state.mu.RLock()
			addAppSection()
			if ti.state.daemonAvailable {
				if ti.state.loggedIn {
					addVpnSection(ti)
				}
				addAccountSection(ti)
			} else {
				addDaemonSection(ti)
			}
			ti.state.mu.RUnlock()
			if ti.DebugMode {
				addDebugSection(ti)
			}
			addQuitItem()
			systray.Refresh()
			<-ti.redrawChan
			if ti.DebugMode {
				fmt.Println(time.Now().String(), "Redraw")
			}
			systray.ResetMenu()
		}
	}()
}
