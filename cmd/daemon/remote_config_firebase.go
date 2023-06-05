//go:build firebase

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config/remote"
)

var FirebaseToken = ""

func remoteConfigGetterImplementation() remote.RemoteConfigGetter {
	return remote.NewRConfig(remote.UpdatePeriod, FirebaseToken)
}
