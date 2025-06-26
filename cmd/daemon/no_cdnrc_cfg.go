//go:build !cdnrc

package main

import (
	"fmt"
)

type RemoteConfigGetterStub struct{}

func (r RemoteConfigGetterStub) GetTelioConfig() (string, error) {
	return "", fmt.Errorf("no remote config getter was compiled into the app")
}

type RemoteStorage interface {
	GetRemoteFile(string) ([]byte, error)
}

func getRemoteConfigGetter(_, _, _ string, _ RemoteStorage) RemoteConfigGetterStub {
	return RemoteConfigGetterStub{}
}
func (r RemoteConfigGetterStub) IsFeatureEnabled(string) bool { return false }
func (r RemoteConfigGetterStub) GetFeatureParam(_, _ string)  {}
func (r RemoteConfigGetterStub) LoadConfig() error            { return nil }
