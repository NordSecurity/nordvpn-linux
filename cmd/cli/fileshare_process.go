//go:build fileshare_process

package main

import "github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"

var FileshareURL = fileshare_process.FileshareURL

func BuildFileshareProcessManager() fileshare_process.FileshareProcess {
	return fileshare_process.NewGRPCFileshareProcess()
}
