package networker

import (
	"errors"
	"testing"

	"context"
	"net"
	"net/netip"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	firewallmock "github.com/NordSecurity/nordvpn-linux/test/mock/firewall"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
	mapset "github.com/deckarep/golang-set/v2"
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
		firewallmock.NewFirewall(),
		workingDeviceList,
		&workingRoutingSetup{},
		&workingHostSetter{},
		workingRouter{},
		workingRouter{},
		0,
		false,
		&workingIpv6{},
		false,
		&mock.SysctlSetterMock{},
		config.Allowlist{},
		&mock.SysctlSetterMock{},
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

func workingDeviceList(mapset.Set[string]) mapset.Set[string] {
	return mapset.NewSet(mock.En0Interface.Name)
}

type workingRoutingSetup struct {
	EnableLocalTraffic bool
}

func (r *workingRoutingSetup) SetupRoutingRules(enableLan bool, _ bool, _ []string) error {
	r.EnableLocalTraffic = enableLan
	return nil
}
func (*workingRoutingSetup) CleanupRouting() error { return nil }
func (*workingRoutingSetup) TableID() uint         { return 0 }
func (*workingRoutingSetup) Enable() error         { return nil }
func (*workingRoutingSetup) Disable() error        { return nil }
func (*workingRoutingSetup) IsEnabled() bool       { return true }

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

var noopNetworkerCallback = func(startTime time.Time, err error) {}

func TestCombined_Start(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name                     string
		gateway                  routes.GatewayRetriever
		allowlistRouter          routes.Service
		dns                      dns.Setter
		vpn                      vpn.VPN
		fw                       firewall.Service
		arpIgnore                bool
		expectedARPIgnore        bool
		disconnectCallbackCalled bool
		err                      error
		setARPIgnoreErr          error
	}{
		{
			name:            "nil vpn",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             nil,
			fw:              firewallmock.NewFirewall(),
			err:             errNilVPN,
		},
		{
			name:            "vpn start failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             mock.FailingVPN{},
			fw:              firewallmock.NewFirewall(),
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "firewall failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             mock.WorkingInactiveVPN{},
			fw:              &firewallmock.Firewall{Err: mock.ErrOnPurpose},
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "dns failure",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             failingDNS{},
			vpn:             mock.WorkingInactiveVPN{},
			fw:              firewallmock.NewFirewall(),
			err:             mock.ErrOnPurpose,
		},
		{
			name:            "successful start",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             &mock.WorkingVPN{},
			fw:              firewallmock.NewFirewall(),
			err:             nil,
		},
		{
			name:                     "restart",
			gateway:                  workingGateway{},
			allowlistRouter:          workingRouter{},
			dns:                      &workingDNS{},
			vpn:                      &mock.ActiveVPN{},
			fw:                       firewallmock.NewFirewall(),
			disconnectCallbackCalled: true,
			err:                      nil,
		},
		{
			name:              "start with arp ignore",
			gateway:           workingGateway{},
			allowlistRouter:   workingRouter{},
			dns:               &workingDNS{},
			vpn:               &mock.WorkingVPN{},
			fw:                firewallmock.NewFirewall(),
			err:               nil,
			arpIgnore:         true,
			expectedARPIgnore: true,
		},
		{
			name:                     "restart with arp ignore",
			gateway:                  workingGateway{},
			allowlistRouter:          workingRouter{},
			dns:                      &workingDNS{},
			vpn:                      &mock.ActiveVPN{},
			fw:                       firewallmock.NewFirewall(),
			err:                      nil,
			arpIgnore:                true,
			expectedARPIgnore:        true,
			disconnectCallbackCalled: true,
		},
		{
			name:            "start with arp ignore set error",
			gateway:         workingGateway{},
			allowlistRouter: workingRouter{},
			dns:             &workingDNS{},
			vpn:             &mock.WorkingVPN{},
			fw:              firewallmock.NewFirewall(),
			err:             mock.ErrOnPurpose,
			arpIgnore:       true,
			setARPIgnoreErr: mock.ErrOnPurpose,
		},
		{
			name:                     "restart with arp ignore set error",
			gateway:                  workingGateway{},
			allowlistRouter:          workingRouter{},
			dns:                      &workingDNS{},
			vpn:                      &mock.ActiveVPN{},
			fw:                       firewallmock.NewFirewall(),
			err:                      mock.ErrOnPurpose,
			arpIgnore:                true,
			setARPIgnoreErr:          mock.ErrOnPurpose,
			disconnectCallbackCalled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arpIgnoreSetter := mock.SysctlSetterMock{
				SetErr: test.setARPIgnoreErr,
			}
			netw := NewCombined(
				test.vpn,
				nil,
				test.gateway,
				&subs.Subject[string]{},
				test.allowlistRouter,
				test.dns,
				test.fw,
				nil,
				&workingRoutingSetup{},
				nil,
				workingRouter{},
				nil,
				0,
				false,
				&workingIpv6{},
				test.arpIgnore,
				&arpIgnoreSetter,
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)

			disconnectCallbackCalled := false

			err := netw.Start(
				context.Background(),
				vpn.Credentials{},
				vpn.ServerData{},
				config.NewAllowlist(nil, nil, nil),
				[]string{"1.1.1.1"},
				true,
				func(startTime time.Time, err error) {
					disconnectCallbackCalled = true
				},
			)
			assert.Equal(t, test.disconnectCallbackCalled, disconnectCallbackCalled)
			assert.ErrorIs(t, err, test.err, test.name)
			assert.Equal(t, test.expectedARPIgnore, arpIgnoreSetter.IsSet, "ARP ignore not set to expected value.")
		})
	}
}

func TestCombined_Stop(t *testing.T) {
	category.Set(t, category.Link)

	tests := []struct {
		name              string
		vpn               vpn.VPN
		dns               dns.Setter
		arpIgnore         bool
		expectedARPIgnore bool
		arpIgnoreUnsetErr error
		err               error
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
			name:              "unset arp ignore failure",
			vpn:               &mock.WorkingVPN{},
			dns:               &workingDNS{},
			arpIgnore:         true,
			expectedARPIgnore: true,
			arpIgnoreUnsetErr: mock.ErrOnPurpose,
			err:               mock.ErrOnPurpose,
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
		{
			name:              "successful stop with ARP ignore",
			vpn:               &mock.WorkingVPN{},
			dns:               &workingDNS{},
			arpIgnore:         true,
			expectedARPIgnore: false,
			err:               nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arpIgnoreSetter := mock.SysctlSetterMock{
				// set current ARP ignore to simulate config that would be applied when VPN was started
				IsSet:    test.arpIgnore,
				UnsetErr: test.arpIgnoreUnsetErr,
			}
			netw := NewCombined(
				test.vpn,
				nil,
				workingGateway{},
				&subs.Subject[string]{},
				workingRouter{},
				test.dns,
				firewallmock.NewFirewall(),
				nil,
				&workingRoutingSetup{},
				nil,
				workingRouter{},
				nil,
				0,
				false,
				&workingIpv6{},
				test.arpIgnore,
				&arpIgnoreSetter,
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)

			netw.vpnet = test.vpn
			err := netw.stop()
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.expectedARPIgnore, arpIgnoreSetter.IsSet, "ARP ignore not set to expected value.")
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
				firewallmock.NewFirewall(),
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
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
				firewallmock.NewFirewall(),
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)
			err := netw.UnsetDNS()
			assert.Equal(t, test.hasError, err != nil)
		})
	}
}

func TestCombined_ResetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		fw      firewall.Service
		routing routes.PolicyService
		err     error
	}{
		{
			name:    "success",
			fw:      firewallmock.NewFirewall(),
			routing: &workingRoutingSetup{},
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
				test.fw,
				nil,
				test.routing,
				nil,
				nil,
				nil,
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)
			assert.ErrorIs(t, netw.resetAllowlist(), test.err)
		})
	}
}

func TestCombined_SetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		routing   routes.PolicyService
		rt        routes.Service
		allowlist config.Allowlist
		err       error
	}{
		{
			name:    "router failure",
			routing: &workingRoutingSetup{},
			rt:      failingRouter{},
			allowlist: config.NewAllowlist(
				[]int64{22}, []int64{22}, []string{"1.1.1.1/32"},
			),
			err: nil, // not connected - router is not invoked, no error
		},
		{
			name:      "invalid allowlist",
			routing:   &workingRoutingSetup{},
			rt:        workingRouter{},
			allowlist: config.NewAllowlist(nil, nil, nil),
		},
		{
			name:    "success",
			routing: &workingRoutingSetup{},
			rt:      workingRouter{},
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
				firewallmock.NewFirewall(),
				nil,
				test.routing,
				nil,
				nil,
				nil,
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)
			assert.ErrorIs(t, netw.setAllowlist(test.allowlist), test.err)
		})
	}
}

func TestCombined_UnsetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		rt   routes.Service
		err  error
	}{
		{
			name: "router failure",
			rt:   failingRouter{},
		},
		{
			name: "success",
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
				firewallmock.NewFirewall(),
				workingDeviceList,
				&workingRoutingSetup{},
				nil,
				nil,
				nil,
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
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
		name    string
		rt      routes.Service
		devices device.ListFunc
		routing routes.PolicyService
		err     error
	}{
		{
			name:    "router failure",
			rt:      failingRouter{},
			routing: &workingRoutingSetup{},
			err:     nil, // not connected - router is not invoked, no error
		},
		{
			name:    "success",
			rt:      workingRouter{},
			routing: &workingRoutingSetup{},
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
				firewallmock.NewFirewall(),
				workingDeviceList,
				test.routing,
				nil,
				nil,
				nil,
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
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
		name string
		rt   routes.Service
		err  error
	}{
		{
			name: "router failure",
			rt:   failingRouter{},
		},
		{
			name: "success",
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
				firewallmock.NewFirewall(),
				workingDeviceList,
				&workingRoutingSetup{},
				nil,
				nil,
				nil,
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)
			assert.ErrorIs(t, netw.unsetNetwork(), test.err)
		})
	}
}

func TestCombined_SetMesh(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		fw        firewall.Service
		rt        routes.Service
		publicKey string
		address   string
		err       error
	}{
		{
			name:      "works setting meshnet",
			fw:        firewallmock.NewFirewall(),
			rt:        workingRouter{},
			publicKey: "c2513324-7bac-4dcc-b059-e12df48d7418",
			address:   "100.100.10.1",
		},
		{
			name:      "fails when firewall returns error",
			fw:        &firewallmock.Firewall{Err: mock.ErrOnPurpose},
			rt:        workingRouter{},
			publicKey: "c2513324-7bac-4dcc-b059-e12df48d7418",
			address:   "100.100.10.1",
			err:       mock.ErrOnPurpose,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				&workingMesh{},
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				test.fw,
				workingDeviceList,
				&workingRoutingSetup{},
				&workingHostSetter{},
				workingRouter{},
				workingRouter{},
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)
			err := netw.SetMesh(
				mesh.MachineMap{Peers: mesh.MachinePeers{mesh.MachinePeer{DoIAllowInbound: false}}},
				netip.Addr{},
				"",
			)

			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestCombined_UnSetMesh(t *testing.T) {
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
			name:      "working unset",
			fw:        firewallmock.NewFirewall(),
			rt:        workingRouter{},
			publicKey: "d2513324-7bac-4dcc-b059-e12df48d7418",
			address:   "100.100.10.1",
			ruleName:  "d2513324-7bac-4dcc-b059-e12df48d7418-allow-rule-100.100.10.1",
		},
		{
			name:      "fails when firewall returns error",
			fw:        &firewallmock.Firewall{Err: mock.ErrOnPurpose},
			rt:        workingRouter{},
			publicKey: "c2513324-7bac-4dcc-b059-e12df48d7418",
			address:   "100.100.10.1",
			err:       mock.ErrOnPurpose,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := NewCombined(
				nil,
				&workingMesh{},
				workingGateway{},
				&subs.Subject[string]{},
				test.rt,
				&workingDNS{},
				test.fw,
				workingDeviceList,
				&workingRoutingSetup{},
				&workingHostSetter{},
				workingRouter{},
				workingRouter{},
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)
			netw.isMeshnetSet = true
			assert.ErrorIs(t, netw.UnSetMesh(), test.err)
		})
	}
}

func TestCombined_Reconnect(t *testing.T) {
	category.Set(t, category.Unit)

	// on refresh keep `enableLocalTraffic` value;
	// on UnsetMesh get default value `true`

	router := &workingRoutingSetup{}
	fw := firewallmock.NewFirewall()

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
				fw,
				workingDeviceList,
				router,
				&workingHostSetter{},
				workingRouter{},
				workingRouter{},
				0,
				false,
				&workingIpv6{},
				false,
				&mock.SysctlSetterMock{},
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)
			// activate meshnet
			errSet := netw.SetMesh(
				mesh.MachineMap{},
				netip.Addr{},
				"",
			)
			assert.ErrorIs(t, errSet, test.err)
			// connect to exit node
			_ = netw.Start(
				context.Background(),
				vpn.Credentials{},
				vpn.ServerData{},
				config.NewAllowlist(nil, nil, nil),
				[]string{"1.1.1.1"},
				test.enableLan,
				noopNetworkerCallback,
			)

			// simulate network change event and refreshVPN
			netw.Reconnect(true)
			assert.Equal(t, test.enableLan, router.EnableLocalTraffic)
		})
	}
}

func TestCombined_Refresh(t *testing.T) {
	hostSetter := newMockHostSetter()

	netw := NewCombined(
		nil,
		&workingMesh{},
		workingGateway{},
		&subs.Subject[string]{},
		workingRouter{},
		&workingDNS{},
		firewallmock.NewFirewall(),
		workingDeviceList,
		&workingRoutingSetup{},
		hostSetter,
		workingRouter{},
		workingRouter{},
		0,
		false,
		&workingIpv6{},
		false,
		&mock.SysctlSetterMock{},
		config.Allowlist{},
		&mock.SysctlSetterMock{},
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

	err := netw.SetMesh(machineMap, netip.MustParseAddr("100.64.0.100"), "key")
	assert.Nil(t, err)

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
		firewallmock.NewFirewall(),
		workingDeviceList,
		&workingRoutingSetup{},
		nil,
		workingRouter{},
		nil,
		0,
		false,
		&workingIpv6{},
		false,
		&mock.SysctlSetterMock{},
		config.Allowlist{},
		&mock.SysctlSetterMock{},
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

func TestCombined_SetARPIgnore(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                      string
		vpn                       vpn.VPN
		currentARPIgnoreNetworker bool
		currentARPIgnoreKernel    bool
		targetARPIgnore           bool
		setErr                    error
		unsetErr                  error
		expectedARPIgnore         bool
		errExpected               bool
	}{
		{
			name:              "unset ignore to set active VPN",
			vpn:               mock.ActiveVPN{},
			targetARPIgnore:   true,
			expectedARPIgnore: true,
			errExpected:       false,
		},
		{
			name:                      "set ignore to unset active VPN",
			vpn:                       mock.ActiveVPN{},
			currentARPIgnoreKernel:    true,
			currentARPIgnoreNetworker: true,
			targetARPIgnore:           false,
			expectedARPIgnore:         false,
			errExpected:               false,
		},
		{
			name:              "set ignore is not applied to kernel when VPN is not active",
			vpn:               mock.WorkingInactiveVPN{},
			targetARPIgnore:   true,
			expectedARPIgnore: false,
			errExpected:       false,
		},
		{
			name:                      "unset ignore is not applied to kernel when VPN is not active",
			vpn:                       mock.WorkingInactiveVPN{},
			currentARPIgnoreKernel:    true,
			currentARPIgnoreNetworker: false,
			targetARPIgnore:           false,
			expectedARPIgnore:         true,
			errExpected:               false,
		},
		{
			name:                      "set is always applied to kernel even if already set when VPN is active",
			vpn:                       mock.ActiveVPN{},
			currentARPIgnoreKernel:    false,
			currentARPIgnoreNetworker: true,
			targetARPIgnore:           true,
			expectedARPIgnore:         true,
			errExpected:               false,
		},
		{
			name:                      "unset is always applied to kernel even if already unset when VPN is active",
			vpn:                       mock.ActiveVPN{},
			currentARPIgnoreKernel:    true,
			currentARPIgnoreNetworker: false,
			targetARPIgnore:           false,
			expectedARPIgnore:         false,
			errExpected:               false,
		},
		{
			name:              "set ignore error",
			vpn:               mock.ActiveVPN{},
			targetARPIgnore:   true,
			setErr:            errors.New("set err"),
			expectedARPIgnore: false,
			errExpected:       true,
		},
		{
			name:                      "unset ignore error",
			vpn:                       mock.ActiveVPN{},
			currentARPIgnoreNetworker: true,
			targetARPIgnore:           false,
			unsetErr:                  errors.New("unset err"),
			expectedARPIgnore:         false,
			errExpected:               true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arpIgnoreSetter := mock.SysctlSetterMock{}
			arpIgnoreSetter.SetErr = test.setErr
			arpIgnoreSetter.UnsetErr = test.unsetErr
			arpIgnoreSetter.IsSet = test.currentARPIgnoreKernel

			netw := NewCombined(
				test.vpn,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				false,
				nil,
				false,
				&arpIgnoreSetter,
				config.Allowlist{},
				&mock.SysctlSetterMock{},
			)

			netw.ignoreARP = test.currentARPIgnoreNetworker

			err := netw.SetARPIgnore(test.targetARPIgnore)
			if test.errExpected {
				assert.Error(t, err, "Expected error not returned when setting ARP ignore.")
			} else {
				assert.NoError(t, err, "Unexpected error returned when setting ARP ignore.")
			}

			assert.Equal(t, test.expectedARPIgnore, arpIgnoreSetter.IsSet, "ARP ignore was set to unexpected value.")
		})
	}
}
