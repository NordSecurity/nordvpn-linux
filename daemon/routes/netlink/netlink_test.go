package netlink

import (
	"net"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/vishvananda/netlink"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_Add(t *testing.T) {
	category.Set(t, category.Route)

	gateway, iface, err := Retriever{}.Retrieve(netip.Prefix{}, 0)
	require.NoError(t, err)

	bits := gateway.As4()
	bits[3]++

	defaultRoute := route(t, gateway, netip.AddrFrom4([4]byte{bits[0], bits[1], bits[2], 0}), 32)
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
					Device: iface,
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
				err = preRouter.Add(route)
				require.NoError(t, err)
			}
			router := Router{}
			for i, route := range test.routes {
				if test.errOn != i {
					// Remove route in case it already exists. Error can be
					// ignored here
					netlinkRoute := toNetlinkRoute(route)
					netlink.RouteDel(&netlinkRoute)
				}
				err := router.Add(route)
				assert.True(t, (test.errOn == i) == (err != nil), err)
				if test.errOn == i {
					if test.err != nil {
						assert.Equal(t, test.err, err)
					}
					continue
				}
				checkRouter := Router{}
				assert.ErrorIs(
					t,
					checkRouter.Add(route),
					routes.ErrRouteToOtherDestinationExists,
				)
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

func TestRouter_Has(t *testing.T) {
	category.Set(t, category.Unit)

	exRoute := route(t, netip.MustParseAddr("1.2.3.4"), netip.MustParseAddr("1.2.0.0"), 16)
	lo := net.Interface{Name: "lo"}
	local := net.Interface{Name: "local"}

	existingRoutes := []routes.Route{
		{Device: lo, Subnet: netip.MustParsePrefix("127.0.0.1/32")},
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
