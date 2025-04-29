//go:build !cdnrc

package main

import (
	"fmt"
)

type RemoteConfigGetterStub struct{}

func (r RemoteConfigGetterStub) GetTelioConfig() (string, error) {
	return "", fmt.Errorf("no remote config getter was compiled into the app")
}

func getRemoteConfigGetter(string) RemoteConfigGetterStub {
	return RemoteConfigGetterStub{}
}
