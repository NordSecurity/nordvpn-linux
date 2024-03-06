//go:build !fileshare_process

package main

import (
	"fmt"
	"os"

	"github.com/NordSecurity/nordvpn-linux/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var FileshareURL = fmt.Sprintf("%s://%s", internal.Proto, internal.GetFilesharedSocket(os.Getuid()))

func BuildFileshareProcessManager() fileshare_process.FileshareProcess {
	return fileshare_process.NoopFileshareProcess{}
}
