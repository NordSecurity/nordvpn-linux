//go:build drop && !fileshare_process

package main

import (
	"os"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/service"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

func fileshareImplementation() service.Fileshare {
	if snapconf.IsUnderSnap() {
		return fileshare_process.NewGRPCFileshareProcess()
	}

	switch {
	case os.Getenv(internal.ListenPID) == strconv.Itoa(os.Getpid()):
		// Try to use systemd, but fallback to fork
		return service.NewCombinedFileshare(&service.SystemdFileshare{}, &service.ForkFileshare{})
	default:
		/*
			Start filesharing directly on non-systemd scenarios.
			This comes with a drawback that after system reboot filesharing daemon will not
			be started automatically as the main daemon is started before the user session starts.
		*/
		return &service.ForkFileshare{}
	}

}
