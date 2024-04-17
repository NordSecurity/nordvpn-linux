//go:build !telio

package main

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
)

type noopConfigGetter struct{}

func (noopConfigGetter) GetConfig(string) (string, error) {
	return "", fmt.Errorf("config is not available")
}

func vpnLibConfigGetterImplementation(_ config.Manager) vpn.LibConfigGetter {
	return noopConfigGetter{}
}
