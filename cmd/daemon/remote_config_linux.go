//go:build !firebase

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
)

type noopVersionGetter struct{}

func (noopVersionGetter) GetValue(key string) (string, error) {
	return "", nil
}

func (noopVersionGetter) GetTelioConfig(string) (string, error) {
	return `{"direct": {}}`, nil
}

func remoteConfigGetterImplementation(_ config.Manager) remote.RemoteConfigGetter {
	return noopVersionGetter{}
}
