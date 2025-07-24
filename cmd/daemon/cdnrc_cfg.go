//go:build cdnrc

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func getRemoteConfigGetter(buildTarget config.BuildTarget, rpath string, cdn core.RemoteStorage, appRollout int) *remote.CdnRemoteConfig {
	return remote.NewCdnRemoteConfig(buildTarget, rpath, internal.ConfigFilesPathCommon, cdn, appRollout)
}
