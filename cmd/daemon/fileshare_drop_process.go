//go:build drop && fileshare_process

package main

import (
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/service"
)

func fileshareImplementation() service.Fileshare {
	return fileshare_process.NewGRPCFileshareProcess()
}
