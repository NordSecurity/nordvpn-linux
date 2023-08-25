//go:build !firebase

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
)

type mockVersionGetter struct{}

func (mockVersionGetter) GetValue(key string) (string, error) {
	return "", nil
}

func (mockVersionGetter) GetTelioConfig(string) (string, error) {
	return "{}", nil
}

func remoteConfigGetterImplementation(_ config.Manager) remote.RemoteConfigGetter {
	return mockVersionGetter{}
}
