package tray

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"

	"github.com/NordSecurity/systray"
)

const (
	NotifierStartDelay        = 3 * time.Second
	PollingUpdateInterval     = 1 * time.Second
	PollingFullUpdateInterval = 60 * time.Second
)

var (
	Client           pb.DaemonClient
	MeshClient       meshpb.MeshnetClient
	FileshareClient  filesharepb.FileshareClient
	NotifyEnabled    bool
	DebugMode        bool
	notifier         dbusNotifier
	redrawChan       chan struct{}
	updateChan       chan bool
	iconConnected    = "nordvpn-tray-blue"
	iconDisconnected = "nordvpn-tray-white"
)

func OnReady() {
	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")
	systray.SetIconName(iconDisconnected)

	// TODO: Detect running DE and set iconDisconnected to "nordvpn-tray-black" on KDE/Plasma,
	// and to "nordvpn-tray-gray" on before-Gnome Ubuntu versions

	redrawChan = make(chan struct{})
	updateChan = make(chan bool)

	time.AfterFunc(NotifierStartDelay, func() { notifier.start() })

	ticker := time.Tick(PollingUpdateInterval)
	go pollingMonitor(Client, MeshClient, updateChan, ticker)

	go func() {
		for {
			state.mu.RLock()
			addAppSection()
			if state.daemonAvailable {
				if state.loggedIn {
					addVpnSection()
					// Disabled for now: addMeshnetSection()
				}
				addAccountSection()
			} else {
				addDaemonSection()
			}
			state.mu.RUnlock()
			if DebugMode {
				addDebugSection()
			}
			addQuitItem()
			systray.Refresh()
			<-redrawChan
			if DebugMode {
				fmt.Println(time.Now().String(), "Redraw")
			}
			systray.ResetMenu()
		}
	}()
}
