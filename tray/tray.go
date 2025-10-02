package tray

import (
	"context"
	"errors"
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
	IconBlack string = "nordvpn-tray-black"
	IconGray  string = "nordvpn-tray-gray"
	IconWhite string = "nordvpn-tray-white"
	IconBlue  string = "nordvpn-tray-blue"
	logTag           = "[systray]"
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
	mu               sync.RWMutex
	countries        []string
	specialtyServers []string
}

func (cp *ConnectionSelector) fetchCountries(client pb.DaemonClient) ([]string, error) {
	resp, err := client.Countries(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	result := sortedConnections(resp.Servers)

	cp.mu.Lock()
	cp.countries = result
	cp.mu.Unlock()

	return slices.Clone(cp.countries), nil
}

func (cp *ConnectionSelector) fetchSpecialtyServers(client pb.DaemonClient) ([]string, error) {
	resp, err := client.Groups(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}

	result := sortedConnections(resp.Servers)

	cp.mu.Lock()
	cp.specialtyServers = result
	cp.mu.Unlock()

	return slices.Clone(cp.specialtyServers), nil
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
	checkboxSync        *CheckboxSynchronizer
}

type trayState struct {
	daemonAvailable      bool
	loggedIn             bool
	vpnActive            bool
	notificationsStatus  Status
	trayStatus           Status
	daemonError          string
	accountName          string
	vpnStatus            pb.ConnectionState
	vpnName              string
	vpnHostname          string
	vpnCity              string
	vpnCountry           string
	vpnVirtualLocation   bool
	initialSyncCompleted bool
	connSelector         ConnectionSelector
	mu                   sync.RWMutex
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
		checkboxSync:      NewCheckboxSynchronizer(),
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

func waitUntilTrayStatusIsReceived(fetchTrayStatus func() Status, doneChan chan<- struct{}) error {
	if fetchTrayStatus == nil || doneChan == nil {
		return errors.New("invalid arguments")
	}

	return RetryWithBackoff(
		context.Background(),
		BackoffConfig{MaxDelay: time.Second, MaxRetries: 10},
		func(ctx context.Context) error {
			if fetchTrayStatus() == Invalid {
				return fmt.Errorf("failed to get tray status")
			}

			select {
			case doneChan <- struct{}{}:
				close(doneChan)
				doneChan = nil
			default:
				// Channel already being processed or closed
			}

			return nil
		},
	)
}

// syncWithDaemon performs processes needed before systray can be shown
func (ti *Instance) syncWithDaemon() {
	for {
		getTrayStatusFunc := func() Status {
			changed := ti.updateSettings()
			ti.redraw(changed)
			return ti.state.trayStatus
		}
		err := waitUntilTrayStatusIsReceived(getTrayStatusFunc, ti.initialDataLoadChan)
		if err != nil {
			log.Printf("%s %s Waiting for tray state: %s. Retrying.\n", logTag, internal.ErrorPrefix, err)
			continue
		}
		break
	}
}

func (ti *Instance) Start() {
	ti.configureDebugMode()
	ti.updateIconsSelection()

	ti.state.vpnStatus = pb.ConnectionState_DISCONNECTED
	ti.state.notificationsStatus = Invalid
	ti.renderChan = make(chan struct{})
	ti.initialDataLoadChan = make(chan struct{})

	go ti.syncWithDaemon()
}

func (ti *Instance) OnExit() {
	ti.stateListener.Stop()
}

func (ti *Instance) OnReady(ctx context.Context) {
	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")

	ti.notifier.start()
	ti.stateListener.Start()

	go ti.renderLoop(ctx)

	ti.state.mu.Lock()
	ti.updateIcon()
	ti.state.mu.Unlock()
}

func (ti *Instance) updateIcon() {
	if ti.state.vpnStatus == pb.ConnectionState_CONNECTED {
		systray.SetIconName(ti.iconConnected)
	} else {
		systray.SetIconName(ti.iconDisconnected)
	}
}

func (ti *Instance) renderLoop(ctx context.Context) {
	for {
		// Wait for any checkbox operations to complete before rebuilding menu
		ti.checkboxSync.WaitForOperations()
		systray.ResetMenu()

		ti.state.mu.RLock()
		ti.updateIcon()
		buildConnectionSection(ti)
		buildSettingsSection(ti)
		buildAccountSection(ti)
		buildDaemonErrorSection(ti)
		addDebugSection(ti)
		buildQuitButton(ti)
		systray.Refresh()
		ti.state.mu.RUnlock()
		<-ti.renderChan

		if ti.debugMode {
			log.Println(internal.DebugPrefix, "Render")
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
