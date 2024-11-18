package quench

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// Start(context.Context, Credentials, ServerData) error
// Stop() error
// State() State // required because of OpenVPN
// IsActive() bool
// Tun() tunnel.T // required because of OpenVPN
// NetworkChanged() error
// // GetConnectionParameters returns ServerData of current connection and true if connection is established, or empty
// // ServerData and false if it isn't.
// GetConnectionParameters() (ServerData, bool)

type Quench struct{}

func New() *Quench {
	return &Quench{}
}

func (*Quench) Start(context.Context, vpn.Credentials, vpn.ServerData) error { panic("unimplemented") }
func (*Quench) Stop() error                                                  { panic("unimplemented") }
func (*Quench) State() vpn.State                                             { panic("unimplemented") }
func (*Quench) IsActive() bool                                               { return false }
func (*Quench) Tun() tunnel.T                                                { return &tunnel.Tunnel{} }
func (*Quench) NetworkChanged() error                                        { panic("unimplemented") }
func (*Quench) GetConnectionParameters() (vpn.ServerData, bool)              { panic("unimplemented") }
