package tray

import (
	"fmt"
	"os"
	"strings"
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
	client           pb.DaemonClient
	debugMode        bool
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

func NewTrayInstance(client pb.DaemonClient) *Instance {
	return &Instance{client: client}
}

func OnReady(ti *Instance) {
	if os.Getenv("NORDVPN_TRAY_DEBUG") == "1" {
		ti.debugMode = true
	} else {
		ti.debugMode = false
	}

	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")

	ti.iconConnected = "nordvpn-tray-blue"
	ti.iconDisconnected = "nordvpn-tray-white"

	currentDesktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	if strings.Contains(currentDesktop, "kde") {
		// TODO: Kubuntu uses dark tray background instead KDE default white
		ti.iconDisconnected = "nordvpn-tray-black"
	}
	if strings.Contains(currentDesktop, "mate") {
		ti.iconDisconnected = "nordvpn-tray-gray"
	}

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
			if ti.debugMode {
				addDebugSection(ti)
			}
			addQuitItem()
			systray.Refresh()
			<-ti.redrawChan
			if ti.debugMode {
				fmt.Println(time.Now().String(), "Redraw")
			}
			systray.ResetMenu()
		}
	}()
}
