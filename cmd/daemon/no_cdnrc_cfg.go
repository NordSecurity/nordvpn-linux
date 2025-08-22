//go:build !cdnrc

package main

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
)

type RemoteConfigGetterStub struct{}

func (r RemoteConfigGetterStub) GetTelioConfig() (string, error) {
	return "", fmt.Errorf("no remote config getter was compiled into the app")
}

func getRemoteConfigGetter(_ config.BuildTarget, _ string, _ core.RemoteStorage, _ remote.Analytics, _ int) RemoteConfigGetterStub {
	return RemoteConfigGetterStub{}
}
func (r RemoteConfigGetterStub) IsFeatureEnabled(string) bool                { return false }
func (r RemoteConfigGetterStub) GetFeatureParam(_, _ string) (string, error) { return "", nil }
func (r RemoteConfigGetterStub) LoadConfig() error                           { return nil }
func (r RemoteConfigGetterStub) LoadConfigFromDisk()                         {}
func (r RemoteConfigGetterStub) Subscribe(remote.RemoteConfigNotifier)       {}
