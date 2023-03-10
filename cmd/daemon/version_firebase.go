//go:build firebase

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config/remote"
)

var FirebaseToken = ""

func versionGetterImplementation() remote.SupportedVersionGetter {
	return remote.NewRConfig(remote.UpdatePeriod, FirebaseToken)
}
