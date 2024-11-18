//go:build !quench

package main

import "github.com/NordSecurity/nordvpn-linux/daemon/vpn"

func getQuenchVPN(fwmark uint32) vpn.VPN {
	return nil
}
