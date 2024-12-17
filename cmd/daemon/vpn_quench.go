//go:build quench

package main

const QuenchEnabled = true

import "github.com/NordSecurity/nordvpn-linux/daemon/vpn/quench"

func getQuenchVPN(fwmark uint32) (*quench.Quench, error) {
	return quench.New(fwmark), nil
}
