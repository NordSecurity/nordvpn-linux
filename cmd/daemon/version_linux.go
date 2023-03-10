//go:build !firebase

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/coreos/go-semver/semver"
)

type mockVersionGetter struct{}

func (mockVersionGetter) MinimalVersion() (*semver.Version, error) {
	return semver.NewVersion("0.0.0")
}

func (mockVersionGetter) GetMinFeatureVersion(featureKey string) (*semver.Version, error) {
	return semver.NewVersion("0.0.0")
}

func versionGetterImplementation() remote.SupportedVersionGetter {
	return mockVersionGetter{}
}
