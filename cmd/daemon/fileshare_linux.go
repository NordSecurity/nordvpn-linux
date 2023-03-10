//go:build !drop

package main

import (
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/meshnet/mock"
)

func fileshareImplementation() meshnet.Fileshare {
	return mock.Fileshare{}
}
