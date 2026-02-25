//go:build !quench

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
)

type noopNordWhisperConfigGetter struct{}

func (noopNordWhisperConfigGetter) GetConfig() (vpn.NordWhisperFeatureConfig, error) {
	return vpn.NewNordWhisperFeatureConfig(), nil
}

func vpnNordWhisperConfigGetterImplementation(_ config.Manager, _ remote.ConfigGetter) vpn.NordWhisperConfigGetter {
	return noopNordWhisperConfigGetter{}
}
