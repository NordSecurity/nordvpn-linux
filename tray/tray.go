package tray

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/notify"

	"github.com/NordSecurity/systray"
)

const (
	NotifierStartDelay        = 3 * time.Second
	PollingUpdateInterval     = 5 * time.Second
	PollingFullUpdateInterval = 60 * time.Second
	AccountInfoUpdateInterval = 24 * time.Hour
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
	fileshareClient  filesharepb.FileshareClient
	accountInfo      accountInfo
	debugMode        bool
	notifier         dbusNotifier
	redrawChan       chan struct{}
	initialChan      chan struct{}
	updateChan       chan bool
	iconConnected    string
	iconDisconnected string
	state            trayState
	quitChan         chan<- norduser.StopRequest
}

type trayState struct {
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
	mu                  sync.RWMutex
}

// Not thread safe. Lock mu before using
func (state *trayState) serverName() string {
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

func NewTrayInstance(client pb.DaemonClient, fileshareClient filesharepb.FileshareClient, quitChan chan<- norduser.StopRequest) *Instance {
	return &Instance{client: client, fileshareClient: fileshareClient, quitChan: quitChan}
}

func (ti *Instance) WaitInitialTrayStatus() Status {
	<-ti.initialChan
	ti.state.mu.RLock()
	defer ti.state.mu.RUnlock()
	return ti.state.trayStatus
}

func (ti *Instance) Start() {
	if os.Getenv("NORDVPN_TRAY_DEBUG") == "1" {
		ti.debugMode = true
	} else {
		ti.debugMode = false
	}

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

	ti.state.vpnStatus = pb.ConnectionState_DISCONNECTED
	ti.state.notificationsStatus = Invalid
	ti.redrawChan = make(chan struct{})
	ti.initialChan = make(chan struct{})
	ti.updateChan = make(chan bool)

	time.AfterFunc(NotifierStartDelay, func() { ti.notifier.start() })

	go ti.pollingMonitor()
}

func (ti *Instance) OnExit() {
	ti.state.mu.Lock()
	ti.state.systrayRunning = false
	ti.state.mu.Unlock()
}

func (ti *Instance) OnReady() {
	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")

	ti.state.mu.Lock()
	if ti.state.vpnStatus == pb.ConnectionState_DISCONNECTED {
		systray.SetIconName(ti.iconDisconnected)
	} else {
		systray.SetIconName(ti.iconConnected)
	}
	ti.state.systrayRunning = true
	ti.state.mu.Unlock()

	go func() {
		for {
			ti.state.mu.RLock()
			if ti.state.daemonAvailable {
				if ti.state.loggedIn {
					addVpnSection(ti)
				}
				addSettingsSection(ti)
				addAccountSection(ti)
			}
			if ti.state.daemonError != "" {
				addDaemonErrorSection(ti)
			}
			ti.state.mu.RUnlock()
			if ti.debugMode {
				addDebugSection(ti)
			}
			addQuitItem(ti)
			systray.Refresh()
			<-ti.redrawChan
			ti.state.mu.RLock()
			if !ti.state.systrayRunning {
				ti.state.mu.RUnlock()
				break
			}
			ti.state.mu.RUnlock()
			if ti.debugMode {
				log.Println(internal.DebugPrefix, "Redraw")
			}
			systray.ResetMenu()
		}
	}()
}
