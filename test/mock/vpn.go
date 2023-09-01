package mock

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// WorkingVPN stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type WorkingVPN struct {
	isActive bool
	StartErr error
}

func (w *WorkingVPN) Start(vpn.Credentials, vpn.ServerData) error {
	w.isActive = w.StartErr == nil
	return w.StartErr
}
func (w *WorkingVPN) Stop() error    { w.isActive = false; return nil }
func (*WorkingVPN) State() vpn.State { return vpn.ConnectedState }
func (w *WorkingVPN) IsActive() bool { return w.isActive }
func (*WorkingVPN) Tun() tunnel.T    { return WorkingT{} }

type WorkingInactiveVPN struct{}

func (WorkingInactiveVPN) Start(vpn.Credentials, vpn.ServerData) error { return nil }
func (WorkingInactiveVPN) Stop() error                                 { return nil }
func (WorkingInactiveVPN) State() vpn.State                            { return vpn.ConnectedState }
func (WorkingInactiveVPN) IsActive() bool                              { return false }
func (WorkingInactiveVPN) Tun() tunnel.T                               { return WorkingT{} }

// FailingVPN stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type FailingVPN struct{}

func (FailingVPN) Start(vpn.Credentials, vpn.ServerData) error { return ErrOnPurpose }
func (FailingVPN) Stop() error                                 { return ErrOnPurpose }
func (FailingVPN) State() vpn.State                            { return vpn.ExitedState }
func (FailingVPN) IsActive() bool                              { return false }
func (FailingVPN) Tun() tunnel.T                               { return WorkingT{} }
