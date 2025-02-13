//go:build quench

package main

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/quench"
)

func getNordWhisperVPN(fwmark uint32, envIsDev bool, events *vpn.Events) (*quench.Quench, error) {
	return quench.New(fwmark, envIsDev, events), nil
}
