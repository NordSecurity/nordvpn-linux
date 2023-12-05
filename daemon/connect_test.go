package daemon

import (
	"net"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"

	"github.com/stretchr/testify/assert"
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

type UniqueAddress struct{}

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
			netw:      &testnetworker.Mock{},
			fw:        &workingFirewall{},
			allowlist: config.NewAllowlist(nil, nil, nil),
			router:    &workingRouter{},
			gateway:   newGatewayMock(netip.Addr{}),
			expected:  ConnectEvent{Code: internal.CodeConnected},
		},
		{
			name:     "successful reconnect",
			netw:     &testnetworker.Mock{},
			fw:       &workingFirewall{},
			router:   &workingRouter{},
			gateway:  newGatewayMock(netip.Addr{}),
			expected: ConnectEvent{Code: internal.CodeConnected},
		},
		{
			name:     "failed connect",
			netw:     testnetworker.Failing{},
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
