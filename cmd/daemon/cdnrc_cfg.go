//go:build cdnrc

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type CDN interface {
	GetRemoteFile(name string) ([]byte, error)
}

func getRemoteConfigGetter(ver, env string, cdn CDN) *remote.CdnRemoteConfig {
	return remote.NewCdnRemoteConfig(ver, env, internal.ConfigFilesPathCommon, cdn)
}
