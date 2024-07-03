package mock

import (
	"net/netip"

	teliogo "github.com/NordSecurity/libtelio-go/v5"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	_ "github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx/libtelio/symbols" // required for linking process
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

const (
	StatsStart         = iota
	StatsStop          = iota
	StatsNetworkChange = iota
	statsLastValue     = iota
)

// WorkingVPN stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type WorkingVPN struct {
	isActive          bool
	StartErr          error
	ErrNetworkChanges error
	ExecutionStats    [statsLastValue]int
}

func (w *WorkingVPN) Start(vpn.Credentials, vpn.ServerData) error {
	w.ExecutionStats[StatsStart]++

	w.isActive = w.StartErr == nil
	return w.StartErr
}

func (w *WorkingVPN) Stop() error {
	w.ExecutionStats[StatsStop]++

	if !w.isActive {
		return w.StartErr
	}

	w.isActive = false
	return nil
}
func (w *WorkingVPN) State() vpn.State { return vpn.ConnectedState }
func (w *WorkingVPN) IsActive() bool   { return w.isActive }
func (*WorkingVPN) Tun() tunnel.T      { return WorkingT{} }
func (w *WorkingVPN) NetworkChanged() error {
	w.ExecutionStats[StatsNetworkChange]++

	return w.ErrNetworkChanges
}

type WorkingInactiveVPN struct{}

func (WorkingInactiveVPN) Start(vpn.Credentials, vpn.ServerData) error { return nil }
func (WorkingInactiveVPN) Stop() error                                 { return nil }
func (WorkingInactiveVPN) State() vpn.State                            { return vpn.ConnectedState }
func (WorkingInactiveVPN) IsActive() bool                              { return false }
func (WorkingInactiveVPN) Tun() tunnel.T                               { return WorkingT{} }
func (WorkingInactiveVPN) NetworkChanged() error                       { return nil }

// FailingVPN stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type FailingVPN struct{}

func (FailingVPN) Start(vpn.Credentials, vpn.ServerData) error { return ErrOnPurpose }
func (FailingVPN) Stop() error                                 { return ErrOnPurpose }
func (FailingVPN) State() vpn.State                            { return vpn.ExitedState }
func (FailingVPN) IsActive() bool                              { return false }
func (FailingVPN) Tun() tunnel.T                               { return WorkingT{} }
func (FailingVPN) NetworkChanged() error                       { return ErrOnPurpose }

// ActiveVPN stub of a github.com/NordSecurity/nordvpn-linux/daemon/vpn.VPN interface.
type ActiveVPN struct{}

func (ActiveVPN) Start(vpn.Credentials, vpn.ServerData) error { return nil }
func (ActiveVPN) Stop() error                                 { return nil }
func (ActiveVPN) State() vpn.State                            { return vpn.ExitedState }
func (ActiveVPN) IsActive() bool                              { return true }
func (ActiveVPN) Tun() tunnel.T                               { return WorkingT{} }
func (ActiveVPN) NetworkChanged() error                       { return nil }

type MeshnetAndVPN struct {
	WorkingVPN
	MeshEnableError error
}

func (w *MeshnetAndVPN) Enable(netip.Addr, string) error { return w.MeshEnableError }
func (*MeshnetAndVPN) Disable() error                    { return nil }
func (*MeshnetAndVPN) Refresh(mesh.MachineMap) error     { return nil }
func (*MeshnetAndVPN) StatusMap() (map[string]teliogo.NodeState, error) {
	return map[string]teliogo.NodeState{}, nil
}
