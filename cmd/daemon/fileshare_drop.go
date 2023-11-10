//go:build drop

package main

import (
	"os"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/fileshare/daemon"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func fileshareImplementation() daemon.Fileshare {
	switch {
	case os.Getenv(internal.ListenPID) == strconv.Itoa(os.Getpid()):
		// Try to use systemd, but fallback to fork
		return daemon.NewCombinedFileshare(&daemon.SystemdFileshare{}, &daemon.ForkFileshare{})
	default:
		/*
			Start filesharing directly on non-systemd scenarios.
			This comes with a drawback that after system reboot filesharing daemon will not
			be started automatically as the main daemon is started before the user session starts.
		*/
		return &daemon.ForkFileshare{}
	}

}
