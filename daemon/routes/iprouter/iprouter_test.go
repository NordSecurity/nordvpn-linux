package iprouter

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPRouter_Add(t *testing.T) {
	category.Set(t, category.Route)

	gateway, iface, err := routes.IPGatewayRetriever{}.Default(false)
	require.NoError(t, err)

	bits := gateway.As4()
	bits[3]++

	defaultRoute := route(t, gateway, netip.AddrFrom4([4]byte{bits[0], bits[1], bits[2], 0}), 24)
	defaultRoute.Device = iface

	tests := []struct {
		name              string
		preExistingRoutes []routes.Route
		routes            []routes.Route
		errOn             int
		err               error
	}{
		{
			name:   "no preExistingRoutes",
			routes: []routes.Route{defaultRoute},
			errOn:  -1,
		},
		{
			name:   "route to default route",
			routes: []routes.Route{defaultRoute, defaultRoute},
			errOn:  1,
		},
		{
			name:              "route already exists",
			preExistingRoutes: []routes.Route{defaultRoute},
			routes:            []routes.Route{defaultRoute},
			errOn:             -1,
		},
		{
			name: "subnet already in use",
			preExistingRoutes: []routes.Route{
				{
					Subnet: defaultRoute.Subnet,
					Device: net.Interface{Name: "lo"},
				},
			},
			routes: []routes.Route{
				defaultRoute,
			},
			errOn: 0,
			err:   routes.ErrRouteToOtherDestinationExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			preRouter := Router{}
			for _, route := range test.preExistingRoutes {
				exec.Command("ip", "route", "delete", route.Subnet.String(), "via", route.Gateway.String()).Run()
				err := preRouter.Add(route)

				assert.NoError(t, err)
				out, err := exec.Command("ip", "route").CombinedOutput()
				log.Println(string(out))
			}
			router := Router{}
			for i, route := range test.routes {
				// Ignore errors here. This is just for test preparation
				if test.errOn != i {
					exec.Command("ip", "route", "delete", route.Subnet.String(), "via", route.Gateway.String()).Run()
				}
				err := router.Add(route)
				assert.True(t, (test.errOn == i) == (err != nil), err)
				if test.errOn == i {
					if test.err != nil {
						assert.Equal(t, test.err, err)
					}
					continue
				}
				///
				out, err := exec.Command("ip", "route").CombinedOutput()
				assert.NoError(t, err, string(out))
				assert.Contains(t, string(out), fmt.Sprintf("%s via %s dev", route.Subnet.String(), route.Gateway.String()))
				assert.True(t, router.has(route))
				assert.Equal(t, i+1, len(router.routes))
			}
			// Cleanup
			err = router.Flush()
			assert.NoError(t, err)
			preRouter.Flush()
			assert.Equal(t, 0, len(router.routes))
		})
	}
}

func TestIPRouterAddIPv6(t *testing.T) {
	//@TODO gitlab runner sysctl flag has no effect
	category.Set(t, category.Link)

	ip := "2a01:7e00::f13c:92ff:fe7d:d4d"
	name := "iprouterdummy1"
	iface := net.Interface{Name: name}
	out, err := exec.Command("ip", "link", "add", name, "type", "dummy").CombinedOutput()
	assert.NoError(t, err, string(out))
	out, err = exec.Command("ip", "address", "add", ip, "dev", name).CombinedOutput()
	assert.NoError(t, err, string(out))
	out, err = exec.Command("ip", "link", "set", name, "up").CombinedOutput()
	assert.NoError(t, err, string(out))
	defer exec.Command("ip", "link", "delete", name).CombinedOutput()

	newRoutes := []routes.Route{
		{
			Subnet: netip.MustParsePrefix("2000::/3"),
			Device: iface,
		},
		{
			Gateway: netip.MustParseAddr("fe80::1"),
			Subnet:  netip.MustParsePrefix("2606:4700:4700::1111/128"),
			Device:  iface,
		},
	}

	router := Router{}
	defer router.Flush()

	for _, route := range newRoutes {
		err = router.Add(route)
		assert.NoError(t, err)
		out, err = exec.Command("ip", "-6", "route").CombinedOutput()
		assert.NoError(t, err)
		if route.Gateway == (netip.Addr{}) {
			assert.Contains(t, string(out), fmt.Sprintf("%s via %s dev", network.ToRouteString(route.Subnet), route.Gateway.String()))
		} else {
			assert.Contains(t, string(out), fmt.Sprintf("%s dev %s", network.ToRouteString(route.Subnet), route.Device.Name))
		}
	}

	// missing interface/destination
	route := routes.Route{
		Subnet: netip.MustParsePrefix("fe80::34be:67ff:fef4:3505/128"),
	}
	err = router.Add(route)
	assert.Error(t, err)
}

func TestIPRouter_Has(t *testing.T) {
	category.Set(t, category.Unit)

	exRoute := route(t, netip.MustParseAddr("1.2.3.4"), netip.MustParseAddr("1.2.0.0"), 16)
	lo := net.Interface{Name: "lo"}
	local := net.Interface{Name: "local"}

	existingRoutes := []routes.Route{
		{Device: lo, Subnet: netip.MustParsePrefix("127.0.0.1/32")},
		{Device: lo, Subnet: netip.MustParsePrefix("::1/128")},
	}

	tests := []struct {
		name     string
		list     []routes.Route
		route    routes.Route
		contains bool
	}{
		{
			name:     "nil routing table",
			list:     nil,
			route:    routes.Route{},
			contains: false,
		},
		{
			name:     "empty routing table",
			list:     []routes.Route{},
			route:    routes.Route{},
			contains: false,
		},
		{
			name:     "found in table with single route",
			list:     []routes.Route{exRoute},
			route:    exRoute,
			contains: true,
		},
		{
			name: "found ipv4 in table with multiple routes",
			list: existingRoutes,
			route: routes.Route{
				Device: lo,
				Subnet: netip.MustParsePrefix("127.0.0.1/32"),
			},
			contains: true,
		},
		{
			name: "not found in table with multiple routes",
			list: existingRoutes,
			route: routes.Route{
				Device: local,
				Subnet: netip.MustParsePrefix("127.0.0.1/24"),
			},
			contains: false,
		},
		{
			name: "found ipv6 in table with multiple routes",
			list: existingRoutes,
			route: routes.Route{
				Device: lo,
				Subnet: netip.MustParsePrefix("::1/128"),
			},
			contains: true,
		},
		{
			name:     "found in table with one empty route",
			list:     []routes.Route{{}, {}, exRoute},
			route:    exRoute,
			contains: true,
		},
		{
			name:     "not found in table with empty routes",
			list:     []routes.Route{{}},
			route:    exRoute,
			contains: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := Router{routes: test.list}
			assert.Equal(t, test.contains, router.has(test.route))
		})
	}
}

func route(t *testing.T, destination netip.Addr, maskIP netip.Addr, cidrMask int) routes.Route {
	t.Helper()
	return routes.Route{
		Gateway: destination,
		Subnet:  netip.PrefixFrom(maskIP, cidrMask),
	}
}

func routeInterface(t *testing.T, interfaceName string, network string) routes.Route {
	t.Helper()
	return routes.Route{
		Subnet: netip.MustParsePrefix(network),
		Device: net.Interface{Name: interfaceName},
	}
}

func TestExistsInRoutingTableOutput(t *testing.T) {
	category.Set(t, category.Unit)

	exRoute := route(t, netip.MustParseAddr("1.2.3.4"), netip.MustParseAddr("1.2.0.0"), 16)
	exOut := `default via 1.2.3.4 dev wlan0 proto dhcp metric 600
1.2.0.0/16 via 1.2.3.4 dev wlan0`
	exOutNotFound := `default via 1.2.3.4 dev wlan0 proto dhcp metric 600
1.2.0.0/16 via 1.2.3.5 dev wlan0`

	tests := []struct {
		out    string
		route  routes.Route
		exists bool
	}{
		{
			out:    exOut,
			route:  exRoute,
			exists: true,
		},
		{
			out:    exOutNotFound,
			route:  exRoute,
			exists: false,
		},
		{
			out:    "1.1.1.1 via 192.168.0.101 dev enp39s0 proto static metric 100",
			route:  route(t, netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("1.1.1.1"), 32),
			exists: true,
		},
		{
			out:    "::1 dev lo proto kernel metric 256 pref medium",
			route:  routeInterface(t, "lo", "::1/128"),
			exists: true,
		},
		{
			out:    "2606:4700:4700::1111 dev enp39s0 metric 1024 pref medium",
			route:  routeInterface(t, "enp39s0", "2606:4700:4700::1111/128"),
			exists: true,
		},
		{
			out:    "fe80::/64 dev tun0 proto kernel metric 256 pref medium",
			route:  routeInterface(t, "enp39s0", "fe80::/64"),
			exists: false,
		},
		{
			out:    "fe80::/64 dev enp39s0 proto kernel metric 256 pref medium",
			route:  routeInterface(t, "enp39s0", "fe80::/64"),
			exists: true,
		},
	}

	for _, test := range tests {
		t.Run(test.out, func(t *testing.T) {
			assert.Equal(t, test.exists,
				existsInRoutingTableOutput([]byte(test.out), test.route),
			)
		})
	}
}

func TestExistsInRoutingTable(t *testing.T) {
	category.Set(t, category.Route)

	gateway, iface, err := routes.IPGatewayRetriever{}.Default(false)
	require.NoError(t, err)

	bits := gateway.As4()

	route := route(t, gateway, netip.AddrFrom4([4]byte{bits[0], bits[1], bits[2], 0}), 24)
	route.Device = iface
	router := Router{}

	// Cleanup before the execution
	_ = exec.Command("ip", "route", "delete", route.Subnet.String(), "via", route.Gateway.String()).Run()
	require.NoError(t, err)

	// Just assume that such route does not exist in current system
	exists, err := existsInRoutingTable(route)
	require.NoError(t, err)
	require.False(t, exists)
	err = router.Add(route)
	require.NoError(t, err)

	// Route should have appeared in the system
	exists, err = existsInRoutingTable(route)
	require.NoError(t, err)
	require.True(t, exists)

	// Cleanup
	err = router.Flush()
	require.NoError(t, err)
}
