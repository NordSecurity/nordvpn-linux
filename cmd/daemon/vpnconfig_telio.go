//go:build telio

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx/libtelio"
)

func vpnLibConfigGetterImplementation(cm config.Manager, rcConfig remote.ConfigGetter) vpn.LibConfigGetter {
	return libtelio.NewTelioConfig(rcConfig.GetTelioConfig)
}
