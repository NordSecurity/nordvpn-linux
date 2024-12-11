package meshnet

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	processChecker processChecker
	mu             sync.Mutex
}

func NewPortAccessController(cm config.Manager, netw Networker, reg mesh.Registry) FilesharePortAccessController {
	return FilesharePortAccessController{
		cm:             cm,
		netw:           netw,
		reg:            reg,
		filesharePID:   0,
		processChecker: defaultProcChecker{},
	}
}

func (eventHandler *FilesharePortAccessController) OnProcessStarted(ev ProcEvent) {
	if eventHandler.filesharePID != 0 {
		// fileshare already started and we noted the PID, no need to
		// process next events anymore until the PID gets reset in [EventHandler.OnProcessStopped]
		return
	}
	if !eventHandler.processChecker.isFileshareProcess(ev.PID) {
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

// processChecker allows checking if process with specified [PID]
// is a fileshare process.
type processChecker interface {
	isFileshareProcess(PID) bool
}

// defaultProcChecker allows checking if process specified by [PID]
// is a fileshare process by reading its path via `/proc/<pid>/cmdline`.
type defaultProcChecker struct{}

func (defaultProcChecker) isFileshareProcess(pid PID) bool {
	// ignore older processes, fileshare is always
	// younger than the daemon so it has higher PID
	if pid < PID(os.Getpid()) {
		return false
	}

	procPath, err := readProcPath(pid)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to read process path from /proc", err)
		return false
	}

	return procPath == internal.FileshareBinaryPath
}

func readProcPath(pid PID) (string, error) {
	pidStr := strconv.FormatUint(uint64(pid), 10)
	cmdlinePath := filepath.Join("/proc", pidStr, "cmdline")

	cmdline, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return "", err
	}
	args := strings.Split(string(cmdline), "\x00")
	if len(args) == 0 {
		return "", ErrIncorrectCmdlineContent
	}
	return args[0], nil
}
