package tray

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
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
	// CountryListUpdateInterval = 60 * time.Minute
	// AccountInfoUpdateInterval = 24 * time.Hour
	logTag = "[systray]"
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
}

// getAccountInfo use cache to not query API every time
func (ai *accountInfo) getAccountInfo(client pb.DaemonClient) (*pb.AccountResponse, error) {
	var err error
	ai.accountInfo, err = client.AccountInfo(context.Background(), &pb.AccountRequest{Full: true})
	if err != nil {
		return nil, err
	}
	return ai.accountInfo, nil
}

func (ai *accountInfo) reset() {
	ai.accountInfo = nil
}

type ConnectionSelector struct {
	mu        sync.RWMutex
	countries []string
}

func (cp *ConnectionSelector) listCountries(client pb.DaemonClient) ([]string, error) {
	resp, err := client.Countries(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	result := sortedConnections(resp.Servers)

	cp.mu.Lock()
	cp.countries = result
	out := slices.Clone(cp.countries)
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
	client              pb.DaemonClient
	fileshareClient     filesharepb.FileshareClient
	accountInfo         accountInfo
	debugMode           bool
	notifier            dbusNotifier
	renderChan          chan struct{}
	initialDataLoadChan chan struct{}
	iconConnected       string
	iconDisconnected    string
	state               trayState
	quitChan            chan<- norduser.StopRequest
	stateListener       *stateListener
	connSensor          *connectionSettingsChangeSensor
	recentConnections   *recentConnectionsManager
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
	obj := &Instance{
		client:            client,
		fileshareClient:   fileshareClient,
		quitChan:          quitChan,
		connSensor:        newConnectionSettingsChangeSensor(),
		recentConnections: newRecentConnectionsManager(client),
	}
	obj.stateListener = newStateListener(client, obj.onDaemonStateEvent)
	return obj
}

func (ti *Instance) WaitInitialTrayStatus() Status {
	<-ti.initialDataLoadChan
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

func (ti *Instance) onDaemonStateEvent(item *pb.AppState) {
	switch st := item.GetState().(type) {
	case *pb.AppState_Error:
		log.Printf("%s %s Received daemon error state\n", logTag, internal.ErrorPrefix)
		ti.updateDaemonConnectionStatus(internal.ErrDaemonConnectionRefused.Error())

	case *pb.AppState_ConnectionStatus:
		ti.updateVpnStatus()
		ti.updateRecentConnections()

	case *pb.AppState_LoginEvent:
		ti.updateLoginStatus()

	case *pb.AppState_SettingsChange:
		ti.setSettings(st.SettingsChange)
		// identify whether we need to also update a country list
		ti.connSensor.Set(connectionSettings{
			Obfuscated:      st.SettingsChange.Obfuscate,
			Protocol:        st.SettingsChange.Protocol,
			Technology:      st.SettingsChange.Technology,
			VirtualLocation: st.SettingsChange.VirtualLocation,
		})

		if ti.connSensor.Detected() {
			ti.updateCountryList()
			ti.updateRecentConnections()
		}

	case *pb.AppState_UpdateEvent:
		if st.UpdateEvent == pb.UpdateEvent_SERVERS_LIST_UPDATE {
			ti.updateCountryList()
		}

	case *pb.AppState_AccountModification:
		ti.updateAccountInfo()

	default:
		log.Printf("%s %s Unknown state type: %T\n", logTag, internal.WarningPrefix, item)
	}
}

func (ti *Instance) Start() {
	ti.configureDebugMode()
	ti.updateIconsSelection()

	ti.state.vpnStatus = pb.ConnectionState_DISCONNECTED
	ti.state.notificationsStatus = Invalid
	ti.renderChan = make(chan struct{})
	ti.initialDataLoadChan = make(chan struct{})

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// wait until we see the daemon alive
		_ = RetryWithBackoff(
			ctx,
			DefaultBackoffConfig(),
			func(ctx context.Context) error {
				return ti.ping()
			},
		)

		// wait until we know whether tray should be enabled
		err := RetryWithBackoff(
			ctx,
			BackoffConfig{MaxDelay: time.Second, MaxRetries: 10},
			func(ctx context.Context) error {
				ti.update()
				if ti.state.trayStatus == Invalid {
					return fmt.Errorf("failed to get tray status")
				}

				if ti.initialDataLoadChan != nil {
					select {
					case ti.initialDataLoadChan <- struct{}{}:
						close(ti.initialDataLoadChan)
						ti.initialDataLoadChan = nil
					default:
						// Channel already being processed or closed
					}
				}
				cancel()
				return nil
			},
		)
		if err != nil {
			log.Printf("%s %s waiting for tray state: %s\n", logTag, internal.ErrorPrefix, err)
		}
	}()
}

func (ti *Instance) OnExit() {
	ti.stateListener.Stop()
	ti.state.mu.Lock()
	ti.state.systrayRunning = false
	ti.state.mu.Unlock()
}

func (ti *Instance) OnReady() {
	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")

	ti.notifier.start()
	ti.stateListener.Start()

	go ti.renderLoop()

	ti.state.mu.Lock()
	if ti.state.vpnStatus == pb.ConnectionState_DISCONNECTED {
		systray.SetIconName(ti.iconDisconnected)
	} else {
		systray.SetIconName(ti.iconConnected)
	}
	ti.state.systrayRunning = true
	ti.state.mu.Unlock()
}

func (ti *Instance) renderLoop() {
	for {
		systray.ResetMenu()

		ti.state.mu.RLock()
		buildConnectionSection(ti)
		buildSettingsSection(ti)
		buildAccountSection(ti)
		buildDaemonErrorSection(ti)
		ti.state.mu.RUnlock()
		addDebugSection(ti)
		buildQuitButton(ti)
		systray.Refresh()
		<-ti.renderChan

		ti.state.mu.RLock()
		if !ti.state.systrayRunning {
			ti.state.mu.RUnlock()
			break
		}
		ti.state.mu.RUnlock()
		if ti.debugMode {
			log.Println(internal.DebugPrefix, "Render")
		}
	}
}
