//go:build cdnrc

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type RemoteStorage interface {
	GetRemoteFile(name string) ([]byte, error)
}

func getRemoteConfigGetter(ver, env, rpath string, cdn RemoteStorage) *remote.CdnRemoteConfig {
	return remote.NewCdnRemoteConfig(ver, env, rpath, internal.ConfigFilesPathCommon, cdn)
}
