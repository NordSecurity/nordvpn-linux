//go:build firebase

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
)

var FirebaseToken = ""

func remoteConfigGetterImplementation(cm config.Manager) remote.RemoteConfigGetter {
	return remote.NewRConfig(remote.UpdatePeriod, remote.NewFirebaseService(FirebaseToken), cm)
}
