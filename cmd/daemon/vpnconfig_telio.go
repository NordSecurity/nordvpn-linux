//go:build telio

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx/libtelio"
)

func vpnLibConfigGetterImplementation(cm config.Manager) vpn.LibConfigGetter {
	rcConfig := remote.NewRConfig(remote.UpdatePeriod, remote.NewFirebaseService(FirebaseToken), cm)
	return libtelio.NewTelioConfig(rcConfig)
}
