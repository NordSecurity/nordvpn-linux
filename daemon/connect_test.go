package daemon

import (
	"errors"
	"net"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet/exitnode"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/tunnel"

	"github.com/stretchr/testify/assert"
)

var (
	errOnPurpose = errors.New("on purpose")
)

type workingRouter struct{}

func (workingRouter) Add(route routes.Route) error { return nil }
func (workingRouter) Flush() error                 { return nil }

type mockGateway struct{ ip netip.Addr }

func newGatewayMock(ip netip.Addr) mockGateway {
	return mockGateway{ip: ip}
}

func (g mockGateway) Default(bool) (netip.Addr, net.Interface, error) {
	return g.ip, net.Interface{Name: "noname"}, nil
}

type mockEndpointResolver struct{ ip netip.Addr }

func newEndpointResolverMock(ip netip.Addr) mockEndpointResolver {
	return mockEndpointResolver{ip: ip}
}

func (g mockEndpointResolver) Resolve(netip.Addr) ([]netip.Addr, error) {
	return []netip.Addr{g.ip}, nil
}

type workingFirewall struct{}

func (workingFirewall) Add([]firewall.Rule) error { return nil }
func (workingFirewall) Delete([]string) error     { return nil }
func (workingFirewall) Enable() error             { return nil }
func (workingFirewall) Disable() error            { return nil }
func (workingFirewall) IsEnabled() bool           { return true }

type workingTunnel struct{}

func (workingTunnel) Interface() net.Interface { return en0Interface }
func (workingTunnel) IPs() []netip.Addr {
	return []netip.Addr{netip.MustParseAddr("172.105.90.114")}
}

func (workingTunnel) TransferRates() (tunnel.Statistics, error) {
	return tunnel.Statistics{Tx: 1337, Rx: 1337}, nil
}

type failingTunnel struct{}

func (failingTunnel) Interface() net.Interface { return net.Interface{} }
func (failingTunnel) IPs() []netip.Addr        { return nil }
func (failingTunnel) TransferRates() (tunnel.Statistics, error) {
	return tunnel.Statistics{}, errOnPurpose
}

type workingVPN struct{}

func (workingVPN) Start(
	vpn.Credentials,
	vpn.ServerData,
) error {
	return nil
}
func (workingVPN) Stop() error      { return nil }
func (workingVPN) State() vpn.State { return vpn.ConnectedState }
func (workingVPN) IsActive() bool   { return true }
func (workingVPN) Tun() tunnel.T    { return workingTunnel{} }

type failingVPN struct{}

func (failingVPN) Start(
	vpn.Credentials,
	vpn.ServerData,
) error {
	return errOnPurpose
}
func (failingVPN) Stop() error      { return errOnPurpose }
func (failingVPN) State() vpn.State { return vpn.ExitedState }
func (failingVPN) IsActive() bool   { return false }
func (failingVPN) Tun() tunnel.T    { return failingTunnel{} }

type workingNetworker struct{}

func (workingNetworker) Start(
	vpn.Credentials,
	vpn.ServerData,
	config.Allowlist,
	config.DNS,
) error {
	return nil
}
func (workingNetworker) Stop() error           { return nil }
func (workingNetworker) UnSetMesh() error      { return nil }
func (workingNetworker) SetDNS([]string) error { return nil }
func (workingNetworker) UnsetDNS() error       { return nil }
func (workingNetworker) IsVPNActive() bool     { return true }
func (workingNetworker) IsMeshnetActive() bool { return true }
func (workingNetworker) ConnectionStatus() (networker.ConnectionStatus, error) {
	return networker.ConnectionStatus{}, nil
}

func (workingNetworker) EnableFirewall() error                               { return nil }
func (workingNetworker) DisableFirewall() error                              { return nil }
func (workingNetworker) EnableRouting()                                      {}
func (workingNetworker) DisableRouting()                                     {}
func (workingNetworker) PermitIPv6() error                                   { return nil }
func (workingNetworker) DenyIPv6() error                                     { return nil }
func (workingNetworker) SetAllowlist(config.Allowlist) error                 { return nil }
func (workingNetworker) UnsetAllowlist() error                               { return nil }
func (workingNetworker) IsNetworkSet() bool                                  { return false }
func (workingNetworker) SetKillSwitch(config.Allowlist) error                { return nil }
func (workingNetworker) UnsetKillSwitch() error                              { return nil }
func (workingNetworker) Connect(netip.Addr, string) error                    { return nil }
func (workingNetworker) Disconnect() error                                   { return nil }
func (workingNetworker) Refresh(mesh.MachineMap) error                       { return nil }
func (workingNetworker) Allow(mesh.Machine) error                            { return nil }
func (workingNetworker) Block(mesh.Machine) error                            { return nil }
func (workingNetworker) SetVPN(vpn.VPN, exitnode.MasqueradeSetter)           {}
func (workingNetworker) LastServerName() string                              { return "" }
func (workingNetworker) SetLanDiscoveryAndResetMesh(bool, mesh.MachinePeers) {}
func (workingNetworker) SetLanDiscovery(bool)                                {}

type UniqueAddress struct{}

type failingNetworker struct{}

func (failingNetworker) Start(
	vpn.Credentials,
	vpn.ServerData,
	config.Allowlist,
	config.DNS,
) error {
	return errOnPurpose
}
func (failingNetworker) Stop() error           { return errOnPurpose }
func (failingNetworker) UnSetMesh() error      { return errOnPurpose }
func (failingNetworker) SetDNS([]string) error { return errOnPurpose }
func (failingNetworker) UnsetDNS() error       { return errOnPurpose }
func (failingNetworker) IsVPNActive() bool     { return false }
func (failingNetworker) IsMeshnetActive() bool { return false }
func (failingNetworker) ConnectionStatus() (networker.ConnectionStatus, error) {
	return networker.ConnectionStatus{}, nil
}

func (failingNetworker) EnableFirewall() error                               { return errOnPurpose }
func (failingNetworker) DisableFirewall() error                              { return errOnPurpose }
func (failingNetworker) EnableRouting()                                      {}
func (failingNetworker) DisableRouting()                                     {}
func (failingNetworker) PermitIPv6() error                                   { return errOnPurpose }
func (failingNetworker) DenyIPv6() error                                     { return errOnPurpose }
func (failingNetworker) SetAllowlist(config.Allowlist) error                 { return errOnPurpose }
func (failingNetworker) UnsetAllowlist() error                               { return errOnPurpose }
func (failingNetworker) IsNetworkSet() bool                                  { return false }
func (failingNetworker) SetKillSwitch(config.Allowlist) error                { return errOnPurpose }
func (failingNetworker) UnsetKillSwitch() error                              { return errOnPurpose }
func (failingNetworker) Connect(netip.Addr, string) error                    { return errOnPurpose }
func (failingNetworker) Disconnect() error                                   { return errOnPurpose }
func (failingNetworker) Refresh(mesh.MachineMap) error                       { return errOnPurpose }
func (failingNetworker) Allow(mesh.Machine) error                            { return errOnPurpose }
func (failingNetworker) Block(mesh.Machine) error                            { return errOnPurpose }
func (failingNetworker) SetVPN(vpn.VPN, exitnode.MasqueradeSetter)           {}
func (failingNetworker) LastServerName() string                              { return "" }
func (failingNetworker) SetLanDiscoveryAndResetMesh(bool, mesh.MachinePeers) {}
func (failingNetworker) SetLanDiscovery(bool)                                {}

func TestConnect(t *testing.T) {
	category.Set(t, category.Route)

	tests := []struct {
		name        string
		netw        networker.Networker
		fw          firewall.Service
		allowlist   config.Allowlist
		router      routes.Agent
		gateway     routes.GatewayRetriever
		nameservers []string
		expected    ConnectEvent
	}{
		{
			name:      "successful connect",
			netw:      workingNetworker{},
			fw:        &workingFirewall{},
			allowlist: config.NewAllowlist(nil, nil, nil),
			router:    &workingRouter{},
			gateway:   newGatewayMock(netip.Addr{}),
			expected:  ConnectEvent{Code: internal.CodeConnected},
		},
		{
			name:     "successful reconnect",
			netw:     workingNetworker{},
			fw:       &workingFirewall{},
			router:   &workingRouter{},
			gateway:  newGatewayMock(netip.Addr{}),
			expected: ConnectEvent{Code: internal.CodeConnected},
		},
		{
			name:     "failed connect",
			netw:     failingNetworker{},
			fw:       &workingFirewall{},
			router:   &workingRouter{},
			gateway:  newGatewayMock(netip.Addr{}),
			expected: ConnectEvent{Code: internal.CodeFailure, Message: "on purpose"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			channel := make(chan ConnectEvent)
			go Connect(
				channel,
				vpn.Credentials{},
				vpn.ServerData{},
				test.allowlist,
				test.nameservers,
				test.netw,
			)
			assert.Equal(t, ConnectEvent{Code: internal.CodeConnecting}, <-channel)
			assert.Equal(t, test.expected, <-channel)
		})
	}
}

var en0Interface = net.Interface{
	Index:        1,
	MTU:          5,
	Name:         "en0",
	HardwareAddr: []byte("00:00:5e:00:53:01"),
	Flags:        net.FlagMulticast,
}

func TestMaskIPRouteOutput(t *testing.T) {
	input4 := `default dev nordtun table 205 scope link\n
default via 180.144.168.176 dev wlp0s20f3 proto dhcp metric 20600\n
172.31.100.100/24 dev nordtun proto kernel scope link src 192.168.200.203\n
114.237.30.247/16 dev wlp0s20f3 scope link metric 1000\n
local 10.128.10.7 dev wlp0s20f3 table local proto kernel scope link src 26.14.182.220`

	maskedInput4 := maskIPRouteOutput(input4)

	expectedOutput4 := `default dev nordtun table 205 scope link\n
default via *** dev wlp0s20f3 proto dhcp metric 20600\n
172.31.100.100/24 dev nordtun proto kernel scope link src 192.168.200.203\n
***/16 dev wlp0s20f3 scope link metric 1000\n
local 10.128.10.7 dev wlp0s20f3 table local proto kernel scope link src ***`

	assert.Equal(t, expectedOutput4, maskedInput4)

	input6 := `default dev nordtun table 205 scope link\n
	default via fd31:482b:86d9:7142::1 dev wlp0s20f3 proto dhcp metric 20600\n
	8d02:d70f:76b4:162e:d12f:b0e6:204a:59d1 dev nordtun proto kernel scope link src 24ef:7163:ffd8:4ee7:16f8:008b:e52b:0a68\n
	1e66:9f56:66b5:b846:8d27:d0b5:0821:c819 dev wlp0s20f3 scope link metric 1000\n
	local fc81:9a6e:dcf2:20a7::2 dev wlp0s20f3 table local proto kernel scope link src fdf3:cbf9:573c:8e15::3`

	maskedInput6 := maskIPRouteOutput(input6)

	expectedOutput6 := `default dev nordtun table 205 scope link\n
	default via fd31:482b:86d9:7142::1 dev wlp0s20f3 proto dhcp metric 20600\n
	*** dev nordtun proto kernel scope link src ***\n
	*** dev wlp0s20f3 scope link metric 1000\n
	local fc81:9a6e:dcf2:20a7::2 dev wlp0s20f3 table local proto kernel scope link src fdf3:cbf9:573c:8e15::3`

	assert.Equal(t, expectedOutput6, maskedInput6)
}
