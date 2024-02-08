//go:build drop

package main

import (
	"github.com/NordSecurity/nordvpn-linux/fileshare_process"
)

func fileshareImplementation() fileshare_process.GRPCFileshareProcess {
	return fileshare_process.GRPCFileshareProcess{}
}
