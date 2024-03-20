package tray

import (
	"fmt"
	"os"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/snapconf"

	"github.com/NordSecurity/systray"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: Deduplicate DaemonURL, FileshareURL and getFileshareURL with cmd/cli/main.go

const (
	NotifierStartDelay        = 3 * time.Second
	PollingUpdateInterval     = 1 * time.Second
	PollingFullUpdateInterval = 60 * time.Second
)

var (
	DaemonURL        = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
	FileshareURL     = getFileshareURL()
	client           pb.DaemonClient
	meshClient       meshpb.MeshnetClient
	fileshareClient  filesharepb.FileshareClient
	dbusNotifier     = &DbusNotifier{}
	redrawChan       chan struct{}
	updateChan       chan bool
	notifyEnabled    bool
	debugMode        bool
	iconConnected    = "nordvpn-tray-blue"
	iconDisconnected = "nordvpn-tray-white"
)

func getFileshareURL() string {
	if snapconf.IsUnderSnap() {
		return fileshare_process.FileshareURL
	}
	return fmt.Sprintf("%s://%s", internal.Proto, internal.GetFilesharedSocket(os.Getuid()))
}

func startDbusNotifier() {
	notifier, err := NewDbusNotifier()
	if err == nil {
		notification("info", "Started DbusNotifier")
		dbusNotifier = notifier
	} else {
		notification("error", "Failed to start DbusNotifier: %s", err)
	}
}

func main() {

	conn, err := grpc.Dial(
		DaemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	fileshareConn, err := grpc.Dial(
		FileshareURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err == nil {
		client = pb.NewDaemonClient(conn)
		meshClient = meshpb.NewMeshnetClient(conn)
		fileshareClient = filesharepb.NewFileshareClient(fileshareConn)
	}

	// TODO: Detect running DE and set iconDisconnected to "nordvpn-tray-black" on KDE/Plasma and to "nordvpn-tray-gray" on old Ubuntu

	notifyEnabled = true
	debugMode = true
	redrawChan = make(chan struct{})
	updateChan = make(chan bool)

	time.AfterFunc(NotifierStartDelay, startDbusNotifier)
	systray.Run(onReady, onExit)
}

func onExit() {
	if debugMode {
		now := time.Now()
		fmt.Println("Exit at", now.String())
	}
}

func onReady() {
	systray.SetTitle("NordVPN")
	systray.SetTooltip("NordVPN")
	systray.SetIconName(iconDisconnected)
	ticker := time.Tick(PollingUpdateInterval)
	go pollingMonitor(client, updateChan, ticker)

	go func() {
		for {
			state.mu.RLock()
			addAppSection()
			if state.daemonAvailable {
				if state.loggedIn {
					addVpnSection()
					addMeshnetSection()
				}
				addAccountSection()
			} else {
				addDaemonSection()
			}
			state.mu.RUnlock()
			if debugMode {
				addDebugSection()
			}
			addQuitItem()
			systray.Refresh()
			<-redrawChan
			if debugMode {
				fmt.Println(time.Now().String(), "Redraw")
			}
			systray.ResetMenu()
		}
	}()
}
