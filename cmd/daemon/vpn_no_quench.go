//go:build !quench

package main

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

func getNordWhisperVPN(fwmark uint32, _ bool) (vpn.VPN, error) {
	return noopNordWhisper{}, ErrNordWhisperDisabled
}

// noopNordWhisper is a noop implementation of NordWhisper used in build where NordWhisper is not enabled
type noopNordWhisper struct{}

func (noopNordWhisper) Start(context.Context, vpn.Credentials, vpn.ServerData) error { return nil }
func (noopNordWhisper) Stop() error                                                  { return nil }
func (noopNordWhisper) State() vpn.State                                             { return "" }
func (noopNordWhisper) IsActive() bool                                               { return false }
func (noopNordWhisper) Tun() tunnel.T                                                { return &tunnel.Tunnel{} }
func (noopNordWhisper) NetworkChanged() error                                        { return nil }
func (noopNordWhisper) GetConnectionParameters() (vpn.ServerData, bool) {
	return vpn.ServerData{}, false
}
