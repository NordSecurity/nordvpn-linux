//go:build !quench

package main

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
)

func getQuenchVPN(fwmark uint32) (vpn.VPN, error) {
	return nil, fmt.Errorf("quench is not enabled")
}
