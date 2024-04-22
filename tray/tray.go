package tray

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/snapconf"

	"github.com/NordSecurity/systray"
)

const (
	NotifierStartDelay        = 3 * time.Second
	PollingUpdateInterval     = 5 * time.Second
	PollingFullUpdateInterval = 60 * time.Second
	AccountInfoUpdateInterval = 24 * time.Hour
)

type accountInfo struct {
	accountInfo *pb.AccountResponse
	updateTime  time.Time
}

// getAccountInfo use cache to not query API every time
func (ai *accountInfo) getAccountInfo(client pb.DaemonClient) (*pb.AccountResponse, error) {
	if time.Since(ai.updateTime) > AccountInfoUpdateInterval {
		var err error
		ai.accountInfo, err = client.AccountInfo(context.Background(), &pb.Empty{})
		if err != nil {
			return &pb.AccountResponse{}, fmt.Errorf("retrvieving account info: %w", err)
		}
		ai.updateTime = time.Now()
	}
	return ai.accountInfo, nil
}

type Instance struct {
	client           pb.DaemonClient
	accountInfo      accountInfo
	debugMode        bool
	notifier         dbusNotifier
	redrawChan       chan struct{}
	updateChan       chan bool
	iconConnected    string
	iconDisconnected string
	state            trayState
	quitChan         chan<- norduser.StopRequest
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

func NewTrayInstance(client pb.DaemonClient, quitChan chan<- norduser.StopRequest) *Instance {
	return &Instance{client: client, quitChan: quitChan}
}

func getIconPath(name string) string {
	const iconPath = "/usr/share/icons/hicolor/scalable/apps"
	if snapconf.IsUnderSnap() {
		return internal.PrefixStaticPath(path.Join(iconPath, name+".svg"))
	}

	return name
}

func OnReady(ti *Instance) {
	if os.Getenv("NORDVPN_TRAY_DEBUG") == "1" {
		ti.debugMode = true
	} else {
		ti.debugMode = false
	}

	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")

	ti.iconConnected = getIconPath("nordvpn-tray-blue")
	ti.iconDisconnected = getIconPath("nordvpn-tray-white")

	currentDesktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	if strings.Contains(currentDesktop, "kde") {
		// TODO: Kubuntu uses dark tray background instead KDE default white
		ti.iconDisconnected = getIconPath("nordvpn-tray-black")
	}
	if strings.Contains(currentDesktop, "mate") {
		ti.iconDisconnected = getIconPath("nordvpn-tray-gray")
	}

	systray.SetIconName(ti.iconDisconnected)
	ti.state.vpnStatus = "Disconnected"
	ti.state.notifyEnabled = false
	ti.redrawChan = make(chan struct{})
	ti.updateChan = make(chan bool)

	time.AfterFunc(NotifierStartDelay, func() { ti.notifier.start() })

	go ti.pollingMonitor()

	go func() {
		for {
			ti.state.mu.RLock()
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
			addQuitItem(ti)
			systray.Refresh()
			<-ti.redrawChan
			if ti.debugMode {
				fmt.Println(time.Now().String(), "Redraw")
			}
			systray.ResetMenu()
		}
	}()
}
