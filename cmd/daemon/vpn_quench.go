//go:build quench

package main

import "github.com/NordSecurity/nordvpn-linux/daemon/vpn/quench"

func getQuenchVPN(fwmark uint32) (*quench.Quench, error) {
	return quench.New(fwmark), nil
}
