//go:build drop

package main

import (
	"os"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/meshnet/fork"
	"github.com/NordSecurity/nordvpn-linux/meshnet/systemd"
)

func fileshareImplementation() meshnet.Fileshare {
	switch {
	case os.Getenv(internal.ListenPID) == strconv.Itoa(os.Getpid()):
		return systemd.Fileshare{}
	default:
		/*
			Start filesharing directly on non-systemd scenarios.
			This comes with a drawback that after system reboot filesharing daemon will not
			be started automatically as the main daemon is started before the user session starts.
		*/
		return &fork.Fileshare{}
	}

}
