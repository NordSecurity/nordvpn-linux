package mock

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// WorkingVPN stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type WorkingVPN struct {
	isActive     bool
	StartErr     error
	StateChannel chan vpn.State
	sync.Mutex
}

func (w *WorkingVPN) Start(vpn.Credentials, vpn.ServerData) error {
	w.Lock()
	defer w.Unlock()

	w.isActive = w.StartErr == nil
	go func() {
		w.StateChannel <- vpn.ConnectedState
	}()

	return w.StartErr
}
func (w *WorkingVPN) Stop() error    { w.isActive = false; return nil }
func (*WorkingVPN) State() vpn.State { return vpn.ConnectedState }
func (w *WorkingVPN) IsActive() bool { return w.isActive }
func (*WorkingVPN) Tun() tunnel.T    { return WorkingT{} }
func (w *WorkingVPN) StateChanged() <-chan vpn.State {
	w.Lock()
	defer w.Unlock()

	if w.StateChannel == nil {
		w.StateChannel = make(chan vpn.State)
	}
	return w.StateChannel
}

type WorkingInactiveVPN struct{}

func (WorkingInactiveVPN) Start(vpn.Credentials, vpn.ServerData) error { return nil }
func (WorkingInactiveVPN) Stop() error                                 { return nil }
func (WorkingInactiveVPN) State() vpn.State                            { return vpn.ConnectedState }
func (WorkingInactiveVPN) IsActive() bool                              { return false }
func (WorkingInactiveVPN) Tun() tunnel.T                               { return WorkingT{} }
func (WorkingInactiveVPN) StateChanged() <-chan vpn.State              { return nil }

// FailingVPN stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type FailingVPN struct{}

func (FailingVPN) Start(vpn.Credentials, vpn.ServerData) error { return ErrOnPurpose }
func (FailingVPN) Stop() error                                 { return ErrOnPurpose }
func (FailingVPN) State() vpn.State                            { return vpn.ExitedState }
func (FailingVPN) IsActive() bool                              { return false }
func (FailingVPN) Tun() tunnel.T                               { return FailingTunnel{} }
func (FailingVPN) StateChanged() <-chan vpn.State              { return nil }

// ActiveVPN stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type ActiveVPN struct {
	StateChannel chan vpn.State
	sync.Mutex
}

func (a *ActiveVPN) Start(vpn.Credentials, vpn.ServerData) error {
	a.Lock()
	defer a.Unlock()
	go func() {
		a.StateChannel <- vpn.ConnectedState
	}()
	return nil
}
func (*ActiveVPN) Stop() error      { return nil }
func (*ActiveVPN) State() vpn.State { return vpn.ExitedState }
func (*ActiveVPN) IsActive() bool   { return true }
func (*ActiveVPN) Tun() tunnel.T    { return WorkingT{} }
func (a *ActiveVPN) StateChanged() <-chan vpn.State {
	a.Lock()
	defer a.Unlock()

	if a.StateChannel == nil {
		a.StateChannel = make(chan vpn.State)
	}
	return a.StateChannel
}
