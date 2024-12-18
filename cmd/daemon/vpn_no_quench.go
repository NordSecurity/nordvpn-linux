//go:build !quench

package main

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
)

func getNordWhisperVPN(fwmark uint32) (vpn.VPN, error) {
	return nil, fmt.Errorf("NordWhisper is not enabled")
}
