package networker

// Separate package is used due to circular dependency issues

import (
	"context"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
)

type Mock struct {
	Dns               []string
	Allowlist         config.Allowlist
	VpnActive         bool
	MeshActive        bool
	ConnectRetries    int
	LanDiscovery      bool
	MeshPeers         mesh.MachinePeers
	MeshnetRetries    int
	SetDNSErr         error
	SetAllowlistErr   error
	UnsetAllowlistErr error
}

func (Mock) Start(
	context.Context,
	vpn.Credentials,
	vpn.ServerData,
	config.Allowlist,
	config.DNS,
	bool,
) error {
	return nil
}
func (*Mock) Stop() error      { return nil }
func (*Mock) UnSetMesh() error { return nil }

func (m *Mock) SetDNS(nameservers []string) error {
	m.Dns = nameservers
	return m.SetDNSErr
}

func (*Mock) UnsetDNS() error { return nil }

func (m *Mock) IsVPNActive() bool {
	m.ConnectRetries++
	return m.VpnActive || m.ConnectRetries > 5
}

func (*Mock) ConnectionStatus() networker.ConnectionStatus {
	return networker.ConnectionStatus{}
}

func (*Mock) EnableFirewall() error  { return nil }
func (*Mock) DisableFirewall() error { return nil }
func (*Mock) EnableRouting()         {}
func (*Mock) DisableRouting()        {}

func (m *Mock) SetAllowlist(allowlist config.Allowlist) error {
	if m.SetAllowlistErr != nil {
		return m.SetAllowlistErr
	}

	m.Allowlist = allowlist
	return nil
}

func (m *Mock) UnsetAllowlist() error {
	if m.UnsetAllowlistErr != nil {
		return m.UnsetAllowlistErr
	}

	m.Allowlist.Ports.TCP = make(config.PortSet)
	m.Allowlist.Ports.UDP = make(config.PortSet)
	m.Allowlist.Subnets = make(config.Subnets)
	return nil
}

func (*Mock) IsNetworkSet() bool { return false }
func (m *Mock) IsMeshnetActive() bool {
	m.MeshnetRetries++
	return m.MeshActive || m.MeshnetRetries > 5
}
func (*Mock) SetKillSwitch(config.Allowlist) error { return nil }
func (*Mock) UnsetKillSwitch() error               { return nil }
func (*Mock) PermitIPv6() error                    { return nil }
func (*Mock) DenyIPv6() error                      { return nil }
func (*Mock) SetVPN(vpn.VPN)                       {}
func (*Mock) LastServerName() string               { return "" }

func (m *Mock) SetLanDiscoveryAndResetMesh(enabled bool, peers mesh.MachinePeers) {
	m.MeshPeers = peers
	m.LanDiscovery = enabled
}

func (m *Mock) SetLanDiscovery(enabled bool) {
	m.LanDiscovery = enabled
}

func (*Mock) UnsetFirewall() error { return nil }

func (*Mock) GetConnectionParameters() (vpn.ServerData, bool) { return vpn.ServerData{}, false }

type Failing struct{}

func (Failing) Start(
	context.Context,
	vpn.Credentials,
	vpn.ServerData,
	config.Allowlist,
	config.DNS,
	bool,
) error {
	return mock.ErrOnPurpose
}
func (Failing) Stop() error           { return mock.ErrOnPurpose }
func (Failing) UnSetMesh() error      { return mock.ErrOnPurpose }
func (Failing) SetDNS([]string) error { return mock.ErrOnPurpose }
func (Failing) UnsetDNS() error       { return mock.ErrOnPurpose }
func (Failing) IsVPNActive() bool     { return false }
func (Failing) IsMeshnetActive() bool { return false }
func (Failing) ConnectionStatus() networker.ConnectionStatus {
	return networker.ConnectionStatus{}
}

func (Failing) EnableFirewall() error                               { return mock.ErrOnPurpose }
func (Failing) DisableFirewall() error                              { return mock.ErrOnPurpose }
func (Failing) EnableRouting()                                      {}
func (Failing) DisableRouting()                                     {}
func (Failing) PermitIPv6() error                                   { return mock.ErrOnPurpose }
func (Failing) DenyIPv6() error                                     { return mock.ErrOnPurpose }
func (Failing) SetAllowlist(config.Allowlist) error                 { return mock.ErrOnPurpose }
func (Failing) UnsetAllowlist() error                               { return mock.ErrOnPurpose }
func (Failing) IsNetworkSet() bool                                  { return false }
func (Failing) SetKillSwitch(config.Allowlist) error                { return mock.ErrOnPurpose }
func (Failing) UnsetKillSwitch() error                              { return mock.ErrOnPurpose }
func (Failing) Connect(netip.Addr, string) error                    { return mock.ErrOnPurpose }
func (Failing) Disconnect() error                                   { return mock.ErrOnPurpose }
func (Failing) Refresh(mesh.MachineMap) error                       { return mock.ErrOnPurpose }
func (Failing) Allow(mesh.Machine) error                            { return mock.ErrOnPurpose }
func (Failing) Block(mesh.Machine) error                            { return mock.ErrOnPurpose }
func (Failing) SetVPN(vpn.VPN)                                      {}
func (Failing) LastServerName() string                              { return "" }
func (Failing) SetLanDiscoveryAndResetMesh(bool, mesh.MachinePeers) {}
func (Failing) SetLanDiscovery(bool)                                {}
func (Failing) UnsetFirewall() error                                { return mock.ErrOnPurpose }
func (Failing) GetConnectionParameters() (vpn.ServerData, bool)     { return vpn.ServerData{}, false }
