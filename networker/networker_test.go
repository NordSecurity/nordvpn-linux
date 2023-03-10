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

type workingDNS struct{}

func (workingDNS) Set(string, []string) error { return nil }
func (workingDNS) Unset(string) error         { return nil }

type failingDNS struct{}

func (failingDNS) Set(string, []string) error { return errors.ErrOnPurpose }
func (failingDNS) Unset(string) error         { return errors.ErrOnPurpose }

type workingIpv6 struct{}

func (workingIpv6) Block() error   { return nil }
func (workingIpv6) Unblock() error { return nil }

type workingFirewall struct{}

func (workingFirewall) Add([]firewall.Rule) error { return nil }
func (workingFirewall) Delete([]string) error     { return nil }
func (workingFirewall) Enable() error             { return nil }
func (workingFirewall) Disable() error            { return nil }
func (workingFirewall) IsEnabled() bool           { return true }

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

type workingExitNode struct{}

func (workingExitNode) Enable() error                      { return nil }
func (workingExitNode) ResetPeers(mesh.MachinePeers) error { return nil }
func (workingExitNode) DisablePeer(netip.Addr) error       { return nil }
func (workingExitNode) Disable() error                     { return nil }

type workingMesh struct{}

func (workingMesh) Enable(netip.Addr, string) error { return nil }
func (workingMesh) Disable() error                  { return nil }
func (workingMesh) IsActive() bool                  { return false }
func (workingMesh) Refresh(mesh.MachineMap) error   { return nil }
func (workingMesh) Tun() tunnel.T                   { return testtunnel.Working{} }
func (workingMesh) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}

type workingHostSetter struct{}

func (workingHostSetter) SetHosts(dns.Hosts) error { return nil }
func (workingHostSetter) UnsetHosts() error        { return nil }

func TestCombined_Start(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		gateway         routes.GatewayRetriever
		whitelistRouter routes.Service
		dns             dns.Setter
		vpn             vpn.VPN
		fw              firewall.Service
		devices         device.ListFunc
		routing         routes.PolicyService
		err             error
	}{
		{
			name:            "nil vpn",
			gateway:         workingGateway{},
			whitelistRouter: workingRouter{},
			dns:             workingDNS{},
			vpn:             nil,
			fw:              workingFirewall{},
			devices:         workingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errNilVPN,
		},
		{
			name:            "vpn start failure",
			gateway:         workingGateway{},
			whitelistRouter: workingRouter{},
			dns:             workingDNS{},
			vpn:             testvpn.Failing{},
			fw:              workingFirewall{},
			devices:         workingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errors.ErrOnPurpose,
		},
		{
			name:            "firewall failure",
			gateway:         workingGateway{},
			whitelistRouter: workingRouter{},
			dns:             workingDNS{},
			vpn:             testvpn.WorkingInactive{},
			fw:              failingFirewall{},
			devices:         workingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errors.ErrOnPurpose,
		},
		{
			name:            "dns failure",
			gateway:         workingGateway{},
			whitelistRouter: workingRouter{},
			dns:             failingDNS{},
			vpn:             testvpn.WorkingInactive{},
			fw:              workingFirewall{},
			devices:         workingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errors.ErrOnPurpose,
		},
		{
			name:            "device listing failure",
			gateway:         workingGateway{},
			whitelistRouter: workingRouter{},
			dns:             workingDNS{},
			vpn:             testvpn.WorkingInactive{},
			fw:              workingFirewall{},
			devices:         failingDeviceList,
			routing:         workingRoutingSetup{},
			err:             errors.ErrOnPurpose,
		},
		{
			name:            "successful start",
			gateway:         workingGateway{},
			whitelistRouter: workingRouter{},
			dns:             workingDNS{},
			vpn:             testvpn.Working{},
			fw:              workingFirewall{},
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
				test.whitelistRouter,
				test.dns,
				&workingIpv6{},
				test.fw,
				test.devices,
				test.routing,
				nil,
				workingRouter{},
				nil,
				nil,
				0,
			)
			err := netw.Start(
				vpn.Credentials{},
				vpn.ServerData{},
				config.NewWhitelist(nil, nil, nil),
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
			dns:  workingDNS{},
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
			dns:  workingDNS{},
			err:  errors.ErrOnPurpose,
		},
		{
			name: "successful stop",
			vpn:  testvpn.Working{},
			dns:  workingDNS{},
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
				workingFirewall{},
				nil,
				workingRoutingSetup{},
				nil,
				workingRouter{},
				nil,
				nil,
				0,
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
			netw := NewCombined(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0)
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
			dns:         workingDNS{},
			nameservers: []string{},
			hasError:    false,
		},
		{
			name:        "nil nameservers",
			dns:         workingDNS{},
			nameservers: nil,
			hasError:    false,
		},
		{
			name:        "two nameservers",
			dns:         workingDNS{},
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
				workingFirewall{},
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
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
			dns:      workingDNS{},
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
				workingFirewall{},
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
			)
			netw.vpnet = testvpn.Working{}
			err := netw.unsetDNS()
			assert.Equal(t, test.hasError, err != nil)
		})
	}
}

func TestCombined_ResetWhitelist(t *testing.T) {
	category.Set(t, category.Unit)

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
			routing: workingRoutingSetup{},
			err:     errors.ErrOnPurpose,
		},
		{
			name:    "device listing failure",
			fw:      workingFirewall{},
			devices: failingDeviceList,
			err:     errors.ErrOnPurpose,
			routing: workingRoutingSetup{},
		},
		{
			name:    "success",
			fw:      workingFirewall{},
			devices: workingDeviceList,
			routing: workingRoutingSetup{},
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
				workingDNS{},
				workingIpv6{},
				test.fw,
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
			)
			assert.ErrorIs(t, netw.resetWhitelist(), test.err)
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
			fw:      workingFirewall{},
			devices: failingDeviceList,
			err:     errors.ErrOnPurpose,
			routing: workingRoutingSetup{},
		},
		{
			name:    "success",
			fw:      workingFirewall{},
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
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
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
			fw:   workingFirewall{},
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
				0,
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
			fw:      workingFirewall{},
			devices: failingDeviceList,
			err:     errors.ErrOnPurpose,
			routing: workingRoutingSetup{},
		},
		{
			name:    "success",
			fw:      workingFirewall{},
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
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
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
			fw:   workingFirewall{},
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
				0,
			)
			assert.ErrorIs(t, netw.stopAllowedIPv6Traffic(), test.err)
		})
	}
}

func TestCombined_SetWhitelist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		devices   device.ListFunc
		routing   routes.PolicyService
		rt        routes.Service
		fw        firewall.Service
		whitelist config.Whitelist
		err       error
	}{
		{
			name:    "device listing failure",
			devices: failingDeviceList,
			routing: workingRoutingSetup{},
			rt:      workingRouter{},
			fw:      workingFirewall{},
			whitelist: config.NewWhitelist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: errors.ErrOnPurpose,
		},
		{
			name:    "router failure",
			devices: workingDeviceList,
			routing: workingRoutingSetup{},
			rt:      failingRouter{},
			fw:      workingFirewall{},
			whitelist: config.NewWhitelist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: errors.ErrOnPurpose,
		},
		{
			name:    "firewall failure",
			devices: workingDeviceList,
			routing: workingRoutingSetup{},
			rt:      workingRouter{},
			fw:      failingFirewall{},
			whitelist: config.NewWhitelist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: errors.ErrOnPurpose,
		},
		{
			name:      "invalid whitelist",
			devices:   workingDeviceList,
			routing:   workingRoutingSetup{},
			rt:        workingRouter{},
			fw:        workingFirewall{},
			whitelist: config.NewWhitelist(nil, nil, nil),
		},
		{
			name:    "success",
			devices: workingDeviceList,
			routing: workingRoutingSetup{},
			rt:      workingRouter{},
			fw:      workingFirewall{},
			whitelist: config.NewWhitelist(
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
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
			)
			assert.ErrorIs(t, netw.setWhitelist(test.whitelist), test.err)
		})
	}
}

func TestCombined_UnsetWhitelist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		fw   firewall.Service
		rt   routes.Service
		err  error
	}{
		{
			name: "firewall failure",
			fw:   failingFirewall{},
			rt:   workingRouter{},
			err:  errors.ErrOnPurpose,
		},
		{
			name: "router failure",
			fw:   workingFirewall{},
			rt:   failingRouter{},
			err:  errors.ErrOnPurpose,
		},
		{
			name: "success",
			fw:   workingFirewall{},
			rt:   workingRouter{},
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
				workingDeviceList,
				workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
			)
			err := netw.unsetWhitelist()
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestCombined_SetNetwork(t *testing.T) {
	category.Set(t, category.Unit)

	UDPPorts := []int64{550, 200, 100}
	TCPPorts := []int64{220, 35}

	tests := []struct {
		name    string
		fw      firewall.Service
		rt      routes.Service
		devices device.ListFunc
		routing routes.PolicyService
		err     error
	}{
		{
			name:    "firewall failure",
			fw:      failingFirewall{},
			rt:      workingRouter{},
			devices: workingDeviceList,
			routing: workingRoutingSetup{},
			err:     errors.ErrOnPurpose,
		},
		{
			name:    "router failure",
			fw:      workingFirewall{},
			rt:      failingRouter{},
			devices: workingDeviceList,
			routing: workingRoutingSetup{},
			err:     errors.ErrOnPurpose,
		},
		{
			name:    "device listing failure",
			fw:      workingFirewall{},
			rt:      workingRouter{},
			devices: failingDeviceList,
			routing: workingRoutingSetup{},
			err:     errors.ErrOnPurpose,
		},
		{
			name:    "success",
			fw:      workingFirewall{},
			rt:      workingRouter{},
			routing: workingRoutingSetup{},
			devices: workingDeviceList,
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
				test.devices,
				test.routing,
				nil,
				nil,
				nil,
				nil,
				0,
			)
			err := netw.setNetwork(
				config.NewWhitelist(
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
		name string
		fw   firewall.Service
		rt   routes.Service
		err  error
	}{
		{
			name: "firewall failure",
			fw:   failingFirewall{},
			rt:   workingRouter{},
			err:  errors.ErrOnPurpose,
		},
		{
			name: "router failure",
			fw:   workingFirewall{},
			rt:   failingRouter{},
			err:  errors.ErrOnPurpose,
		},
		{
			name: "success",
			fw:   workingFirewall{},
			rt:   workingRouter{},
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
				workingDeviceList,
				workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
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
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			name:      "a1",
			fw:        workingFirewall{},
			rt:        workingRouter{},
			publicKey: "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:   "100.100.10.1",
			ruleName:  "ac30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
		},
		{
			name:      "a2",
			fw:        workingFirewall{},
			rt:        workingRouter{},
			publicKey: "a70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:   "100.100.10.1",
			ruleName:  "a70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
		},
		{
			name:      "a3",
			fw:        workingFirewall{},
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
				workingDeviceList,
				workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
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
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			name:      "b1",
			fw:        workingFirewall{},
			rt:        workingRouter{},
			publicKey: "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e",
			address:   "100.100.10.1",
			ruleName:  "bc30c01d-9ab8-4b25-9d5f-8a4bb2c5c78e-allow-rule-100.100.10.1",
		},
		{
			name:      "b2",
			fw:        workingFirewall{},
			rt:        workingRouter{},
			publicKey: "b70ad213-fa09-4ae4-890b-bea12697b9f0",
			address:   "100.100.10.1",
			ruleName:  "b70ad213-fa09-4ae4-890b-bea12697b9f0-allow-rule-100.100.10.1",
		},
		{
			name:      "b3",
			fw:        workingFirewall{},
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
				workingDeviceList,
				workingRoutingSetup{},
				nil,
				nil,
				nil,
				nil,
				0,
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
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			fw:        workingFirewall{},
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
				workingDeviceList,
				workingRoutingSetup{},
				workingHostSetter{},
				workingRouter{},
				workingRouter{},
				workingExitNode{},
				0,
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
		rt        routes.Service
		publicKey string
		address   string
		ruleName  string
		err       error
	}{
		{
			fw:        workingFirewall{},
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
				workingDeviceList,
				workingRoutingSetup{},
				workingHostSetter{},
				workingRouter{},
				workingRouter{},
				workingExitNode{},
				0,
			)
			netw.isMeshnetSet = true
			assert.ErrorIs(t, test.err, netw.UnSetMesh())
		})
	}
}

func TestCombined_allow(t *testing.T) {
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
				0,
			)
			netw.allow(test.name, netip.MustParseAddr(test.address))
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
				0,
			)
			netw.allow(test.name, netip.MustParseAddr(test.address))
			assert.Equal(t, netw.rules[0], test.ruleName)
			netw.block(test.name, netip.MustParseAddr(test.address))
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
				0,
			)
			err := netw.allow(test.name, netip.MustParseAddr(test.address))
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
				0,
			)
			// Should fail to block rule non existing
			expectedErrorMsg := fmt.Sprintf("Allow rule does not exist for %s", test.ruleName)
			err := netw.block(test.name, netip.MustParseAddr(test.address))
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
				0,
			)
			err := netw.allow(test.name, netip.MustParseAddr(test.address))
			assert.Equal(t, nil, err)
			assert.Equal(t, netw.rules[0], test.ruleName)
			// Should fail to add rule second time
			expectedErrorMsg := fmt.Sprintf("allow rule already exist for %s", test.ruleName)
			err = netw.allow(test.name, netip.MustParseAddr(test.address))
			assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		})
	}
}
