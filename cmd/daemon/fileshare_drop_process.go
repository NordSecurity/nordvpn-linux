//go:build drop && fileshare_process

package main

import (
	"github.com/NordSecurity/nordvpn-linux/fileshare/service"
	"github.com/NordSecurity/nordvpn-linux/fileshare_process"
)

func fileshareImplementation() service.Fileshare {
	return fileshare_process.NewGRPCFileshareProcess()
}
