//go:build !telio

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
)

type noopConfigGetter struct{}

func (noopConfigGetter) GetConfig(string) (string, error) {
	return `{"direct": {}}`, nil
}

func vpnLibConfigGetterImplementation(_ config.Manager) vpn.LibConfigGetter {
	return noopConfigGetter{}
}
