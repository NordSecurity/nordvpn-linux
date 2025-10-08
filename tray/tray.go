package tray

import (
	"context"
	"log"
	"os"
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
	NotifierStartDelay        = 3 * time.Second
	PollingUpdateInterval     = 5 * time.Second
	PollingFullUpdateInterval = 60 * time.Second
	AccountInfoUpdateInterval = 24 * time.Hour
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
	since := time.Since(ai.updateTime)
	if ai.updateTime.IsZero() {
		log.Println(internal.DebugPrefix, "accountInfo cache is empty; will query daemon")
	} else {
		log.Printf("%s accountInfo cache age: %s (ttl: %s)\n", internal.DebugPrefix, since, AccountInfoUpdateInterval)
	}

	if since > AccountInfoUpdateInterval {
		log.Println(internal.InfoPrefix, "Requesting fresh account info from daemon")
		start := time.Now()
		var err error
		ai.accountInfo, err = client.AccountInfo(context.Background(), &pb.AccountRequest{Full: true})
		if err != nil {
			log.Printf("%s failed to fetch account info: %v\n", internal.ErrorPrefix, err)
			return nil, err
		}
		ai.updateTime = time.Now()
		log.Printf("%s account info updated successfully in %s; next refresh after %s\n", internal.InfoPrefix, time.Since(start), AccountInfoUpdateInterval)
	} else {
		log.Println(internal.DebugPrefix, "Using cached account info")
	}
	return ai.accountInfo, nil
}

func (ai *accountInfo) reset() {
	log.Println(internal.InfoPrefix, "Resetting cached account info")
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
	log.Println(internal.InfoPrefix, "Creating new tray instance")
	return &Instance{client: client, fileshareClient: fileshareClient, quitChan: quitChan}
}

func (ti *Instance) WaitInitialTrayStatus() Status {
	log.Println(internal.InfoPrefix, "Waiting for initial tray status signal on initialChan…")
	<-ti.initialChan
	ti.state.mu.RLock()
	defer ti.state.mu.RUnlock()
	log.Printf("%s initial tray status received: %s\n", internal.InfoPrefix, ti.statusToString(ti.state.trayStatus))
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
	desktopEnv := sysinfo.GetDisplayDesktopEnvironment()
	log.Printf("%s detected desktop environment: %q\n", internal.InfoPrefix, desktopEnv)
	selected := selectIcon(desktopEnv)
	pathDisconnected := notify.GetIconPath(selected)
	pathConnected := notify.GetIconPath(IconBlue)
	log.Printf("%s icon selection — disconnected:%q, connected:%q\n", internal.InfoPrefix, pathDisconnected, pathConnected)
	ti.iconDisconnected = pathDisconnected
	ti.iconConnected = pathConnected
}

// configureDebugMode configures debug mode based on the environment variable.
func (ti *Instance) configureDebugMode() {
	val := os.Getenv("NORDVPN_TRAY_DEBUG")
	if val == "1" {
		ti.debugMode = true
		log.Println(internal.InfoPrefix, "Debug mode ENABLED via NORDVPN_TRAY_DEBUG=1")
	} else {
		ti.debugMode = false
		log.Printf("%s Debug mode disabled (NORDVPN_TRAY_DEBUG=%q)\n", internal.InfoPrefix, val)
	}
}

func (ti *Instance) Start() {
	log.Println(internal.InfoPrefix, "Starting tray instance")
	ti.configureDebugMode()
	ti.updateIconsSelection()

	// initialize state and channels
	ti.state.vpnStatus = pb.ConnectionState_DISCONNECTED
	ti.state.notificationsStatus = Invalid
	ti.redrawChan = make(chan struct{})
	ti.initialChan = make(chan struct{})
	ti.updateChan = make(chan bool)
	log.Printf("%s Channels created (redrawChan, initialChan, updateChan). Initial vpnStatus=%s, notificationsStatus=%s\n",
		internal.DebugPrefix, ti.vpnConnStateToString(ti.state.vpnStatus), ti.statusToString(ti.state.notificationsStatus))

	time.AfterFunc(NotifierStartDelay, func() {
		log.Println(internal.InfoPrefix, "Starting dbus notifier in AfterFunc")
		ti.notifier.start()
	})
	log.Printf("%s dbus notifier scheduled to start after %s\n", internal.DebugPrefix, NotifierStartDelay)

	log.Println(internal.InfoPrefix, "Starting polling monitor in new goroutine")
	go ti.pollingMonitor()
}

func (ti *Instance) OnExit() {
	log.Println(internal.InfoPrefix, "OnExit: marking systray as not running")
	ti.state.mu.Lock()
	ti.state.systrayRunning = false
	ti.state.mu.Unlock()
}

func (ti *Instance) OnReady() {
	log.Println(internal.InfoPrefix, "OnReady: configuring systray title and tooltip")
	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")

	ti.state.mu.Lock()
	if ti.state.vpnStatus == pb.ConnectionState_DISCONNECTED {
		log.Printf("%s OnReady: setting disconnected icon: %q\n", internal.DebugPrefix, ti.iconDisconnected)
		systray.SetIconName(ti.iconDisconnected)
	} else {
		log.Printf("%s OnReady: setting connected icon: %q\n", internal.DebugPrefix, ti.iconConnected)
		systray.SetIconName(ti.iconConnected)
	}
	ti.state.systrayRunning = true
	log.Println(internal.InfoPrefix, "OnReady: systrayRunning set to true")
	ti.state.mu.Unlock()

	log.Println(internal.InfoPrefix, "OnReady: launching UI builder loop goroutine")
	go func() {
		for {
			// snapshot state under read lock for decision logging
			ti.state.mu.RLock()
			da := ti.state.daemonAvailable
			li := ti.state.loggedIn
			derr := ti.state.daemonError
			siRunning := ti.state.systrayRunning
			ti.state.mu.RUnlock()

			log.Printf("%s UI loop tick — daemonAvailable=%t, loggedIn=%t, daemonError=%q, systrayRunning=%t\n",
				internal.DebugPrefix, da, li, derr, siRunning)

			ti.state.mu.RLock()
			if ti.state.daemonAvailable {
				if ti.state.loggedIn {
					log.Println(internal.DebugPrefix, "Adding VPN section to tray menu")
					addVpnSection(ti)
				} else {
					log.Println(internal.DebugPrefix, "Skipping VPN section (not logged in)")
				}
				log.Println(internal.DebugPrefix, "Adding Settings section to tray menu")
				addSettingsSection(ti)
				log.Println(internal.DebugPrefix, "Adding Account section to tray menu")
				addAccountSection(ti)
			} else {
				log.Println(internal.DebugPrefix, "Daemon not available — skipping sections that require it")
			}
			if ti.state.daemonError != "" {
				log.Printf("%s Adding Daemon Error section: %q\n", internal.ErrorPrefix, ti.state.daemonError)
				addDaemonErrorSection(ti)
			}
			ti.state.mu.RUnlock()

			if ti.debugMode {
				log.Println(internal.DebugPrefix, "Adding Debug section to tray menu")
				addDebugSection(ti)
			}
			log.Println(internal.DebugPrefix, "Adding Quit item to tray menu")
			addQuitItem(ti)

			log.Println(internal.DebugPrefix, "Refreshing systray UI")
			systray.Refresh()

			log.Println(internal.DebugPrefix, "UI loop waiting for redraw signal on redrawChan…")
			<-ti.redrawChan

			// after wake-up, check if systray still running
			ti.state.mu.RLock()
			if !ti.state.systrayRunning {
				ti.state.mu.RUnlock()
				log.Println(internal.InfoPrefix, "UI loop exiting — systrayRunning is false")
				break
			}
			ti.state.mu.RUnlock()

			if ti.debugMode {
				log.Println(internal.DebugPrefix, "Redraw requested — resetting menu")
			}
			systray.ResetMenu()
		}
	}()
}

// Helpers
func (ti *Instance) statusToString(s Status) string {
	switch s {
	case Enabled:
		return "Enabled"
	case Disabled:
		return "Disabled"
	case Invalid:
		return "Invalid"
	default:
		return "<unknown>"
	}
}

func (ti *Instance) vpnConnStateToString(s pb.ConnectionState) string {
	switch s {
	case pb.ConnectionState_CONNECTED:
		return "CONNECTED"
	case pb.ConnectionState_CONNECTING:
		return "CONNECTING"
	case pb.ConnectionState_DISCONNECTED:
		return "DISCONNECTED"
	default:
		return "<unknown>"
	}
}

