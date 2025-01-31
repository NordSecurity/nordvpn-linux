package networker

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/device"
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

func GetTestCombined() *Combined {
	return NewCombined(
		&mock.WorkingVPN{},
		&workingMesh{},
		workingGateway{},
		&subs.Subject[string]{},
		workingRouter{},
		&workingDNS{},
		&workingIpv6{},
		newWorkingFirewall(),
		workingAllowlistRouting{},
		workingDeviceList,
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

func (w workingGateway) Retrieve(netip.Prefix, uint) (netip.Addr, net.Interface, error) {
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

type workingFirewall struct {
	rules map[string]firewall.Rule
}

func newWorkingFirewall() *workingFirewall {
	return &workingFirewall{
		rules: make(map[string]firewall.Rule),
	}
}

func (f *workingFirewall) Add(rules []firewall.Rule) error {
	if f.rules == nil {
		return nil
	}

	for _, rule := range rules {
		f.rules[rule.Name] = rule
	}

	return nil
}

func (f *workingFirewall) Delete(rules []string) error {
	if f.rules == nil {
		return nil
	}

	for _, ruleName := range rules {
		delete(f.rules, ruleName)
	}

	return nil
}

func (workingFirewall) Enable() error   { return nil }
func (workingFirewall) Disable() error  { return nil }
func (workingFirewall) IsEnabled() bool { return true }
func (workingFirewall) Flush() error    { return nil }

type workingAllowlistRouting struct{}

func (workingAllowlistRouting) EnablePorts([]int, string, string) error    { return nil }
func (workingAllowlistRouting) EnableSubnets([]netip.Prefix, string) error { return nil }
func (workingAllowlistRouting) Disable() error                             { return nil }

type failingFirewall struct{}

func (failingFirewall) Add([]firewall.Rule) error { return mock.ErrOnPurpose }
func (failingFirewall) Delete([]string) error     { return mock.ErrOnPurpose }
func (failingFirewall) Enable() error             { return mock.ErrOnPurpose }
func (failingFirewall) Disable() error            { return mock.ErrOnPurpose }
func (failingFirewall) IsEnabled() bool           { return false }
func (failingFirewall) Flush() error              { return mock.ErrOnPurpose }

type meshnetterFirewall struct{}

// Check if fw rule generated correctly
func (meshnetterFirewall) Add(rules []firewall.Rule) error {
	for _, rule := range rules {
		if rule.Direction != firewall.Inbound {
			return fmt.Errorf("Rule direction is not inbound")
		}
		if rule.Allow != true {
			return fmt.Errorf("Rule blocks packets")
		}
	}
	return nil
}
func (meshnetterFirewall) Delete([]string) error { return nil }
func (meshnetterFirewall) Enable() error         { return nil }
func (meshnetterFirewall) Disable() error        { return nil }
func (meshnetterFirewall) IsEnabled() bool       { return true }
func (meshnetterFirewall) Flush() error          { return nil }

func workingDeviceList() ([]net.Interface, error) {
	return []net.Interface{mock.En0Interface}, nil
}

func failingDeviceList() ([]net.Interface, error) { return nil, mock.ErrOnPurpose }

type workingRoutingSetup struct {
	EnableLocalTraffic bool
}

func (r *workingRoutingSetup) SetupRoutingRules(_ bool, enableLan bool, _ bool, _ []string) error {
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

func (e *workingExitNode) ResetPeers(peers mesh.MachinePeers,
	lan bool,
	killswitch bool,
	enableAllowlist bool,
	allowlistedSubnets config.Allowlist) error {
	e.peers = peers
	e.LanAvailable = lan
	return nil
}

func (*workingExitNode) DisablePeer(netip.Addr) error { return nil }
func (*workingExitNode) Disable() error               { return nil }

func (e *workingExitNode) ResetFirewall(lan bool,
	killswitch bool,
	enableAllowlist bool,
	allowlist config.Allowlist) error {
	e.LanAvailable = lan
	return nil
}

type workingMesh struct {
	enableErr         error
	networkChangedErr error
}

func (w *workingMesh) Enable(netip.Addr, string) error { return w.enableErr }
func (*workingMesh) Disable() error                    { return nil }
func (*workingMesh) IsActive() bool                    { return false }
func (*workingMesh) Refresh(mesh.MachineMap) error     { return nil }
func (*workingMesh) Tun() tunnel.T                     { return mock.WorkingT{} }
func (*workingMesh) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}
func (w *workingMesh) NetworkChanged() error { return w.networkChangedErr }

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

func TestCombined_Start(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name            string
		gateway         routes.GatewayRetriever
		allowlistRouter routes.Service
		dns             dns.Setter
		vpn             vpn.VPN
		fw              firewall.Service
		allowlist       allowlist.Routing
		devices         device.ListFunc
		routing         routes.PolicyService
		err             error
	}{
		{
			name:            "nil vpn",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             nil,
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         &workingRoutingSetup{},
			err:             errNilVPN,
		},
		{
			name:            "vpn start failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             mock.FailingVPN{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         &workingRoutingSetup{},
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "firewall failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             mock.WorkingInactiveVPN{},
			fw:              failingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         &workingRoutingSetup{},
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "dns failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             failingDNS{},
			vpn:             mock.WorkingInactiveVPN{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         &workingRoutingSetup{},
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "device listing failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             mock.WorkingInactiveVPN{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         failingDeviceList,
			routing:         &workingRoutingSetup{},
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "successful start",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             &mock.WorkingVPN{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         &workingRoutingSetup{},
			err:             nil,
		},
		{
			name:            "restart",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             &mock.ActiveVPN{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
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
				test.fw,
				test.allowlist,
				test.devices,
				test.routing,
				nil,
				workingRouter{},
				nil,
				&workingExitNode{},
				0,
				false,
			)
			err := netw.Start(
				context.Background(),
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
				&workingFirewall{},
				workingAllowlistRouting{},
				nil,
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
			vpn:      mock.ActiveVPN{},
			expected: tunnel.Statistics{Tx: 1337, Rx: 1337},
		},
		{
			name: "inactive vpn",
			vpn:  mock.WorkingInactiveVPN{},
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
			netw := NewCombined(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, false)
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
				&workingFirewall{},
				nil,
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
				&workingFirewall{},
				nil,
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

	tests := []struct {
		name      string
		fw        firewall.Service
		allowlist allowlist.Routing
		devices   device.ListFunc
		routing   routes.PolicyService
		err       error
	}{
		{
			name:      "firewall failure",
			fw:        failingFirewall{},
			allowlist: workingAllowlistRouting{},
			devices:   workingDeviceList,
			routing:   &workingRoutingSetup{},
			err:       mock.ErrOnPurpose,
		},
		{
			name:      "device listing failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			devices:   failingDeviceList,
			err:       mock.ErrOnPurpose,
			routing:   &workingRoutingSetup{},
		},
		{
			name:      "success",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			devices:   workingDeviceList,
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
				workingRouter{},
				&workingDNS{},
				workingIpv6{},
				test.fw,
				test.allowlist,
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			assert.ErrorIs(t, netw.resetAllowlist(), test.err)
		})
	}
}

func TestCombined_BlockTraffic(t *testing.T) {
	category.Set(t, category.Route)

	tests := []struct {
		name    string
		fw      firewall.Service
		devices device.ListFunc
		routing routes.PolicyService
		err     error
	}{
		{
			name:    "firewall failure",
			fw:      failingFirewall{},
			devices: workingDeviceList,
			err:     mock.ErrOnPurpose,
			routing: &workingRoutingSetup{},
		},
		{
			name:    "device listing failure",
			fw:      &workingFirewall{},
			devices: failingDeviceList,
			err:     mock.ErrOnPurpose,
			routing: &workingRoutingSetup{},
		},
		{
			name:    "success",
			fw:      &workingFirewall{},
			devices: workingDeviceList,
			routing: &workingRoutingSetup{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// It's fine to pass nils to values provided via constructor
			// which are not used in the test.
			netw := NewCombined(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				test.fw,
				nil,
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			assert.ErrorIs(t, netw.blockTraffic(), test.err)
		})
	}
}

func TestCombined_UnblockTraffic(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		fw   firewall.Service
		err  error
	}{
		{
			name: "firewall failure",
			fw:   failingFirewall{},
			err:  mock.ErrOnPurpose,
		},
		{
			name: "success",
			fw:   &workingFirewall{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// It's fine to pass nils to values provided via constructor
			// which are not used in the test.
			netw := NewCombined(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				test.fw,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			assert.ErrorIs(t, netw.unblockTraffic(), test.err)
		})
	}
}

func TestCombined_AllowIPv6Traffic(t *testing.T) {
	category.Set(t, category.Route)

	tests := []struct {
		name    string
		fw      firewall.Service
		devices device.ListFunc
		routing routes.PolicyService
		err     error
	}{
		{
			name:    "firewall failure",
			fw:      failingFirewall{},
			devices: workingDeviceList,
			err:     mock.ErrOnPurpose,
			routing: &workingRoutingSetup{},
		},
		{
			name:    "device listing failure",
			fw:      &workingFirewall{},
			devices: failingDeviceList,
			err:     mock.ErrOnPurpose,
			routing: &workingRoutingSetup{},
		},
		{
			name:    "success",
			fw:      &workingFirewall{},
			devices: workingDeviceList,
			routing: &workingRoutingSetup{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// It's fine to pass nils to values provided via constructor
			// which are not used in the test.
			netw := NewCombined(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				test.fw,
				nil,
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			assert.ErrorIs(t, netw.allowIPv6Traffic(), test.err)
		})
	}
}

func TestCombined_StopAllowedIPv6Traffic(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		fw   firewall.Service
		err  error
	}{
		{
			name: "firewall failure",
			fw:   failingFirewall{},
			err:  mock.ErrOnPurpose,
		},
		{
			name: "success",
			fw:   &workingFirewall{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// It's fine to pass nils to values provided via constructor
			// which are not used in the test.
			netw := NewCombined(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				test.fw,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			assert.ErrorIs(t, netw.stopAllowedIPv6Traffic(), test.err)
		})
	}
}

func TestCombined_SetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name             string
		devices          device.ListFunc
		routing          routes.PolicyService
		rt               routes.Service
		fw               firewall.Service
		allowlistRouting allowlist.Routing
		allowlist        config.Allowlist
		err              error
	}{
		{
			name:             "device listing failure",
			devices:          failingDeviceList,
			routing:          &workingRoutingSetup{},
			rt:               workingRouter{},
			fw:               &workingFirewall{},
			allowlistRouting: workingAllowlistRouting{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: mock.ErrOnPurpose,
		},
		{
			name:             "router failure",
			devices:          workingDeviceList,
			routing:          &workingRoutingSetup{},
			rt:               failingRouter{},
			fw:               &workingFirewall{},
			allowlistRouting: workingAllowlistRouting{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: nil, // not connected - router is not invoked, no error
		},
		{
			name:             "firewall failure",
			devices:          workingDeviceList,
			routing:          &workingRoutingSetup{},
			rt:               workingRouter{},
			fw:               failingFirewall{},
			allowlistRouting: workingAllowlistRouting{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: mock.ErrOnPurpose,
		},
		{
			name:             "invalid allowlist",
			devices:          workingDeviceList,
			routing:          &workingRoutingSetup{},
			rt:               workingRouter{},
			fw:               &workingFirewall{},
			allowlistRouting: &workingAllowlistRouting{},
			allowlist:        config.NewAllowlist(nil, nil, nil),
		},
		{
			name:             "success",
			devices:          workingDeviceList,
			routing:          &workingRoutingSetup{},
			rt:               workingRouter{},
			fw:               &workingFirewall{},
			allowlistRouting: workingAllowlistRouting{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
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
				test.allowlistRouting,
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			assert.ErrorIs(t, netw.setAllowlist(test.allowlist), test.err)
		})
	}
}

func TestCombined_UnsetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		fw        firewall.Service
		allowlist allowlist.Routing
		rt        routes.Service
		err       error
	}{
		{
			name:      "firewall failure",
			fw:        failingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			err:       mock.ErrOnPurpose,
		},
		{
			name:      "router failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        failingRouter{},
		},
		{
			name:      "success",
			fw:        &workingFirewall{},
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
				workingDeviceList,
				&workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			err := netw.unsetAllowlist()
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestCombined_SetNetwork(t *testing.T) {
	category.Set(t, category.Unit)

	UDPPorts := []int64{550, 200, 100}
	TCPPorts := []int64{220, 35}

	tests := []struct {
		name      string
		fw        firewall.Service
		allowlist allowlist.Routing
		rt        routes.Service
		devices   device.ListFunc
		routing   routes.PolicyService
		err       error
	}{
		{
			name:      "firewall failure",
			fw:        failingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			devices:   workingDeviceList,
			routing:   &workingRoutingSetup{},
			err:       mock.ErrOnPurpose,
		},
		{
			name:      "router failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        failingRouter{},
			devices:   workingDeviceList,
			routing:   &workingRoutingSetup{},
			err:       nil, // not connected - router is not invoked, no error
		},
		{
			name:      "device listing failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			devices:   failingDeviceList,
			routing:   &workingRoutingSetup{},
			err:       mock.ErrOnPurpose,
		},
		{
			name:      "success",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			routing:   &workingRoutingSetup{},
			devices:   workingDeviceList,
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
				test.devices,
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
		fw        firewall.Service
		allowlist allowlist.Routing
		rt        routes.Service
		err       error
	}{
		{
			name:      "firewall failure",
			fw:        failingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			err:       mock.ErrOnPurpose,
		},
		{
			name:      "router failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        failingRouter{},
		},
		{
			name:      "success",
			fw:        &workingFirewall{},
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
				workingDeviceList,
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

func TestCombined_AllowIncoming(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		fw         firewall.Service
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
			fw:         &workingFirewall{},
			allowlist:  workingAllowlistRouting{},
			rt:         workingRouter{},
			publicKey:  "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:    "100.100.10.1",
			ruleName:   "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "a2",
			fw:         &workingFirewall{},
			allowlist:  workingAllowlistRouting{},
			rt:         workingRouter{},
			publicKey:  "a70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:    "100.100.10.1",
			ruleName:   "a70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "a3",
			fw:         &workingFirewall{},
			allowlist:  workingAllowlistRouting{},
			rt:         workingRouter{},
			publicKey:  "a2513324-7bac-4dcc-b059-e12df48d7418",
			address:    "100.100.10.1",
			ruleName:   "a2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "lan not allowed",
			fw:         &workingFirewall{},
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
				test.fw,
				test.allowlist,
				workingDeviceList,
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
		fw        firewall.Service
		allowlist allowlist.Routing
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			name:      "b1",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:   "100.100.10.1",
			ruleName:  "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
		},
		{
			name:      "b2",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "b70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:   "100.100.10.1",
			ruleName:  "b70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
		},
		{
			name:      "b3",
			fw:        &workingFirewall{},
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
				test.fw,
				test.allowlist,
				workingDeviceList,
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
		fw        firewall.Service
		allowlist allowlist.Routing
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			fw:        &workingFirewall{},
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
				test.fw,
				test.allowlist,
				workingDeviceList,
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
		fw        firewall.Service
		allowlist allowlist.Routing
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			fw:        &workingFirewall{},
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
				test.fw,
				test.allowlist,
				workingDeviceList,
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
	fw := &workingFirewall{}

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
			meshnet := &workingMesh{}
			meshnet.networkChangedErr = mock.ErrOnPurpose
			netw := NewCombined(
				nil,
				meshnet,
				workingGateway{},
				&subs.Subject[string]{},
				nil,
				&workingDNS{},
				&workingIpv6{},
				fw,
				nil,
				workingDeviceList,
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
				context.Background(),
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
	tests := []struct {
		name               string
		ruleName           string
		lanAllowedRuleName string
		address            string
		lanAllowed         bool
	}{
		{
			name:       "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:    "100.100.10.1",
			ruleName:   "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "a70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:    "100.100.10.1",
			ruleName:   "a70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			name:       "a2513324-7bac-4dcc-b059-e12df48d7418",
			address:    "100.100.10.1",
			ruleName:   "a2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
			lanAllowed: true,
		},
		{
			name:               "1f391849-f94b-4826-a5ce-acb6e8a4e432",
			address:            "100.100.10.1",
			ruleName:           "1f391849-f94b-4826-a5ce-acb6e8a4e432-allow-rule-100.100.10.1",
			lanAllowedRuleName: "1f391849-f94b-4826-a5ce-acb6e8a4e432-block-lan-rule-100.100.10.1",
			lanAllowed:         false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockFirewall := testfirewall.NewMockFirewall()
			netw := NewCombined(
				nil,
				nil,
				nil,
				&subs.Subject[string]{},
				nil,
				nil,
				nil,
				&mockFirewall,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			err := netw.allowIncoming(test.name, netip.MustParseAddr(test.address), test.lanAllowed)

			assert.Nil(t, err)
			if !test.lanAllowed {
				assert.Equal(t, test.lanAllowedRuleName, netw.rules[0])
				assert.Equal(t, test.ruleName, netw.rules[1])

				assert.Equal(t, test.lanAllowedRuleName, mockFirewall.Rules[1].Name)
				assert.Equal(t, firewall.Inbound, mockFirewall.Rules[1].Direction)

				assert.Equal(t, test.ruleName, mockFirewall.Rules[0].Name)
				assert.Equal(t, firewall.Inbound, mockFirewall.Rules[0].Direction)
			} else {
				assert.Equal(t, test.ruleName, netw.rules[0])

				assert.Equal(t, test.ruleName, mockFirewall.Rules[0].Name)
				assert.Equal(t, firewall.Inbound, mockFirewall.Rules[0].Direction)
			}
		})
	}
}

func TestCombined_Block(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		address  string
		fw       firewall.Service
	}{
		{
			name:     "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:  "100.100.10.1",
			ruleName: "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
		{
			name:     "b70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:  "100.100.10.1",
			ruleName: "b70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
		{
			name:     "b2513324-7bac-4dcc-b059-e12df48d7418",
			address:  "100.100.10.1",
			ruleName: "b2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				nil,
				nil,
				&subs.Subject[string]{},
				nil,
				nil,
				nil,
				test.fw,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			err := netw.allowIncoming(test.name, netip.MustParseAddr(test.address), true)
			assert.Nil(t, err)
			assert.Equal(t, netw.rules[0], test.ruleName)

			err = netw.BlockIncoming(meshnet.UniqueAddress{UID: test.name, Address: netip.MustParseAddr(test.address)})
			assert.Nil(t, err)
			assert.Equal(t, 0, len(netw.rules))
		})
	}
}

func TestCombined_allowGeneratedRule(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		address  string
		fw       firewall.Service
	}{
		{
			name:     "cc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:  "100.100.10.1",
			ruleName: "cc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
		{
			name:     "c70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:  "100.100.10.1",
			ruleName: "c70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
		{
			name:     "c2513324-7bac-4dcc-b059-e12df48d7418",
			address:  "100.100.10.1",
			ruleName: "c2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				nil,
				nil,
				&subs.Subject[string]{},
				nil,
				nil,
				nil,
				test.fw,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			err := netw.allowIncoming(test.name, netip.MustParseAddr(test.address), true)
			assert.Equal(t, nil, err)
			assert.Equal(t, netw.rules[0], test.ruleName)
		})
	}
}

func TestCombined_BlocNonExistingRuleFail(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		address  string
		fw       firewall.Service
	}{
		{
			name:     "dc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:  "100.100.10.1",
			ruleName: "dc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				nil,
				nil,
				&subs.Subject[string]{},
				nil,
				nil,
				nil,
				test.fw,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			// Should fail to block rule non existing
			expectedErrorMsg := fmt.Sprintf("allow rule does not exist for %s", test.ruleName)
			err := netw.BlockIncoming(meshnet.UniqueAddress{UID: test.name, Address: netip.MustParseAddr(test.address)})
			assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		})
	}
}

func TestCombined_allowExistingRuleFail(t *testing.T) {
	tests := []struct {
		name          string
		allowRuleName string
		expectedRules []string
		address       string
	}{
		{
			name:          "ec30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:       "100.100.10.1",
			allowRuleName: "ec30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			expectedRules: []string{"ec30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-block-lan-rule-100.100.10.1", "ec30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockFirewall := testfirewall.NewMockFirewall()

			netw := NewCombined(
				nil,
				nil,
				nil,
				&subs.Subject[string]{},
				nil,
				nil,
				nil,
				&mockFirewall,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			err := netw.allowIncoming(test.name, netip.MustParseAddr(test.address), false)
			assert.Equal(t, nil, err)
			assert.Equal(t, netw.rules, test.expectedRules)
			// Should fail to add rule second time
			expectedErrorMsg := fmt.Sprintf("allow rule already exist for %s", test.allowRuleName)
			err = netw.allowIncoming(test.name, netip.MustParseAddr(test.address), false)
			assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		})
	}
}

func TestCombined_Refresh(t *testing.T) {
	hostSetter := newMockHostSetter()
	fw := newWorkingFirewall()
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
		workingDeviceList,
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
		IP:          machineAddress,
		FQDN:        machineHostName,
		DomainNames: []string{"test-fuji"},
	}
	assert.Equal(t, expectedMachineDnsHost, hostSetter.hosts[0],
		"DNS host was not configured properly for %s, \nexpected config: \n%v, \nactual config: \n%v",
		expectedMachineDnsHost, hostSetter.hosts[0],
	)

	expectedPeer1DnsHost := dns.Host{
		IP:          peer1Address,
		FQDN:        peer1HostName,
		DomainNames: []string{"test-everest"},
	}

	assert.Equal(t, expectedPeer1DnsHost, hostSetter.hosts[1],
		"DNS host was not configured properly for %s, \nexpected config: \n%v, \nactual config: \n%v",
		expectedPeer1DnsHost, hostSetter.hosts[1])

	// transform rules to rule names for printing in case of assertion failure
	ruleNames := []string{}
	for _, rule := range fw.rules {
		ruleNames = append(ruleNames, rule.Name)
	}

	// fileshare is forbidden here, so no new fileshare rules were added
	assert.Equal(t, 5, len(fw.rules), "%d firewall rules were configured, expected 5, rules content: \n%s",
		len(fw.rules),
		strings.Join(ruleNames, "\n"))

	// permit fileshare
	netw.isFilesharePermitted = true

	netw.Refresh(machineMap)

	// now after refresh, fileshare rules are also added
	assert.Equal(t, 6, len(fw.rules), "%d firewall rules were configured, expected 6, rules content: \n%s",
		len(fw.rules),
		strings.Join(ruleNames, "\n"))

	defaultMeshBlockRuleName := "default-mesh-block"

	expectedDefaultMeshBlockFwRule := firewall.Rule{
		Name:           defaultMeshBlockRuleName,
		Direction:      firewall.Inbound,
		RemoteNetworks: []netip.Prefix{defaultMeshSubnet},
		Allow:          false,
		Comment:        "nordvpn-meshnet",
	}

	assert.Equal(t, expectedDefaultMeshBlockFwRule, fw.rules[defaultMeshBlockRuleName],
		"default-mesh-block rule is incorrectly configured, \nexpected config: \n%v, \nactual config: \n%v",
		expectedDefaultMeshBlockFwRule, fw.rules[defaultMeshBlockRuleName])

	expectedDefaultMeshAllowEstablishedFwRule := firewall.Rule{
		Name:           "default-mesh-allow-established",
		Direction:      firewall.Inbound,
		RemoteNetworks: []netip.Prefix{defaultMeshSubnet},
		ConnectionStates: firewall.ConnectionStates{
			SrcAddr: machineAddress,
			States: []firewall.ConnectionState{
				firewall.Related,
				firewall.Established,
			},
		},
		Allow:   true,
		Comment: "nordvpn-meshnet",
	}

	assert.Equal(t, expectedDefaultMeshAllowEstablishedFwRule, fw.rules["default-mesh-allow-established"],
		"default-mesh-allow-established is incorrectly configured, \nexpected config: \n%v, \nactual config: \n%v",
		expectedDefaultMeshAllowEstablishedFwRule, fw.rules["default-mesh-allow-established"])

	machineFwAllowRuleName := fmt.Sprintf("%s-allow-rule-%s", machinePublicKey, machineAddress.String())
	expectedAllowMachineFwRule := firewall.Rule{
		Name:           machineFwAllowRuleName,
		Direction:      firewall.Inbound,
		RemoteNetworks: []netip.Prefix{netip.PrefixFrom(machineAddress, machineAddress.BitLen())},
		Allow:          true,
		Comment:        "nordvpn-meshnet",
	}

	assert.Equal(t, expectedAllowMachineFwRule, fw.rules[machineFwAllowRuleName],
		"allow rule for the host machine is incorrectly configured, \nexpected config: \n%v, \nactual config: \n%v",
		expectedAllowMachineFwRule, fw.rules[machineFwAllowRuleName])

	peer1FwAllowRuleName := fmt.Sprintf("%s-allow-rule-%s", peer1PublicKey, peer1Address.String())
	expectedAllowPeer1Rule := firewall.Rule{
		Name:           peer1FwAllowRuleName,
		Direction:      firewall.Inbound,
		RemoteNetworks: []netip.Prefix{netip.PrefixFrom(peer1Address, peer1Address.BitLen())},
		Allow:          true,
		Comment:        "nordvpn-meshnet",
	}

	assert.Equal(t, expectedAllowPeer1Rule, fw.rules[peer1FwAllowRuleName],
		"allow rule for the peer is incorrectly configured, \nexpected config: \n%v, \nactual config: \n%v",
		expectedAllowPeer1Rule, fw.rules[peer1FwAllowRuleName],
	)

	peer1FwAllowFileshareRuleName := fmt.Sprintf("%s-allow-fileshare-rule-%s", peer1PublicKey, peer1Address.String())
	expectedAllowFilesharePeer1Rule := firewall.Rule{
		Name:           peer1FwAllowFileshareRuleName,
		Direction:      firewall.Inbound,
		Ports:          []int{49111},
		Protocols:      []string{"tcp"},
		PortsDirection: firewall.Destination,
		RemoteNetworks: []netip.Prefix{netip.PrefixFrom(peer1Address, peer1Address.BitLen())},
		Allow:          true,
		Comment:        "nordvpn-meshnet",
	}

	assert.Equal(t, expectedAllowFilesharePeer1Rule, fw.rules[peer1FwAllowFileshareRuleName],
		"allow fileshare rule for the peer is incorrectly configured, \nexpected config: \n%v, \nactual config: \n%v",
		expectedAllowFilesharePeer1Rule, fw.rules[peer1FwAllowFileshareRuleName],
	)

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
		newWorkingFirewall(),
		workingAllowlistRouting{},
		workingDeviceList,
		&workingRoutingSetup{},
		nil,
		workingRouter{},
		nil,
		&workingExitNode{},
		0,
		false,
	)

	ctx := context.Background()
	err := netw.start(
		ctx,
		vpn.Credentials{},
		vpn.ServerData{},
		config.Allowlist{},
		config.DNS{"1.1.1.1"},
	)
	assert.NoError(t, err)
	assert.Equal(t, "1.1.1.1", dns.setDNS[0])

	err = netw.SetDNS([]string{"2.2.2.2"})
	assert.NoError(t, err)
	assert.Equal(t, "2.2.2.2", dns.setDNS[0])

	err = netw.refreshVPN(ctx)
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
			name: "VPN enabled",
			actions: func(c *Combined) {
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
			},
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
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
				c.SetLanDiscovery(true)
			},
			lanAvailable: true,
		},
		{
			name: "lan discovery then vpn",
			actions: func(c *Combined) {
				c.SetLanDiscovery(true)
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
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
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
				c.SetLanDiscovery(false)
			},
			lanAvailable: false,
		},
		{
			name: "vpn then killswitch",
			actions: func(c *Combined) {
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
				_ = c.SetKillSwitch(config.Allowlist{})
			},
			lanAvailable: false,
		},
		{
			name: "vpn then killswitch then lan discovery",
			actions: func(c *Combined) {
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
				_ = c.SetKillSwitch(config.Allowlist{})
				c.SetLanDiscovery(true)
			},
			lanAvailable: true,
		},
		{
			name: "vpn then killswitch then lan discovery then killswitch off",
			actions: func(c *Combined) {
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
				_ = c.SetKillSwitch(config.Allowlist{})
				c.SetLanDiscovery(true)
				_ = c.UnsetKillSwitch()
			},
			lanAvailable: true,
		},
		{
			name: "vpn then killswitch then lan discovery then killswitch off then lan discovery off",
			actions: func(c *Combined) {
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
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
				_ = c.Start(
					context.Background(),
					vpn.Credentials{},
					vpn.ServerData{},
					config.Allowlist{},
					nil,
					true,
				)
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
			combined := GetTestCombined()
			test.actions(combined)
			assert.Equal(t, test.lanAvailable, combined.exitNode.(*workingExitNode).LanAvailable)
		})
	}
}

func rulesToString(t *testing.T, rules []firewall.Rule) string {
	t.Helper()

	ruleNames := []string{}
	for _, rule := range rules {
		ruleNames = append(ruleNames, rule.Name)
	}

	return strings.Join(ruleNames, "\n")
}

func TestResetRouting(t *testing.T) {
	peer1Address := netip.MustParseAddr("163.190.101.26")
	peer1PublicKey := "hCRTygV0hU6AtYrHuEvjOXd0UCobDd48hDJFkOMSmC="

	peer2Address := netip.MustParseAddr("190.53.114.47")
	peer2PublicKey := "IiENMnpmmS4VWdXgCDoytzZozV8d4z5bu103nMrJen="

	peer3Address := netip.MustParseAddr("144.79.247.102")
	peer3PublicKey := "2oob1sS0p8v4G6jxOoDZjq5lmaHfAm2d5CJPRMLKxw="

	peer4Address := netip.MustParseAddr("121.92.239.59")
	peer4PublicKey := "Ha3dzBMzdsrw3pEB3UuJE7NxlcCGZYopqrBN8HSqGK="

	peer5Address := netip.MustParseAddr("36.166.227.80")
	peer5PublicKey := "0oaqVJEXsgAZooshXxHClE3nmTB2O6wVRFfEZy5Yjp="

	peers := mesh.MachinePeers{
		{
			Hostname:             "peer-1",
			Address:              peer1Address,
			PublicKey:            peer1PublicKey,
			DoIAllowInbound:      true,
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: false,
		},
		{
			Hostname:             "peer-2",
			Address:              peer2Address,
			PublicKey:            peer2PublicKey,
			DoIAllowInbound:      true,
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: false,
		},
		{
			Hostname:             "peer-3",
			Address:              peer3Address,
			PublicKey:            peer3PublicKey,
			DoIAllowInbound:      true,
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: true,
		},
		{
			Hostname:             "peer-4",
			Address:              peer4Address,
			PublicKey:            peer4PublicKey,
			DoIAllowInbound:      false,
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: true,
		},
		{
			Hostname:             "peer-5",
			Address:              peer5Address,
			PublicKey:            peer5PublicKey,
			DoIAllowInbound:      true,
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: true,
		},
	}

	tests := []struct {
		name           string
		changedPeerIdx int
		expectedRules  []firewall.Rule
	}{
		{
			name:           "no routing/no lan",
			changedPeerIdx: 0,
			expectedRules: []firewall.Rule{
				{
					Name:      peer1PublicKey + allowIncomingRule + peer1Address.String(),
					Direction: firewall.Inbound,
					RemoteNetworks: []netip.Prefix{
						netip.PrefixFrom(peer1Address, peer1Address.BitLen()),
					},
					Allow:   true,
					Comment: "nordvpn-meshnet",
				},
				{
					Name:      peer1PublicKey + blockLanRule + peer1Address.String(),
					Direction: firewall.Inbound,
					LocalNetworks: []netip.Prefix{
						netip.MustParsePrefix("10.0.0.0/8"),
						netip.MustParsePrefix("172.16.0.0/12"),
						netip.MustParsePrefix("192.168.0.0/16"),
						netip.MustParsePrefix("169.254.0.0/16"),
					},
					RemoteNetworks: []netip.Prefix{
						netip.PrefixFrom(peer1Address, peer1Address.BitLen()),
					},
					Allow:   false,
					Comment: "nordvpn-meshnet",
				},
			},
		},
		{
			name:           "no lan",
			changedPeerIdx: 1,
			expectedRules: []firewall.Rule{
				{
					Name:      peer2PublicKey + allowIncomingRule + peer2Address.String(),
					Direction: firewall.Inbound,
					RemoteNetworks: []netip.Prefix{
						netip.PrefixFrom(peer2Address, peer2Address.BitLen()),
					},
					Allow:   true,
					Comment: "nordvpn-meshnet",
				},
				{
					Name:      peer2PublicKey + blockLanRule + peer2Address.String(),
					Direction: firewall.Inbound,
					LocalNetworks: []netip.Prefix{
						netip.MustParsePrefix("10.0.0.0/8"),
						netip.MustParsePrefix("172.16.0.0/12"),
						netip.MustParsePrefix("192.168.0.0/16"),
						netip.MustParsePrefix("169.254.0.0/16"),
					},
					RemoteNetworks: []netip.Prefix{
						netip.PrefixFrom(peer2Address, peer2Address.BitLen()),
					},
					Allow:   false,
					Comment: "nordvpn-meshnet",
				},
			},
		},
		{
			name:           "no routing",
			changedPeerIdx: 2,
			expectedRules: []firewall.Rule{
				{
					Name:      peer3PublicKey + allowIncomingRule + peer3Address.String(),
					Direction: firewall.Inbound,
					RemoteNetworks: []netip.Prefix{
						netip.PrefixFrom(peer3Address, peer3Address.BitLen()),
					},
					Allow:   true,
					Comment: "nordvpn-meshnet",
				},
				{
					Name:      peer3PublicKey + blockLanRule + peer3Address.String(),
					Direction: firewall.Inbound,
					LocalNetworks: []netip.Prefix{
						netip.MustParsePrefix("10.0.0.0/8"),
						netip.MustParsePrefix("172.16.0.0/12"),
						netip.MustParsePrefix("192.168.0.0/16"),
						netip.MustParsePrefix("169.254.0.0/16"),
					},
					RemoteNetworks: []netip.Prefix{
						netip.PrefixFrom(peer3Address, peer3Address.BitLen()),
					},
					Allow:   false,
					Comment: "nordvpn-meshnet",
				},
			},
		},
		{
			name:           "no inbound",
			changedPeerIdx: 3,
			expectedRules:  []firewall.Rule{},
		},
		{
			name:           "no routing",
			changedPeerIdx: 4,
			expectedRules: []firewall.Rule{
				{
					Name:      peer5PublicKey + allowIncomingRule + peer5Address.String(),
					Direction: firewall.Inbound,
					RemoteNetworks: []netip.Prefix{
						netip.PrefixFrom(peer5Address, peer5Address.BitLen()),
					},
					Allow:   true,
					Comment: "nordvpn-meshnet",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockFirewall := testfirewall.NewMockFirewall()
			exitNode := newWorkingExitNode()

			netw := NewCombined(
				nil,
				nil,
				nil,
				&subs.Subject[string]{},
				nil,
				nil,
				nil,
				&mockFirewall,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				exitNode,
				0,
				false,
			)

			err := netw.ResetRouting(peers[test.changedPeerIdx], peers)

			assert.NoError(t, err)

			// transform expected and acutall rules for printing

			assert.Equal(t, test.expectedRules, mockFirewall.Rules, "Invalid rules configured, \nEXPECTED:\n%s\nGOT:\n%s",
				rulesToString(t, test.expectedRules), rulesToString(t, mockFirewall.Rules))
			assert.Equal(t, peers, exitNode.peers)
		})
	}
}
