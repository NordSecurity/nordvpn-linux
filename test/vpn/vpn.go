package vpn

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/test/errors"
	testtunnel "github.com/NordSecurity/nordvpn-linux/test/tunnel"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// Working stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type Working struct {
	isActive bool
	StartErr error
}

func (w *Working) Start(vpn.Credentials, vpn.ServerData) error {
	w.isActive = w.StartErr == nil
	return w.StartErr
}
func (w *Working) Stop() error    { w.isActive = false; return nil }
func (*Working) State() vpn.State { return vpn.ConnectedState }
func (w *Working) IsActive() bool { return w.isActive }
func (*Working) Tun() tunnel.T    { return testtunnel.Working{} }

type WorkingInactive struct{}

func (WorkingInactive) Start(vpn.Credentials, vpn.ServerData) error { return nil }
func (WorkingInactive) Stop() error                                 { return nil }
func (WorkingInactive) State() vpn.State                            { return vpn.ConnectedState }
func (WorkingInactive) IsActive() bool                              { return false }
func (WorkingInactive) Tun() tunnel.T                               { return testtunnel.Working{} }

// Failing stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type Failing struct{}

func (Failing) Start(vpn.Credentials, vpn.ServerData) error { return errors.ErrOnPurpose }
func (Failing) Stop() error                                 { return errors.ErrOnPurpose }
func (Failing) State() vpn.State                            { return vpn.ExitedState }
func (Failing) IsActive() bool                              { return false }
func (Failing) Tun() tunnel.T                               { return testtunnel.Working{} }
