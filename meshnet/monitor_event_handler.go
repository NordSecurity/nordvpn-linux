package meshnet

import (
	"errors"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

var ErrIncorrectCmdlineContent = errors.New("invalid content of cmdline file of /proc")

// FilesharePortAccessController forbids or permits fileshare port
// use when fileshare process stopped or was restarted accordingly.
type FilesharePortAccessController struct {
	netw           FileshareNetworker
	filesharePID   PID
	processChecker ProcessChecker
}

func NewPortAccessController(netw FileshareNetworker, pc ProcessChecker) *FilesharePortAccessController {
	filesharePID := PID(0)
	// NOTE:if the fileshare is already running, set the initial PID.
	// This can happen only when the daemon was restarted, but nordfileshare
	// process was not - for example there was a panic in daemon.
	PID := pc.GiveProcessPID(internal.FileshareBinaryPath)
	if PID != nil {
		filesharePID = *PID
	}
	return &FilesharePortAccessController{
		netw:           netw,
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

	if !eventHandler.processChecker.IsFileshareProcess(ev.PID) {
		return
	}

	log.Println(internal.InfoPrefix, "updating fileshare process pid to:", ev.PID)
	eventHandler.filesharePID = ev.PID
	go eventHandler.netw.PermitFileshare()
}

func (eventHandler *FilesharePortAccessController) OnProcessStopped(ev ProcEvent) {
	if eventHandler.filesharePID == 0 {
		return
	}

	if eventHandler.filesharePID != ev.PID {
		return
	}

	log.Println(internal.InfoPrefix, "resetting fileshare pid")
	eventHandler.filesharePID = 0
	go eventHandler.netw.ForbidFileshare()
}

// ProcessChecker represents process-related utilities
type ProcessChecker interface {
	IsFileshareProcess(PID) bool
	GiveProcessPID(string) *PID
	CurrentPID() PID
}

// FileshareNetworker represents ability of a networker to permit or forbid fileshare
type FileshareNetworker interface {
	PermitFileshare() error
	ForbidFileshare() error
}
