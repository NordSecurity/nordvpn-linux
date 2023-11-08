package networker

import (
	"net"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/allowlist"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testfirewall "github.com/NordSecurity/nordvpn-linux/test/mock/firewall"
	"github.com/NordSecurity/nordvpn-linux/tunnel"

	"github.com/stretchr/testify/assert"
)

// GetTestCombined returns Combined networker, with all of the possible components initialized to the mocked 'working'
// variants. lanDiscovery is initialized to false(disabled) and connmark initalized to 0.
func GetTestCombined(t *testing.T) *Combined {
	t.Helper()
	return NewCombined(
		&mock.WorkingVPN{},
		&workingMesh{},
		workingGateway{},
		&subs.Subject[string]{},
		workingRouter{},
		&workingDNS{},
		&workingIpv6{},
		newMockFirewallManager(t, workingDeviceList, nil),
		workingAllowlistRouting{},
		&workingRoutingSetup{},
		&workingHostSetter{},
		workingRouter{},
		workingRouter{},
		&workingExitNode{},
		0,
		false,
	)
}

type workingGateway struct{}

func (w workingGateway) Default(bool) (netip.Addr, net.Interface, error) {
	return netip.MustParseAddr("1.1.1.1"), mock.En0Interface, nil
}

type workingRouter struct{}

func (workingRouter) Add(routes.Route) error { return nil }
func (workingRouter) Flush() error           { return nil }
func (workingRouter) Enable(uint) error      { return nil }
func (workingRouter) Disable() error         { return nil }
func (workingRouter) IsEnabled() bool        { return true }

type failingRouter struct{}

func (failingRouter) Add(routes.Route) error { return mock.ErrOnPurpose }
func (failingRouter) Flush() error           { return mock.ErrOnPurpose }
func (failingRouter) Enable(uint) error      { return mock.ErrOnPurpose }
func (failingRouter) Disable() error         { return mock.ErrOnPurpose }
func (failingRouter) IsEnabled() bool        { return false }

type workingDNS struct{ setDNS []string }

func (w *workingDNS) Set(_ string, dns []string) error { w.setDNS = dns; return nil }
func (w *workingDNS) Unset(string) error               { w.setDNS = nil; return nil }

type failingDNS struct{}

func (failingDNS) Set(string, []string) error { return mock.ErrOnPurpose }
func (failingDNS) Unset(string) error         { return mock.ErrOnPurpose }

type workingIpv6 struct{}

func (workingIpv6) Block() error   { return nil }
func (workingIpv6) Unblock() error { return nil }

type workingAllowlistRouting struct{}

func (workingAllowlistRouting) EnablePorts([]int, string, string) error    { return nil }
func (workingAllowlistRouting) EnableSubnets([]netip.Prefix, string) error { return nil }
func (workingAllowlistRouting) Disable() error                             { return nil }

func workingDeviceList() ([]net.Interface, error) {
	return []net.Interface{mock.En0Interface}, nil
}

func failingDeviceList() ([]net.Interface, error) { return nil, mock.ErrOnPurpose }

type workingRoutingSetup struct {
	EnableLocalTraffic bool
}

func (r *workingRoutingSetup) SetupRoutingRules(_ net.Interface, _ bool, enableLan bool) error {
	r.EnableLocalTraffic = enableLan
	return nil
}
func (*workingRoutingSetup) CleanupRouting() error { return nil }
func (*workingRoutingSetup) TableID() uint         { return 0 }
func (*workingRoutingSetup) Enable() error         { return nil }
func (*workingRoutingSetup) Disable() error        { return nil }
func (*workingRoutingSetup) IsEnabled() bool       { return true }

type workingExitNode struct {
	enabled      bool
	peers        mesh.MachinePeers
	LanAvailable bool
}

func newWorkingExitNode() *workingExitNode {
	return &workingExitNode{
		peers: mesh.MachinePeers{},
	}
}

func (e *workingExitNode) Enable() error {
	e.enabled = true
	return nil
}

func (e *workingExitNode) ResetPeers(peers mesh.MachinePeers, lan bool) error {
	e.peers = peers
	e.LanAvailable = lan
	return nil
}

func (*workingExitNode) DisablePeer(netip.Addr) error { return nil }
func (*workingExitNode) Disable() error               { return nil }
func (e *workingExitNode) SetAllowlist(_ config.Allowlist, lan bool) error {
	e.LanAvailable = lan
	return nil
}
func (e *workingExitNode) ResetFirewall(lan bool) error { e.LanAvailable = lan; return nil }

type workingMesh struct {
	enableErr error
}

func (w *workingMesh) Enable(netip.Addr, string) error { return w.enableErr }
func (*workingMesh) Disable() error                    { return nil }
func (*workingMesh) IsActive() bool                    { return false }
func (*workingMesh) Refresh(mesh.MachineMap) error     { return nil }
func (*workingMesh) Tun() tunnel.T                     { return mock.WorkingT{} }
func (*workingMesh) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}

type workingHostSetter struct {
	hosts dns.Hosts
}

func newMockHostSetter() *workingHostSetter {
	return &workingHostSetter{
		hosts: dns.Hosts{},
	}
}

func (h *workingHostSetter) SetHosts(hosts dns.Hosts) error {
	h.hosts = hosts
	return nil
}

func (h *workingHostSetter) UnsetHosts() error {
	h.hosts = dns.Hosts{}
	return nil
}

// newMockFirewallManager returns a mocked instance of firewall.FirewallManager.
// If deviceList is nil, a default function that always returns ([]net.Interface{}, nil) will be used.
// If iptables is nil, test.IptablesMock will be used as IptablesExecutor implementation.
func newMockFirewallManager(t *testing.T, deviceList func() ([]net.Interface, error), iptables firewall.IptablesExecutor) firewall.FirewallManager {
	t.Helper()
	if deviceList == nil {
		deviceList = func() ([]net.Interface, error) { return []net.Interface{}, nil }
	}

	if iptables == nil {
		iptablesMock := testfirewall.NewIptablesMock(false)
		iptables = &iptablesMock
	}
	return firewall.NewFirewallManager(deviceList, iptables, 0, true)
}

// newMockFailingFirewallManager returns a mocked instance of firewall.FirewallManager, same as newMockFirewallManager.
// ErrIptablesFailure is returned by testfirewall.IptablesMock which should be propagated by the FirewallManager.
func newMockFailingFirewallManager(t *testing.T, deviceList func() ([]net.Interface, error)) firewall.FirewallManager {
	t.Helper()
	if deviceList == nil {
		deviceList = func() ([]net.Interface, error) { return []net.Interface{}, nil }
	}

	iptablesMock := testfirewall.NewIptablesMock(true)

	return firewall.NewFirewallManager(deviceList, &iptablesMock, 0, true)
}

func TestCombined_Start(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		gateway         routes.GatewayRetriever
		allowlistRouter routes.Service
		dns             dns.Setter
		firewall        firewall.FirewallManager
		vpn             vpn.VPN
		allowlist       allowlist.Routing
		routing         routes.PolicyService
		err             error
	}{
		{
			name:            "nil vpn",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			firewall:        newMockFirewallManager(t, workingDeviceList, nil),
			vpn:             nil,
			allowlist:       &workingAllowlistRouting{},
			routing:         &workingRoutingSetup{},
			err:             errNilVPN,
		},
		{
			name:            "vpn start failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			firewall:        newMockFirewallManager(t, workingDeviceList, nil),
			vpn:             mock.FailingVPN{},
			allowlist:       &workingAllowlistRouting{},
			routing:         &workingRoutingSetup{},
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "firewall failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			firewall:        newMockFailingFirewallManager(t, workingDeviceList),
			vpn:             mock.WorkingInactiveVPN{},
			allowlist:       &workingAllowlistRouting{},
			routing:         &workingRoutingSetup{},
			err:             testfirewall.ErrIptablesFailure,
		},
		{
			name:            "dns failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             failingDNS{},
			firewall:        newMockFirewallManager(t, workingDeviceList, nil),
			vpn:             mock.WorkingInactiveVPN{},
			allowlist:       &workingAllowlistRouting{},
			routing:         &workingRoutingSetup{},
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "device listing failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			firewall:        newMockFirewallManager(t, failingDeviceList, nil),
			vpn:             mock.WorkingInactiveVPN{},
			allowlist:       &workingAllowlistRouting{},
			routing:         &workingRoutingSetup{},
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "successful start",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			firewall:        newMockFirewallManager(t, workingDeviceList, nil),
			vpn:             &mock.WorkingVPN{},
			allowlist:       &workingAllowlistRouting{},
			routing:         &workingRoutingSetup{},
			err:             nil,
		},
		{
			name:            "restart",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			firewall:        newMockFirewallManager(t, workingDeviceList, nil),
			vpn:             &mock.ActiveVPN{},
			allowlist:       &workingAllowlistRouting{},
			routing:         &workingRoutingSetup{},
			err:             nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				test.vpn,
				nil,
				test.gateway,
				&subs.Subject[string]{},
				test.allowlistRouter,
				test.dns,
				&workingIpv6{},
				test.firewall,
				test.allowlist,
				test.routing,
				nil,
				workingRouter{},
				nil,
				&workingExitNode{},
				0,
				false,
			)
			err := netw.Start(
				vpn.Credentials{},
				vpn.ServerData{},
				config.NewAllowlist(nil, nil, nil),
				[]string{"1.1.1.1"},
				true,
			)
			assert.ErrorIs(t, err, test.err, test.name)
		})
	}
}

func TestCombined_Stop(t *testing.T) {
	category.Set(t, category.Link)

	tests := []struct {
		name string
		vpn  vpn.VPN
		dns  dns.Setter
		err  error
	}{
		{
			name: "nil vpn",
			vpn:  nil,
			dns:  &workingDNS{},
			err:  errNilVPN,
		},
		{
			name: "unset dns failure",
			vpn:  &mock.WorkingVPN{},
			dns:  failingDNS{},
			err:  mock.ErrOnPurpose,
		},
		{
			name: "vpn stop failure",
			vpn:  mock.FailingVPN{},
			dns:  &workingDNS{},
			err:  mock.ErrOnPurpose,
		},
		{
			name: "successful stop",
			vpn:  &mock.WorkingVPN{},
			dns:  &workingDNS{},
			err:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				test.vpn,
				nil,
				workingGateway{},
				&subs.Subject[string]{},
				workingRouter{},
				test.dns,
				&workingIpv6{},
				newMockFirewallManager(t, nil, nil),
				workingAllowlistRouting{},
				&workingRoutingSetup{},
				nil,
				workingRouter{},
				nil,
				&workingExitNode{},
				0,
				false,
			)
			netw.vpnet = test.vpn
			err := netw.stop()
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestCombined_TransferRates(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		vpn      vpn.VPN
		err      error
		expected tunnel.Statistics
	}{
		{
			name:     "active vpn",
			vpn:      activeVPN{},
			expected: tunnel.Statistics{Tx: 1337, Rx: 1337},
		},
		{
			name: "inactive vpn",
			vpn:  inactiveVPN{},
			err:  errInactiveVPN,
		},
		{
			name: "nil vpn",
			err:  errInactiveVPN,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test does not rely on any of the values provided via constructor
			// so it's fine to pass nils to all of them.
			netw := NewCombined(nil, nil, nil, nil, nil, nil, nil, newMockFirewallManager(t, nil, nil), nil, nil, nil, nil, nil, nil, 0, false)
			// injecting VPN implementation without calling netw.Start
			netw.vpnet = test.vpn
			connStus, err := netw.ConnectionStatus()
			stats := tunnel.Statistics{Tx: connStus.Upload, Rx: connStus.Download}
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.expected, stats)
		})
	}
}

func TestCombined_SetDNS(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		dns         dns.Setter
		nameservers []string
		hasError    bool
	}{
		{
			name:        "empty nameservers",
			dns:         &workingDNS{},
			nameservers: []string{},
			hasError:    false,
		},
		{
			name:        "nil nameservers",
			dns:         &workingDNS{},
			nameservers: nil,
			hasError:    false,
		},
		{
			name:        "two nameservers",
			dns:         &workingDNS{},
			nameservers: []string{"103.86.96.100", "103.86.99.100"},
			hasError:    false,
		},
		{
			name:        "failing setter",
			dns:         failingDNS{},
			nameservers: []string{"103.86.96.100", "103.86.99.100"},
			hasError:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				nil,
				workingGateway{},
				&subs.Subject[string]{},
				workingRouter{},
				test.dns,
				&workingIpv6{},
				newMockFirewallManager(t, nil, nil),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			netw.vpnet = &mock.WorkingVPN{}
			err := netw.setDNS(test.nameservers)
			assert.Equal(t, test.hasError, err != nil)
		})
	}
}

func TestCombined_UnsetDNS(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		dns      dns.Setter
		hasError bool
	}{
		{
			name:     "failing unsetter",
			dns:      failingDNS{},
			hasError: true,
		},
		{
			name:     "success unset",
			dns:      &workingDNS{},
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				&mock.ActiveVPN{},
				nil,
				workingGateway{},
				&subs.Subject[string]{},
				workingRouter{},
				test.dns,
				&workingIpv6{},
				newMockFirewallManager(t, nil, nil),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			err := netw.UnsetDNS()
			assert.Equal(t, test.hasError, err != nil)
		})
	}
}

func TestCombined_ResetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	networker := GetTestCombined(t)
	networker.firewallManager = newMockFirewallManager(t, workingDeviceList, nil)

	assert.Nil(t, networker.resetAllowlist())
}

func TestCombined_ResetAllowlist_DeviceListingFailure(t *testing.T) {
	category.Set(t, category.Unit)

	networker := GetTestCombined(t)
	networker.firewallManager = newMockFirewallManager(t, failingDeviceList, nil)

	assert.ErrorIs(t, networker.resetAllowlist(), mock.ErrOnPurpose)
}

func TestCombined_ResetAllowlist_FirewallFailure(t *testing.T) {
	category.Set(t, category.Unit)

	iptablesExecutor := testfirewall.NewIptablesMock(false)
	firewall := newMockFirewallManager(t, workingDeviceList, &iptablesExecutor)

	// we need to add some rules so that the can be removed for testing.
	firewall.SetAllowlist([]int{5000}, []int{6000}, []netip.Prefix{
		netip.MustParsePrefix("1.1.1.1/32"),
	})

	// removing allowlist rule for the subnet will cause iptables failure
	iptablesExecutor.AddErrCommand("-D INPUT -s 1.1.1.1/32 -i en0 -m comment --comment nordvpn -j ACCEPT")

	networker := GetTestCombined(t)
	networker.firewallManager = firewall

	assert.ErrorIs(t, networker.resetAllowlist(), testfirewall.ErrIptablesFailure)
}

func TestCombined_BlockTraffic(t *testing.T) {
	category.Set(t, category.Unit)

	firewallManager := newMockFirewallManager(t, workingDeviceList, nil)

	networker := GetTestCombined(t)
	networker.firewallManager = firewallManager

	assert.Nil(t, networker.blockTraffic())
}

func TestCombined_BlockTraffic_FirewallFailure(t *testing.T) {
	category.Set(t, category.Unit)

	failingIptables := testfirewall.NewIptablesMock(false)
	// iptables will fail when inserting the rule
	failingIptables.AddErrCommand("-I INPUT -i en0 -m comment --comment nordvpn -j DROP")

	firewallManager := newMockFirewallManager(t, workingDeviceList, &failingIptables)

	networker := GetTestCombined(t)
	networker.firewallManager = firewallManager

	assert.ErrorIs(t, networker.blockTraffic(), testfirewall.ErrIptablesFailure)
}

func TestCombined_BlockTraffic_DevicesListingFailure(t *testing.T) {
	category.Set(t, category.Unit)

	firewallManager := newMockFirewallManager(t, failingDeviceList, nil)

	networker := GetTestCombined(t)
	networker.firewallManager = firewallManager

	assert.ErrorIs(t, networker.blockTraffic(), mock.ErrOnPurpose)
}

func TestCombined_UnblockTraffic(t *testing.T) {
	category.Set(t, category.Unit)

	networker := GetTestCombined(t)
	networker.firewallManager = newMockFirewallManager(t, nil, nil)
	networker.blockTraffic()

	assert.Nil(t, networker.unblockTraffic())
}

func TestCombined_UnblockTraffic_FirewallFailure(t *testing.T) {
	category.Set(t, category.Unit)

	failingIptables := testfirewall.NewIptablesMock(false)
	// iptables will fail when deleting the rule
	failingIptables.AddErrCommand("-D INPUT -i en0 -m comment --comment nordvpn -j DROP")

	networker := GetTestCombined(t)
	networker.firewallManager = newMockFirewallManager(t, workingDeviceList, &failingIptables)

	assert.Nil(t, networker.blockTraffic())
	assert.ErrorIs(t, networker.unblockTraffic(), testfirewall.ErrIptablesFailure)
}

func TestCombined_SetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		firewallManager firewall.FirewallManager
		router          routes.Service
		allowlist       config.Allowlist
		err             error
	}{
		{
			name:            "device listing failure",
			firewallManager: newMockFirewallManager(t, failingDeviceList, nil),
			router:          workingRouter{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: mock.ErrOnPurpose,
		},
		{
			name:            "router failure",
			firewallManager: newMockFirewallManager(t, nil, nil),
			router:          failingRouter{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: mock.ErrOnPurpose,
		},
		{
			name:            "firewall failure",
			firewallManager: newMockFailingFirewallManager(t, workingDeviceList),
			router:          workingRouter{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: testfirewall.ErrIptablesFailure,
		},
		{
			name:            "invalid allowlist",
			firewallManager: newMockFirewallManager(t, nil, nil),
			router:          workingRouter{},
			allowlist:       config.NewAllowlist(nil, nil, nil),
		},
		{
			name:            "success",
			firewallManager: newMockFirewallManager(t, nil, nil),
			router:          workingRouter{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			networker := GetTestCombined(t)
			networker.allowlistRouter = test.router
			networker.firewallManager = test.firewallManager

			assert.ErrorIs(t, networker.setAllowlist(test.allowlist), test.err)
		})
	}
}

func TestCombined_UnsetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		allowlist allowlist.Routing
		rt        routes.Service
		err       error
	}{
		{
			name: "router failure",
			rt:   failingRouter{},
			err:  mock.ErrOnPurpose,
		},
		{
			name: "success",
			rt:   workingRouter{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			networker := GetTestCombined(t)
			networker.allowlistRouter = test.rt

			err := networker.unsetAllowlist()
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestCombined_UnsetAllowlist_FirewallFailure(t *testing.T) {
	category.Set(t, category.Unit)

	iptablesMock := testfirewall.NewIptablesMock(false)
	// firewall will fail when removing the rule for the subnet
	iptablesMock.AddErrCommand("-D INPUT -s 1.1.1.1/32 -i en0 -m comment --comment nordvpn -j ACCEPT")

	firewallManager := newMockFirewallManager(t, workingDeviceList, &iptablesMock)

	err := firewallManager.SetAllowlist([]int{}, []int{}, []netip.Prefix{netip.MustParsePrefix("1.1.1.1/32")})
	assert.Nil(t, err)

	networker := GetTestCombined(t)
	networker.firewallManager = firewallManager

	assert.ErrorIs(t, networker.unsetAllowlist(), testfirewall.ErrIptablesFailure)
}

func TestCombined_SetNetwork(t *testing.T) {
	category.Set(t, category.Unit)

	UDPPorts := []int64{550, 200, 100}
	TCPPorts := []int64{220, 35}

	tests := []struct {
		name      string
		fw        firewall.FirewallManager
		allowlist allowlist.Routing
		rt        routes.Service
		routing   routes.PolicyService
		err       error
	}{
		{
			name:      "firewall failure",
			fw:        newMockFailingFirewallManager(t, workingDeviceList),
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			routing:   &workingRoutingSetup{},
			err:       testfirewall.ErrIptablesFailure,
		},
		{
			name:      "router failure",
			fw:        newMockFirewallManager(t, workingDeviceList, nil),
			allowlist: workingAllowlistRouting{},
			rt:        failingRouter{},
			routing:   &workingRoutingSetup{},
			err:       mock.ErrOnPurpose,
		},
		{
			name:      "device listing failure",
			fw:        newMockFirewallManager(t, failingDeviceList, nil),
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			routing:   &workingRoutingSetup{},
			err:       mock.ErrOnPurpose,
		},
		{
			name:      "success",
			fw:        newMockFirewallManager(t, workingDeviceList, nil),
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			routing:   &workingRoutingSetup{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				nil,
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				&workingIpv6{},
				test.fw,
				test.allowlist,
				test.routing,
				nil,
				nil,
				nil,
				&workingExitNode{},
				0,
				false,
			)
			assert.False(t, netw.IsNetworkSet())
			err := netw.setNetwork(
				config.NewAllowlist(
					UDPPorts,
					TCPPorts, []string{"192.168.0.1/24", "1.1.1.1/32"},
				),
			)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestCombined_UnsetNetwork(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		fw        firewall.FirewallManager
		allowlist allowlist.Routing
		rt        routes.Service
		err       error
	}{
		{
			name:      "router failure",
			fw:        newMockFirewallManager(t, nil, nil),
			allowlist: workingAllowlistRouting{},
			rt:        failingRouter{},
			err:       mock.ErrOnPurpose,
		},
		{
			name:      "success",
			fw:        newMockFirewallManager(t, nil, nil),
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				nil,
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				&workingIpv6{},
				test.fw,
				test.allowlist,
				&workingRoutingSetup{},
				nil,
				nil,
				nil,
				&workingExitNode{},
				0,
				false,
			)
			assert.ErrorIs(t, netw.unsetNetwork(), test.err)
		})
	}
}

func TestCombined_UnsetNetwork_FirewallFailure(t *testing.T) {
	iptablesMock := testfirewall.NewIptablesMock(false)
	iptablesMock.AddErrCommand("-D INPUT -i en0 -m connmark --mark 0 -m comment --comment nordvpn -j ACCEPT")

	firewallManager := newMockFirewallManager(t, workingDeviceList, &iptablesMock)

	assert.Nil(t, firewallManager.ApiAllowlist())

	networker := GetTestCombined(t)
	networker.firewallManager = firewallManager

	assert.ErrorIs(t, networker.unsetNetwork(), testfirewall.ErrIptablesFailure)
}

func TestCombined_AllowIncoming(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		allowlist  allowlist.Routing
		rt         routes.Service
		publicKey  string
		address    string
		ruleName   string
		lanAllowed bool
		err        error
	}{
		{
			name:       "a1",
			allowlist:  workingAllowlistRouting{},
			rt:         workingRouter{},
			publicKey:  "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:    "100.100.10.1",
			ruleName:   "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			allowlist:  workingAllowlistRouting{},
			rt:         workingRouter{},
			publicKey:  "a70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:    "100.100.10.1",
			ruleName:   "a70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "a3",
			allowlist:  workingAllowlistRouting{},
			rt:         workingRouter{},
			publicKey:  "a2513324-7bac-4dcc-b059-e12df48d7418",
			address:    "100.100.10.1",
			ruleName:   "a2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "lan not allowed",
			allowlist:  workingAllowlistRouting{},
			rt:         workingRouter{},
			publicKey:  "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:    "100.100.10.1",
			ruleName:   "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			lanAllowed: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				nil,
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				&workingIpv6{},
				newMockFirewallManager(t, workingDeviceList, nil),
				test.allowlist,
				&workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			uniqueAddress := meshnet.UniqueAddress{UID: test.publicKey, Address: netip.MustParseAddr(test.address)}
			err := netw.AllowIncoming(uniqueAddress, test.lanAllowed)
			assert.Equal(t, nil, err)
		})
	}
}

func TestCombined_BlockIncoming(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		allowlist allowlist.Routing
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			name:      "b1",
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:   "100.100.10.1",
			ruleName:  "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
		},
		{
			name:      "b2",
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "b70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:   "100.100.10.1",
			ruleName:  "b70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
		},
		{
			name:      "b3",
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "b2513324-7bac-4dcc-b059-e12df48d7418",
			address:   "100.100.10.1",
			ruleName:  "b2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				nil,
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				&workingIpv6{},
				newMockFirewallManager(t, workingDeviceList, nil),
				test.allowlist,
				&workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			uniqueAddress := meshnet.UniqueAddress{UID: test.publicKey, Address: netip.MustParseAddr(test.address)}
			err := netw.AllowIncoming(uniqueAddress, true)
			assert.Equal(t, nil, err)
			err = netw.BlockIncoming(uniqueAddress)
			assert.Equal(t, nil, err)
		})
	}
}

func TestCombined_SetMesh(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		allowlist allowlist.Routing
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "c2513324-7bac-4dcc-b059-e12df48d7418",
			address:   "100.100.10.1",
			ruleName:  "c2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
		},
	}

	for _, test := range tests {
		t.Run(test.publicKey, func(t *testing.T) {
			netw := NewCombined(
				nil,
				&workingMesh{},
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				&workingIpv6{},
				newMockFirewallManager(t, workingDeviceList, nil),
				test.allowlist,
				&workingRoutingSetup{},
				&workingHostSetter{},
				workingRouter{},
				workingRouter{},
				&workingExitNode{},
				0,
				false,
			)
			assert.ErrorIs(t, test.err, netw.SetMesh(
				mesh.MachineMap{},
				netip.Addr{},
				"",
			))
		})
	}
}

func TestCombined_UnSetMesh(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		allowlist allowlist.Routing
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "d2513324-7bac-4dcc-b059-e12df48d7418",
			address:   "100.100.10.1",
			ruleName:  "d2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
		},
	}

	for _, test := range tests {
		t.Run(test.publicKey, func(t *testing.T) {
			netw := NewCombined(
				nil,
				&workingMesh{},
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				&workingIpv6{},
				newMockFirewallManager(t, workingDeviceList, nil),
				test.allowlist,
				&workingRoutingSetup{},
				&workingHostSetter{},
				workingRouter{},
				workingRouter{},
				&workingExitNode{},
				0,
				false,
			)
			netw.isMeshnetSet = true
			assert.ErrorIs(t, test.err, netw.UnSetMesh())
		})
	}
}

func TestCombined_Reconnect(t *testing.T) {
	category.Set(t, category.Unit)

	// on refresh keep `enableLocalTraffic` value;
	// on UnsetMesh get default value `true`

	router := &workingRoutingSetup{}

	tests := []struct {
		name      string
		enableLan bool
		router    routes.PolicyService
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			name:      "enable local traffic",
			enableLan: true,
		},
		{
			name:      "disable local traffic",
			enableLan: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				&workingMesh{},
				workingGateway{},
				&subs.Subject[string]{},
				nil,
				&workingDNS{},
				&workingIpv6{},
				newMockFirewallManager(t, workingDeviceList, nil),
				nil,
				router,
				&workingHostSetter{},
				workingRouter{},
				workingRouter{},
				&workingExitNode{},
				0,
				false,
			)
			// activate meshnet
			assert.ErrorIs(t, test.err, netw.SetMesh(
				mesh.MachineMap{},
				netip.Addr{},
				"",
			))
			// connect to exit node
			_ = netw.Start(
				vpn.Credentials{},
				vpn.ServerData{},
				config.NewAllowlist(nil, nil, nil),
				[]string{"1.1.1.1"},
				test.enableLan,
			)
			// simulate network change event and refreshVPN
			netw.Reconnect(true)
			assert.Equal(t, test.enableLan, router.EnableLocalTraffic)
		})
	}
}

func TestCombined_allowIncoming(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		address    string
		lanAllowed bool
	}{
		{
			name:       "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:    "100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "a70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:    "100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "a2513324-7bac-4dcc-b059-e12df48d7418",
			address:    "100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "1f391849-f94b-4826-a5ce-acb6e8a4e432",
			address:    "100.100.10.1",
			lanAllowed: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockIptables := testfirewall.IptablesMock{}
			mockFirewall := newMockFirewallManager(t, workingDeviceList, &mockIptables)

			netw := NewCombined(
				nil,
				nil,
				nil,
				&subs.Subject[string]{},
				nil,
				nil,
				nil,
				mockFirewall,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			peerUniqueAddress := meshnet.UniqueAddress{UID: test.name, Address: netip.MustParseAddr(test.address)}
			err := netw.AllowIncoming(peerUniqueAddress, test.lanAllowed)

			assert.Nil(t, err)
		})
	}
}

func TestCombined_Block(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		ruleName string
		address  string
	}{
		{
			name:     "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:  "100.100.10.1",
			ruleName: "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
		},
		{
			name:     "b70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:  "100.100.10.1",
			ruleName: "b70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
		},
		{
			name:     "b2513324-7bac-4dcc-b059-e12df48d7418",
			address:  "100.100.10.1",
			ruleName: "b2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockIptables := testfirewall.IptablesMock{}
			mockFirewall := newMockFirewallManager(t, workingDeviceList, &mockIptables)

			netw := NewCombined(
				nil,
				nil,
				nil,
				&subs.Subject[string]{},
				nil,
				nil,
				nil,
				mockFirewall,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)

			peerUniqueAddress := meshnet.UniqueAddress{UID: test.name, Address: netip.MustParseAddr(test.address)}
			err := netw.AllowIncoming(peerUniqueAddress, false)
			assert.Nil(t, err)

			err = netw.BlockIncoming(meshnet.UniqueAddress{UID: test.name, Address: netip.MustParseAddr(test.address)})
			assert.Nil(t, err)
		})
	}
}

func TestCombined_BlocNonExistingRuleFail(t *testing.T) {
	category.Set(t, category.Unit)

	address := meshnet.UniqueAddress{
		UID:     "dc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
		Address: netip.MustParseAddr("100.100.10.1")}

	firewallManager := newMockFirewallManager(t, workingDeviceList, nil)

	networker := GetTestCombined(t)
	networker.firewallManager = firewallManager

	err := networker.BlockIncoming(address)

	assert.Error(t, err)
}

func TestCombined_allowExistingRuleFail(t *testing.T) {
	category.Set(t, category.Unit)

	address := meshnet.UniqueAddress{
		UID:     "dc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
		Address: netip.MustParseAddr("100.100.10.1")}

	firewallManager := newMockFirewallManager(t, workingDeviceList, nil)
	firewallManager.AllowIncoming(address, true)

	networker := GetTestCombined(t)
	networker.firewallManager = firewallManager

	err := networker.AllowIncoming(address, true)

	assert.Error(t, err)
}

func TestCombined_Refresh(t *testing.T) {
	hostSetter := newMockHostSetter()
	fw := newMockFirewallManager(t, nil, nil)
	exitNode := newWorkingExitNode()

	netw := NewCombined(
		nil,
		&workingMesh{},
		workingGateway{},
		&subs.Subject[string]{},
		workingRouter{},
		&workingDNS{},
		&workingIpv6{},
		fw,
		nil,
		&workingRoutingSetup{},
		hostSetter,
		workingRouter{},
		workingRouter{},
		exitNode,
		0,
		false,
	)

	machineHostName := "test-fuji.nord"
	machineAddress := netip.MustParseAddr("210.44.137.135")
	machinePublicKey := "pUuNJ1Tt5M8Y6is6ZoaDjuoUT29ht5c0RHqyz2UhmEt="
	peer1HostName := "test-everest.nord"
	peer1Address := netip.MustParseAddr("56.132.8.3")
	peer1PublicKey := "5AHWT3bNYBNqHfMMCxP9n3lMfnL0qIZiNr1xmEymMYf="

	peers := mesh.MachinePeers{
		mesh.MachinePeer{
			Hostname:          peer1HostName,
			PublicKey:         peer1PublicKey,
			Address:           peer1Address,
			DoIAllowInbound:   true,
			DoIAllowFileshare: true,
		},
		mesh.MachinePeer{
			Hostname:        "test-altai.nord",
			PublicKey:       "53sMImgjlgHiuEc51qkzTlzoxneliK3BBmzjUB2K2L9=",
			Address:         netip.Addr{},
			DoIAllowInbound: true,
		},
	}

	machineMap := mesh.MachineMap{
		Machine: mesh.Machine{
			Hostname:  machineHostName,
			PublicKey: machinePublicKey,
			Address:   machineAddress,
		},
		Peers: peers,
	}

	netw.Refresh(machineMap)

	assert.Equal(t, 2, len(hostSetter.hosts), "%d DNS hosts were configured, expected 2.", len(hostSetter.hosts))

	expectedMachineDnsHost := dns.Host{
		IP:         machineAddress,
		FQDN:       machineHostName,
		DomainName: "test-fuji",
	}
	assert.Equal(t, expectedMachineDnsHost, hostSetter.hosts[0],
		"DNS host was not configured properly for %s, \nexpected config: \n%v, \nactual config: \n%v",
		expectedMachineDnsHost, hostSetter.hosts[0],
	)

	expectedPeer1DnsHost := dns.Host{
		IP:         peer1Address,
		FQDN:       peer1HostName,
		DomainName: "test-everest",
	}

	assert.Equal(t, expectedPeer1DnsHost, hostSetter.hosts[1],
		"DNS host was not configured properly for %s, \nexpected config: \n%v, \nactual config: \n%v",
		expectedPeer1DnsHost, hostSetter.hosts[1])

	assert.True(t, exitNode.enabled, "Exit node is not enabled after network refresh.")
	assert.Equal(t, peers, exitNode.peers,
		"Exit node peers are not configured properly after network refresh: \nexpected:\n%v\nactual:\n%v",
		peers, exitNode.peers)
}

func TestDnsAfterVPNRefresh(t *testing.T) {
	dns := &workingDNS{}
	netw := NewCombined(
		&mock.WorkingVPN{},
		nil,
		nil,
		&subs.Subject[string]{},
		workingRouter{},
		dns,
		&workingIpv6{},
		newMockFirewallManager(t, workingDeviceList, nil),
		workingAllowlistRouting{},
		&workingRoutingSetup{},
		nil,
		workingRouter{},
		nil,
		&workingExitNode{},
		0,
		false,
	)

	err := netw.start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, config.DNS{"1.1.1.1"})
	assert.NoError(t, err)
	assert.Equal(t, "1.1.1.1", dns.setDNS[0])

	err = netw.SetDNS([]string{"2.2.2.2"})
	assert.NoError(t, err)
	assert.Equal(t, "2.2.2.2", dns.setDNS[0])

	err = netw.refreshVPN()
	assert.NoError(t, err)
	assert.Equal(t, "2.2.2.2", dns.setDNS[0])
}

func TestExitNodeLanAvailability(t *testing.T) {
	tests := []struct {
		name         string
		actions      func(*Combined)
		lanAvailable bool
	}{
		{
			name:         "no actions",
			actions:      func(c *Combined) { _ = c.ResetRouting(mesh.MachinePeer{}, nil) },
			lanAvailable: true,
		},
		{
			name:         "killswitch enabled",
			actions:      func(c *Combined) { _ = c.SetKillSwitch(config.Allowlist{}) },
			lanAvailable: false,
		},
		{
			name:         "VPN enabled",
			actions:      func(c *Combined) { _ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true) },
			lanAvailable: false,
		},
		{
			name:         "LAN discovery enabled",
			actions:      func(c *Combined) { c.SetLanDiscovery(true) },
			lanAvailable: true,
		},
		{
			name:         "LAN discovery disabled",
			actions:      func(c *Combined) { c.SetLanDiscovery(false) },
			lanAvailable: true,
		},
		{
			name:         "killswitch then lan discovery",
			actions:      func(c *Combined) { _ = c.SetKillSwitch(config.Allowlist{}); c.SetLanDiscovery(true) },
			lanAvailable: true,
		},
		{
			name:         "lan discovery then killswitch",
			actions:      func(c *Combined) { c.SetLanDiscovery(true); _ = c.SetKillSwitch(config.Allowlist{}) },
			lanAvailable: true,
		},
		{
			name: "vpn then lan discovery",
			actions: func(c *Combined) {
				_ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true)
				c.SetLanDiscovery(true)
			},
			lanAvailable: true,
		},
		{
			name: "lan discovery then vpn",
			actions: func(c *Combined) {
				c.SetLanDiscovery(true)
				_ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true)
			},
			lanAvailable: true,
		},
		{
			name: "lan discovery then killswitch then lan discovery off",
			actions: func(c *Combined) {
				c.SetLanDiscovery(true)
				_ = c.SetKillSwitch(config.Allowlist{})
				c.SetLanDiscovery(false)
			},
			lanAvailable: false,
		},
		{
			name: "lan discovery then vpn then lan discovery off",
			actions: func(c *Combined) {
				c.SetLanDiscovery(true)
				_ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true)
				c.SetLanDiscovery(false)
			},
			lanAvailable: false,
		},
		{
			name: "vpn then killswitch",
			actions: func(c *Combined) {
				_ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true)
				_ = c.SetKillSwitch(config.Allowlist{})
			},
			lanAvailable: false,
		},
		{
			name: "vpn then killswitch then lan discovery",
			actions: func(c *Combined) {
				_ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true)
				_ = c.SetKillSwitch(config.Allowlist{})
				c.SetLanDiscovery(true)
			},
			lanAvailable: true,
		},
		{
			name: "vpn then killswitch then lan discovery then killswitch off",
			actions: func(c *Combined) {
				_ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true)
				_ = c.SetKillSwitch(config.Allowlist{})
				c.SetLanDiscovery(true)
				_ = c.UnsetKillSwitch()
			},
			lanAvailable: true,
		},
		{
			name: "vpn then killswitch then lan discovery then killswitch off then lan discovery off",
			actions: func(c *Combined) {
				_ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true)
				_ = c.SetKillSwitch(config.Allowlist{})
				c.SetLanDiscovery(true)
				_ = c.UnsetKillSwitch()
				c.SetLanDiscovery(false)
			},
			lanAvailable: false,
		},
		{
			name: "vpn then killswitch then lan discovery then killswitch off then lan discovery off then vpn off",
			actions: func(c *Combined) {
				_ = c.Start(vpn.Credentials{}, vpn.ServerData{}, config.Allowlist{}, nil, true)
				_ = c.SetKillSwitch(config.Allowlist{})
				c.SetLanDiscovery(true)
				_ = c.UnsetKillSwitch()
				c.SetLanDiscovery(false)
				_ = c.Stop()
			},
			lanAvailable: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			combined := GetTestCombined(t)
			test.actions(combined)
			assert.Equal(t, test.lanAvailable, combined.exitNode.(*workingExitNode).LanAvailable)
		})
	}
}
