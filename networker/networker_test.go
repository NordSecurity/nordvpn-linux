package networker

import (
	"fmt"
	"net"
	"net/netip"
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
	testdevice "github.com/NordSecurity/nordvpn-linux/test/device"
	"github.com/NordSecurity/nordvpn-linux/test/errors"
	testtunnel "github.com/NordSecurity/nordvpn-linux/test/tunnel"
	testvpn "github.com/NordSecurity/nordvpn-linux/test/vpn"
	"github.com/NordSecurity/nordvpn-linux/tunnel"

	"github.com/stretchr/testify/assert"
)

type workingGateway struct{}

func (w workingGateway) Default(bool) (netip.Addr, net.Interface, error) {
	return netip.MustParseAddr("1.1.1.1"), testdevice.En0Interface, nil
}

type workingRouter struct{}

func (workingRouter) Add(routes.Route) error { return nil }
func (workingRouter) Flush() error           { return nil }
func (workingRouter) Enable(uint) error      { return nil }
func (workingRouter) Disable() error         { return nil }
func (workingRouter) IsEnabled() bool        { return true }

type failingRouter struct{}

func (failingRouter) Add(routes.Route) error { return errors.ErrOnPurpose }
func (failingRouter) Flush() error           { return errors.ErrOnPurpose }
func (failingRouter) Enable(uint) error      { return errors.ErrOnPurpose }
func (failingRouter) Disable() error         { return errors.ErrOnPurpose }
func (failingRouter) IsEnabled() bool        { return false }

type workingDNS struct{ setDNS []string }

func (w *workingDNS) Set(_ string, dns []string) error { w.setDNS = dns; return nil }
func (w *workingDNS) Unset(string) error               { w.setDNS = nil; return nil }

type failingDNS struct{}

func (failingDNS) Set(string, []string) error { return errors.ErrOnPurpose }
func (failingDNS) Unset(string) error         { return errors.ErrOnPurpose }

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

type workingAllowlistRouting struct{}

func (workingAllowlistRouting) EnablePorts([]int, string, string) error    { return nil }
func (workingAllowlistRouting) EnableSubnets([]netip.Prefix, string) error { return nil }
func (workingAllowlistRouting) Disable() error                             { return nil }

type failingFirewall struct{}

func (failingFirewall) Add([]firewall.Rule) error { return errors.ErrOnPurpose }
func (failingFirewall) Delete([]string) error     { return errors.ErrOnPurpose }
func (failingFirewall) Enable() error             { return errors.ErrOnPurpose }
func (failingFirewall) Disable() error            { return errors.ErrOnPurpose }
func (failingFirewall) IsEnabled() bool           { return false }

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

func workingDeviceList() ([]net.Interface, error) {
	return []net.Interface{testdevice.En0Interface}, nil
}

func failingDeviceList() ([]net.Interface, error) { return nil, errors.ErrOnPurpose }

type workingRoutingSetup struct{}

func (workingRoutingSetup) SetupRoutingRules(net.Interface, bool) error { return nil }
func (workingRoutingSetup) CleanupRouting() error                       { return nil }
func (workingRoutingSetup) TableID() uint                               { return 0 }
func (workingRoutingSetup) Enable() error                               { return nil }
func (workingRoutingSetup) Disable() error                              { return nil }
func (workingRoutingSetup) IsEnabled() bool                             { return true }

type workingExitNode struct {
	enabled bool
	peers   mesh.MachinePeers
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

func (e *workingExitNode) ResetPeers(peers mesh.MachinePeers, lanDiscovery bool) error {
	e.peers = peers
	return nil
}

func (*workingExitNode) DisablePeer(netip.Addr) error { return nil }
func (*workingExitNode) Disable() error               { return nil }

type workingMesh struct{}

func (workingMesh) Enable(netip.Addr, string) error { return nil }
func (workingMesh) Disable() error                  { return nil }
func (workingMesh) IsActive() bool                  { return false }
func (workingMesh) Refresh(mesh.MachineMap) error   { return nil }
func (workingMesh) Tun() tunnel.T                   { return testtunnel.Working{} }
func (workingMesh) StatusMap() (map[string]string, error) {
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
			routing:         workingRoutingSetup{},
			err:             errNilVPN,
		},
		{
			name:            "vpn start failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             testvpn.Failing{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errors.ErrOnPurpose,
		},
		{
			name:            "firewall failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             testvpn.WorkingInactive{},
			fw:              failingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errors.ErrOnPurpose,
		},
		{
			name:            "dns failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             failingDNS{},
			vpn:             testvpn.WorkingInactive{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errors.ErrOnPurpose,
		},
		{
			name:            "device listing failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             testvpn.WorkingInactive{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         failingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errors.ErrOnPurpose,
		},
		{
			name:            "successful start",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             testvpn.Working{},
			fw:              &workingFirewall{},
			allowlist:       &workingAllowlistRouting{},
			devices:         workingDeviceList,
			routing:         workingRoutingSetup{},
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
				nil,
				0,
				false,
			)
			err := netw.Start(
				vpn.Credentials{},
				vpn.ServerData{},
				config.NewAllowlist(nil, nil, nil),
				[]string{"1.1.1.1"},
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
			vpn:  testvpn.Working{},
			dns:  failingDNS{},
			err:  errors.ErrOnPurpose,
		},
		{
			name: "vpn stop failure",
			vpn:  testvpn.Failing{},
			dns:  &workingDNS{},
			err:  errors.ErrOnPurpose,
		},
		{
			name: "successful stop",
			vpn:  testvpn.Working{},
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
				nil,
				nil,
				workingRoutingSetup{},
				nil,
				workingRouter{},
				nil,
				nil,
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
			netw.vpnet = testvpn.Working{}
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
			netw.vpnet = testvpn.Working{}
			err := netw.unsetDNS()
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
			routing:   workingRoutingSetup{},
			err:       errors.ErrOnPurpose,
		},
		{
			name:      "device listing failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			devices:   failingDeviceList,
			err:       errors.ErrOnPurpose,
			routing:   workingRoutingSetup{},
		},
		{
			name:      "success",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			devices:   workingDeviceList,
			routing:   workingRoutingSetup{},
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
			err:     errors.ErrOnPurpose,
			routing: workingRoutingSetup{},
		},
		{
			name:    "device listing failure",
			fw:      &workingFirewall{},
			devices: failingDeviceList,
			err:     errors.ErrOnPurpose,
			routing: workingRoutingSetup{},
		},
		{
			name:    "success",
			fw:      &workingFirewall{},
			devices: workingDeviceList,
			routing: workingRoutingSetup{},
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
			err:  errors.ErrOnPurpose,
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
			err:     errors.ErrOnPurpose,
			routing: workingRoutingSetup{},
		},
		{
			name:    "device listing failure",
			fw:      &workingFirewall{},
			devices: failingDeviceList,
			err:     errors.ErrOnPurpose,
			routing: workingRoutingSetup{},
		},
		{
			name:    "success",
			fw:      &workingFirewall{},
			devices: workingDeviceList,
			routing: workingRoutingSetup{},
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
			err:  errors.ErrOnPurpose,
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
			routing:          workingRoutingSetup{},
			rt:               workingRouter{},
			fw:               &workingFirewall{},
			allowlistRouting: workingAllowlistRouting{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: errors.ErrOnPurpose,
		},
		{
			name:             "router failure",
			devices:          workingDeviceList,
			routing:          workingRoutingSetup{},
			rt:               failingRouter{},
			fw:               &workingFirewall{},
			allowlistRouting: workingAllowlistRouting{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: errors.ErrOnPurpose,
		},
		{
			name:             "firewall failure",
			devices:          workingDeviceList,
			routing:          workingRoutingSetup{},
			rt:               workingRouter{},
			fw:               failingFirewall{},
			allowlistRouting: workingAllowlistRouting{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: errors.ErrOnPurpose,
		},
		{
			name:             "invalid allowlist",
			devices:          workingDeviceList,
			routing:          workingRoutingSetup{},
			rt:               workingRouter{},
			fw:               &workingFirewall{},
			allowlistRouting: &workingAllowlistRouting{},
			allowlist:        config.NewAllowlist(nil, nil, nil),
		},
		{
			name:             "success",
			devices:          workingDeviceList,
			routing:          workingRoutingSetup{},
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
			err:       errors.ErrOnPurpose,
		},
		{
			name:      "router failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        failingRouter{},
			err:       errors.ErrOnPurpose,
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
				workingRoutingSetup{},
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
			routing:   workingRoutingSetup{},
			err:       errors.ErrOnPurpose,
		},
		{
			name:      "router failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        failingRouter{},
			devices:   workingDeviceList,
			routing:   workingRoutingSetup{},
			err:       errors.ErrOnPurpose,
		},
		{
			name:      "device listing failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			devices:   failingDeviceList,
			routing:   workingRoutingSetup{},
			err:       errors.ErrOnPurpose,
		},
		{
			name:      "success",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			routing:   workingRoutingSetup{},
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
				nil,
				0,
				false,
			)
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
			err:       errors.ErrOnPurpose,
		},
		{
			name:      "router failure",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        failingRouter{},
			err:       errors.ErrOnPurpose,
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
				workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
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
			name:      "a1",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:   "100.100.10.1",
			ruleName:  "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
		},
		{
			name:      "a2",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "a70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:   "100.100.10.1",
			ruleName:  "a70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
		},
		{
			name:      "a3",
			fw:        &workingFirewall{},
			allowlist: workingAllowlistRouting{},
			rt:        workingRouter{},
			publicKey: "a2513324-7bac-4dcc-b059-e12df48d7418",
			address:   "100.100.10.1",
			ruleName:  "a2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
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
				workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			uniqueAddress := meshnet.UniqueAddress{UID: test.publicKey, Address: netip.MustParseAddr(test.address)}
			err := netw.AllowIncoming(uniqueAddress)
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
				workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
				false,
			)
			uniqueAddress := meshnet.UniqueAddress{UID: test.publicKey, Address: netip.MustParseAddr(test.address)}
			err := netw.AllowIncoming(uniqueAddress)
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
				workingMesh{},
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				&workingIpv6{},
				test.fw,
				test.allowlist,
				workingDeviceList,
				workingRoutingSetup{},
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
				workingMesh{},
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				&workingIpv6{},
				test.fw,
				test.allowlist,
				workingDeviceList,
				workingRoutingSetup{},
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

func TestCombined_allowIncoming(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		address  string
		fw       firewall.Service
	}{
		{
			name:     "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:  "100.100.10.1",
			ruleName: "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
		{
			name:     "a70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:  "100.100.10.1",
			ruleName: "a70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
			fw:       meshnetterFirewall{},
		},
		{
			name:     "a2513324-7bac-4dcc-b059-e12df48d7418",
			address:  "100.100.10.1",
			ruleName: "a2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
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
			netw.allowIncoming(test.name, netip.MustParseAddr(test.address))
			assert.Equal(t, netw.rules[0], test.ruleName)
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
			netw.allowIncoming(test.name, netip.MustParseAddr(test.address))
			assert.Equal(t, netw.rules[0], test.ruleName)
			netw.BlockIncoming(meshnet.UniqueAddress{UID: test.name, Address: netip.MustParseAddr(test.address)})
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
			err := netw.allowIncoming(test.name, netip.MustParseAddr(test.address))
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
			expectedErrorMsg := fmt.Sprintf("Allow rule does not exist for %s", test.ruleName)
			err := netw.BlockIncoming(meshnet.UniqueAddress{UID: test.name, Address: netip.MustParseAddr(test.address)})
			assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		})
	}
}

func TestCombined_allowExistingRuleFail(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		address  string
		fw       firewall.Service
	}{
		{
			name:     "ec30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:  "100.100.10.1",
			ruleName: "ec30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
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
			err := netw.allowIncoming(test.name, netip.MustParseAddr(test.address))
			assert.Equal(t, nil, err)
			assert.Equal(t, netw.rules[0], test.ruleName)
			// Should fail to add rule second time
			expectedErrorMsg := fmt.Sprintf("allow rule already exist for %s", test.ruleName)
			err = netw.allowIncoming(test.name, netip.MustParseAddr(test.address))
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
		workingMesh{},
		workingGateway{},
		&subs.Subject[string]{},
		workingRouter{},
		&workingDNS{},
		&workingIpv6{},
		fw,
		nil,
		workingDeviceList,
		workingRoutingSetup{},
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

	assert.Equal(t, 5, len(fw.rules), "%d firewall rules were configured, expected 5", len(fw.rules))

	defaultMeshBlockRuleName := "default-mesh-block"

	expectedDefaultMeshBlockFwRule := firewall.Rule{
		Name:           defaultMeshBlockRuleName,
		Direction:      firewall.Inbound,
		RemoteNetworks: []netip.Prefix{defaultMeshSubnet},
		Allow:          false,
	}

	assert.Equal(t, expectedDefaultMeshBlockFwRule, fw.rules[defaultMeshBlockRuleName],
		"default-mesh-block rule is incorrectly configured, \nexpected config: \n%v, \nactual config: \n%v",
		expectedDefaultMeshBlockFwRule, fw.rules[defaultMeshBlockRuleName])

	expectedDefaultMeshAllowEstablishedFwRule := firewall.Rule{
		Name:           "default-mesh-allow-established",
		Direction:      firewall.Inbound,
		RemoteNetworks: []netip.Prefix{defaultMeshSubnet},
		ConnectionStates: []firewall.ConnectionState{
			firewall.Related,
			firewall.Established,
		},
		Allow: true,
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
		testvpn.Working{},
		nil,
		nil,
		&subs.Subject[string]{},
		workingRouter{},
		dns,
		&workingIpv6{},
		newWorkingFirewall(),
		workingAllowlistRouting{},
		workingDeviceList,
		workingRoutingSetup{},
		nil,
		workingRouter{},
		nil,
		nil,
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
