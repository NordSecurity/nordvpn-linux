//go:build quench

package main

import "github.com/NordSecurity/nordvpn-linux/daemon/vpn/quench"

func getNordWhisperVPN(fwmark uint32, envIsDev bool) (*quench.Quench, error) {
	return quench.New(fwmark, envIsDev), nil
}
