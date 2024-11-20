//go:build !firebase

package main

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
)

type RemoteConfigGetterStub struct {
}

func (r RemoteConfigGetterStub) GetTelioConfig(version string) (string, error) {
	return "", fmt.Errorf("firebase config was not compiled into the app")
}

func (r RemoteConfigGetterStub) GetQuenchEnabled(version string) (bool, error) {
	return false, nil
}

func getRemoteConfigGetter(_ config.Manager) remote.RemoteConfigGetter {
	return RemoteConfigGetterStub{}
}
