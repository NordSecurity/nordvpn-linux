//go:build !quench

package main

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

func getNordWhisperVPN(fwmark uint32) (vpn.VPN, error) {
	return noopNordWhisper{}, ErrNordWhisperDisabled
}

// noopMesh is a noop implementation of meshnet. It is used when telio
// is not available and should be used only for development purposes
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
