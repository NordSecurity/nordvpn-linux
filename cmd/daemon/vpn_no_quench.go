//go:build !quench

package main

import "github.com/NordSecurity/nordvpn-linux/daemon/vpn"

func getQuenchVPN() vpn.VPN {
	return nil
}
