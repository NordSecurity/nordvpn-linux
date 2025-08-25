package tray

import (
	"context"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/notify"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"

	"github.com/NordSecurity/systray"
)

const (
	NotifierStartDelay                = 3 * time.Second
	PollingUpdateInterval             = 5 * time.Second
	PollingFullUpdateInterval         = 60 * time.Second
	CountryListUpdateInterval         = 60 * time.Minute
	SpecialtyServerListUpdateInterval = 60 * time.Minute
	AccountInfoUpdateInterval         = 24 * time.Hour
)

const (
	IconBlack string = "nordvpn-tray-black"
	IconGray  string = "nordvpn-tray-gray"
	IconWhite string = "nordvpn-tray-white"
	IconBlue  string = "nordvpn-tray-blue"
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
		ai.accountInfo, err = client.AccountInfo(context.Background(), &pb.AccountRequest{Full: true})
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

type ConnectionSelector struct {
	mu                  sync.RWMutex
	countries           []string
	countriesUpdateTime time.Time
}

func (cp *ConnectionSelector) listCountries(client pb.DaemonClient) ([]string, error) {
	cp.mu.Lock()
	needsUpdate := time.Since(cp.countriesUpdateTime) > CountryListUpdateInterval
	if !needsUpdate {
		out := append([]string(nil), cp.countries...)
		cp.mu.Unlock()
		return out, nil
	}
	cp.mu.Unlock()

	resp, err := client.Countries(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	result := sortedConnections(resp.Servers)

	cp.mu.Lock()
	cp.countries = result
	cp.countriesUpdateTime = time.Now()
	out := append([]string(nil), cp.countries...)
	cp.mu.Unlock()
	return out, nil
}

func sortedConnections(sgs []*pb.ServerGroup) []string {
	set := make(map[string]struct{}, len(sgs))
	for _, sg := range sgs {
		if c := strings.TrimSpace(sg.Name); c != "" {
			set[c] = struct{}{}
		}
	}
	list := make([]string, 0, len(set))
	for k := range set {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
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
	connSelector        ConnectionSelector
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

// selectIcon determines the icon color based on the desktop environment.
func selectIcon(desktopEnv string) string {
	switch desktopEnv {
	case "kde":
		// Kubuntu uses a dark tray background instead of KDE's default white.
		return IconBlack
	case "mate":
		return IconGray
	default:
		return IconWhite
	}
}

// updateIconsSelection selects the most appropriate icon based on the desktop environment.
func (ti *Instance) updateIconsSelection() {
	ti.iconDisconnected = notify.GetIconPath(selectIcon(sysinfo.GetDisplayDesktopEnvironment()))
	ti.iconConnected = notify.GetIconPath(IconBlue)
}

// configureDebugMode configures debug mode based on the environment variable.
func (ti *Instance) configureDebugMode() {
	ti.debugMode = os.Getenv("NORDVPN_TRAY_DEBUG") == "1"
}

func (ti *Instance) Start() {
	ti.configureDebugMode()
	ti.updateIconsSelection()

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
