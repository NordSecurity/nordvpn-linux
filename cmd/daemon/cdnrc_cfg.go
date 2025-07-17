//go:build cdnrc

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type RemoteStorage interface {
	GetRemoteFile(name string) ([]byte, error)
}

func getRemoteConfigGetter(buildTarget config.BuildTarget, rpath string, cdn RemoteStorage) *remote.CdnRemoteConfig {
	return remote.NewCdnRemoteConfig(buildTarget, rpath, internal.ConfigFilesPathCommon, cdn)
}
