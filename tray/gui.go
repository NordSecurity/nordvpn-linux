package tray

import (
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/log"
)

var guiPID atomic.Int64

func openGUI() {
	pid := guiPID.Load()
	log.Infof("%s openGUI called, guiPID=%d", logTag, pid)

	if pid > 0 {
		log.Infof("%s process already running (pid=%d), sending SIGUSR1", logTag, pid)
		if p, err := os.FindProcess(int(pid)); err == nil {
			if err := p.Signal(syscall.SIGUSR1); err != nil {
				log.Errorf("%s SIGUSR1 failed: %v", logTag, err)
			} else {
				log.Infof("%s SIGUSR1 sent to pid=%d", logTag, pid)
			}
		} else {
			log.Errorf("%s FindProcess(%d) failed: %v", logTag, pid, err)
		}
		return
	}

	if pid == -1 {
		log.Infof("%s spawn already in progress, skipping", logTag)
		return
	}

	if !guiPID.CompareAndSwap(0, -1) {
		log.Infof("%s CAS failed, someone else is spawning", logTag)
		return
	}

	guiBin, err := exec.LookPath("nordvpn-gui")
	if err != nil {
		log.Errorf("%s nordvpn-gui not found in PATH: %v", logTag, err)
		guiPID.Store(0)
		return
	}
	log.Infof("%s found nordvpn-gui at %s", logTag, guiBin)

	// #nosec G204 -- path comes from exec.LookPath, not user input
	cmd := exec.Command(guiBin)
	cmd.Env = append(os.Environ(), "GDK_BACKEND=x11", "NORDVPN_TRAY_LAUNCH=1")
	if err := cmd.Start(); err != nil {
		log.Errorf("%s failed to start nordvpn-gui: %v", logTag, err)
		guiPID.Store(0)
		return
	}

	spawnedPID := int64(cmd.Process.Pid)
	log.Infof("%s nordvpn-gui spawned with pid=%d", logTag, spawnedPID)
	guiPID.Store(spawnedPID)

	go func() {
		_ = cmd.Wait()
		log.Infof("%s pid=%d exited", logTag, spawnedPID)
		guiPID.Store(0)
	}()
}
