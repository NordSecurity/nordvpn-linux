package meshnet

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var ErrIncorrectCmdlineContent = errors.New("invalid content of cmdline file of /proc")

// FilesharePortAccessController blocks or allows fileshare port when
// fileshare process stopped or was restarted accordingly.
type FilesharePortAccessController struct {
	cm             config.Manager
	netw           Networker
	reg            mesh.Registry
	filesharePID   PID
	processChecker ProcessChecker
	mu             sync.Mutex
}

func NewPortAccessController(
	cm config.Manager,
	netw Networker,
	reg mesh.Registry,
	pc ProcessChecker,
) FilesharePortAccessController {
	filesharePID := PID(0)
	// NOTE:if the fileshare is already running, set the initial PID.
	// This can happen only when the daemon was restarted, but nordfileshare
	// process was not - for example there was a panic in daemon.
	PID := pc.GiveProcessPID(internal.FileshareBinaryPath)
	if PID != nil {
		filesharePID = *PID
	}
	return FilesharePortAccessController{
		cm:             cm,
		netw:           netw,
		reg:            reg,
		filesharePID:   filesharePID,
		processChecker: pc,
	}
}

func (eventHandler *FilesharePortAccessController) OnProcessStarted(ev ProcEvent) {
	if eventHandler.filesharePID != 0 {
		// fileshare already started and we noted the PID, no need to
		// process next events anymore until the PID gets reset in [EventHandler.OnProcessStopped]
		return
	}

	// NOTE: at this point, we can ignore older processes. It's because
	// we checked above that the [eventHandler.filesharePID] is not set
	// which means that nordfileshare process was not running at the time
	// of creation of [FilesharePortAccessController] - constructor checks
	// if nordfilshare is already running - so we know that nordfileshare
	// PID will be higher than the daemon PID.
	if ev.PID < eventHandler.processChecker.CurrentPID() {
		return
	}

	if !eventHandler.processChecker.IsFileshareProcess(ev.PID) {
		return
	}

	log.Println(internal.InfoPrefix, "updating fileshare process pid to:", ev.PID)
	eventHandler.filesharePID = ev.PID
	go eventHandler.allowFileshare()
}

func (eventHandler *FilesharePortAccessController) allowFileshare() error {
	log.Println(internal.InfoPrefix, "allowing fileshare port")

	eventHandler.mu.Lock()
	defer eventHandler.mu.Unlock()

	peers, err := eventHandler.listPeers()
	if err != nil {
		return err
	}

	for _, peer := range peers {
		peerUniqAddr := UniqueAddress{UID: peer.PublicKey, Address: peer.Address}
		if err := eventHandler.netw.AllowFileshare(peerUniqAddr); err != nil {
			return err
		}
	}

	return nil
}

func (eventHandler *FilesharePortAccessController) listPeers() (mesh.MachinePeers, error) {
	var cfg config.Config
	if err := eventHandler.cm.Load(&cfg); err != nil {
		return nil, fmt.Errorf("reading configuration when listing peers: %w", err)
	}

	if cfg.MeshDevice == nil {
		return nil, fmt.Errorf("meshnet is not configured")
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := eventHandler.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		return nil, fmt.Errorf("listing peers: %w", err)
	}
	return peers, nil
}

func (eventHandler *FilesharePortAccessController) OnProcessStopped(ev ProcEvent) {
	if eventHandler.filesharePID != ev.PID {
		return
	}
	log.Println(internal.InfoPrefix, "resetting fileshare pid")
	eventHandler.filesharePID = 0
	go eventHandler.blockFileshare()
}

func (eventHandler *FilesharePortAccessController) blockFileshare() error {
	log.Println(internal.InfoPrefix, "blocking fileshare port")

	eventHandler.mu.Lock()
	defer eventHandler.mu.Unlock()

	peers, err := eventHandler.listPeers()
	if err != nil {
		return err
	}

	for _, peer := range peers {
		peerUniqAddr := UniqueAddress{UID: peer.PublicKey, Address: peer.Address}
		if err := eventHandler.netw.BlockFileshare(peerUniqAddr); err != nil {
			return err
		}
	}

	return nil
}

// ProcessChecker represents process-related utilities
type ProcessChecker interface {
	IsFileshareProcess(PID) bool
	GiveProcessPID(string) *PID
	CurrentPID() PID
}
