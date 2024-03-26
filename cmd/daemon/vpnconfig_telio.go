//go:build telio

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx/libtelio"
)

func vpnLibConfigGetterImplementation(cm config.Manager) vpn.LibConfigGetter {
	return libtelio.NewTelioConfig(cm)
}
