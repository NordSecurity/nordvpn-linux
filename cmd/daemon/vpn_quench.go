//go:build quench

package main

import "github.com/NordSecurity/nordvpn-linux/daemon/vpn/quench"

func getQuenchVPN() *quench.Quench {
	return quench.New()
}
