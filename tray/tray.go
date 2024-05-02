package tray

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/notify"

	"github.com/NordSecurity/systray"
)

const (
	NotifierStartDelay        = 3 * time.Second
	PollingUpdateInterval     = 5 * time.Second
	PollingFullUpdateInterval = 60 * time.Second
	AccountInfoUpdateInterval = 24 * time.Hour
	ConnectedString           = "Connected"
)

type Status int

const (
	Invalid Status = iota
	Enabled
	Disabled
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
			return nil, err
		}
		ai.updateTime = time.Now()
	}
	return ai.accountInfo, nil
}

func (ai *accountInfo) reset() {
	ai.updateTime = time.Time{}
	ai.accountInfo = nil
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
	daemonAvailable     bool
	loggedIn            bool
	vpnActive           bool
	notificationsStatus Status
	daemonError         string
	accountName         string
	vpnStatus           string
	vpnName             string
	vpnHostname         string
	vpnCity             string
	vpnCountry          string
	mu                  sync.RWMutex
}

func NewTrayInstance(client pb.DaemonClient, quitChan chan<- norduser.StopRequest) *Instance {
	return &Instance{client: client, quitChan: quitChan}
}

func OnReady(ti *Instance) {
	if os.Getenv("NORDVPN_TRAY_DEBUG") == "1" {
		ti.debugMode = true
	} else {
		ti.debugMode = false
	}

	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")

	ti.iconConnected = notify.GetIconPath("nordvpn-tray-blue")
	ti.iconDisconnected = notify.GetIconPath("nordvpn-tray-white")

	currentDesktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	if strings.Contains(currentDesktop, "kde") {
		// TODO: Kubuntu uses dark tray background instead KDE default white
		ti.iconDisconnected = notify.GetIconPath("nordvpn-tray-black")
	}
	if strings.Contains(currentDesktop, "mate") {
		ti.iconDisconnected = notify.GetIconPath("nordvpn-tray-gray")
	}

	systray.SetIconName(ti.iconDisconnected)
	ti.state.vpnStatus = "Disconnected"
	ti.state.notificationsStatus = Invalid
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
				log.Println(time.Now().String(), "Redraw")
			}
			systray.ResetMenu()
		}
	}()
}
