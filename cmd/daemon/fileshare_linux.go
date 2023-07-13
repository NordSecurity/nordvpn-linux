//go:build !drop

package main

import (
	"github.com/NordSecurity/nordvpn-linux/fileshare/service"
)

func fileshareImplementation() service.Fileshare {
	return service.MockFileshare{}
}
