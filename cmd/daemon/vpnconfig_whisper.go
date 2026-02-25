//go:build quench

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/quench"
)

func vpnNordWhisperConfigGetterImplementation(cm config.Manager, rcConfig remote.ConfigGetter) vpn.NordWhisperConfigGetter {
	return quench.NewNordWhisperConfig(rcConfig)
}
