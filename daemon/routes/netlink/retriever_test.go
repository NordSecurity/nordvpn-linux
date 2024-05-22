package netlink

import (
	"net"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

func TestOrderRoutesByRules(t *testing.T) {
	category.Set(t, category.Unit)
	for _, test := range []struct {
		name     string
		rules    []netlink.Rule
		routes   []netlink.Route
		links    []netlink.Link
		expected []netlink.Route
	}{
		{
			name:   "default route",
			rules:  []netlink.Rule{rule(netlink.Rule{Table: 254})},
			routes: []netlink.Route{{LinkIndex: 123, Table: 254}},
			links: []netlink.Link{
				&netlink.Device{LinkAttrs: netlink.LinkAttrs{Index: 123}},
			},
			expected: []netlink.Route{{LinkIndex: 123, Table: 254}},
		},
		{
			name:   "no route found",
			rules:  []netlink.Rule{rule(netlink.Rule{Table: 200})},
			routes: []netlink.Route{{LinkIndex: 123, Table: 254}},
			links: []netlink.Link{
				&netlink.Device{LinkAttrs: netlink.LinkAttrs{Index: 123}},
			},
			expected: nil,
		},
		{
			name: "multiple rules applying to same route",
			rules: []netlink.Rule{
				rule(netlink.Rule{
					Table: 254,
				}),
				rule(netlink.Rule{
					Table:  0,
					Invert: true,
				}),
			},
			routes: []netlink.Route{{LinkIndex: 123, Table: 254}},
			links: []netlink.Link{
				&netlink.Device{LinkAttrs: netlink.LinkAttrs{Index: 123}},
			},
			expected: []netlink.Route{{LinkIndex: 123, Table: 254}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			actual := orderRoutesByRules(test.rules, test.routes, test.links)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestRuleAppliesForRoute(t *testing.T) {
	category.Set(t, category.Unit)
	for _, test := range []struct {
		name     string
		rule     netlink.Rule
		route    netlink.Route
		ifgroup  uint32
		expected bool
	}{
		{
			name:     "table matches",
			route:    netlink.Route{Table: 123},
			rule:     netlink.Rule{Table: 123},
			expected: true,
		},
		{
			name:     "table matches inverted",
			route:    netlink.Route{Table: 123},
			rule:     netlink.Rule{Invert: true, Table: 123},
			expected: false,
		},
		{
			name:     "rule with fwmark always ignored",
			route:    netlink.Route{Table: 123},
			rule:     netlink.Rule{Table: 123, Mark: 321},
			expected: false,
		},
		{
			name:     "contains subnet",
			route:    netlink.Route{Table: 123, Dst: mustParseCIDR("1.2.0.0/16")},
			rule:     netlink.Rule{Table: 123},
			expected: true,
		},
		{
			name:     "does not contain full subnet",
			route:    netlink.Route{Table: 123, Dst: mustParseCIDR("1.2.0.0/16")},
			rule:     netlink.Rule{Table: 123, Src: mustParseCIDR("1.2.3.0/24")},
			expected: false,
		},
		{
			name:     "suppress_prefixlength is applied",
			route:    netlink.Route{Table: 123},
			rule:     netlink.Rule{Table: 123, SuppressPrefixlen: 1},
			expected: false,
		},
		{
			name:     "suppress_prefixlength is applied inverted",
			route:    netlink.Route{Table: 123},
			rule:     netlink.Rule{Invert: true, Table: 123, SuppressPrefixlen: 1},
			expected: true,
		},
		{
			name:     "suppress_prefixlength does not suppress more specific routes",
			route:    netlink.Route{Table: 123, Dst: mustParseCIDR("1.2.0.0/16")},
			rule:     netlink.Rule{Table: 123, SuppressPrefixlen: 1},
			expected: true,
		},
		{
			name:     "non matching suppress_ifgroup is ignored",
			route:    netlink.Route{Table: 123},
			rule:     netlink.Rule{Table: 123, SuppressIfgroup: 321},
			ifgroup:  123,
			expected: true,
		},
		{
			name:     "matching suppress_ifgroup suppresses",
			route:    netlink.Route{Table: 123},
			rule:     netlink.Rule{Table: 123, SuppressIfgroup: 321},
			ifgroup:  321,
			expected: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(
				t,
				test.expected,
				ruleAppliesForRoute(rule(test.rule), test.route, test.ifgroup),
			)
		})
	}
}

func TestFilterRoutes(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name        string
		routes      []netlink.Route
		subnet      *net.IPNet
		ignoreTable int
		expected    []netlink.Route
	}{
		{
			name: "filter out ignored table",
			routes: []netlink.Route{
				{Dst: mustParseCIDR("192.168.1.0/24"), Table: 100},
				{Dst: mustParseCIDR("192.168.1.0/25"), Table: 200},
				{Dst: mustParseCIDR("192.168.1.128/25"), Table: 200},
			},
			subnet:      mustParseCIDR("192.168.1.0/24"),
			ignoreTable: 200,
			expected: []netlink.Route{
				{Dst: mustParseCIDR("192.168.1.0/24"), Table: 100},
			},
		},
		{
			name: "filter out routes not containing subnet",
			routes: []netlink.Route{
				{Dst: mustParseCIDR("192.168.0.0/16"), Table: 100},
				{Dst: mustParseCIDR("192.168.1.0/25"), Table: 100},
				{Dst: mustParseCIDR("192.168.2.0/24"), Table: 100},
			},
			subnet:      mustParseCIDR("192.168.1.0/24"),
			ignoreTable: 300,
			expected: []netlink.Route{
				{Dst: mustParseCIDR("192.168.0.0/16"), Table: 100},
			},
		},
		{
			name: "no filtering needed",
			routes: []netlink.Route{
				{Dst: mustParseCIDR("192.168.1.0/24"), Table: 100},
				{Dst: mustParseCIDR("192.168.1.0/23"), Table: 100},
			},
			subnet:      mustParseCIDR("192.168.1.0/24"),
			ignoreTable: 200,
			expected: []netlink.Route{
				{Dst: mustParseCIDR("192.168.1.0/24"), Table: 100},
				{Dst: mustParseCIDR("192.168.1.0/23"), Table: 100},
			},
		},
		{
			name: "all routes filtered out",
			routes: []netlink.Route{
				{Dst: mustParseCIDR("192.168.2.0/24"), Table: 100},
				{Dst: mustParseCIDR("192.168.1.0/25"), Table: 200},
			},
			subnet:      mustParseCIDR("192.168.1.0/24"),
			ignoreTable: 100,
			expected:    []netlink.Route{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := filterRoutes(test.routes, test.subnet, test.ignoreTable)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsSubnet(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name     string
		network  string
		subnet   string
		expected bool
	}{
		{"subnet within network", "192.168.1.0/24", "192.168.1.0/25", true},
		{"subnet within network (2nd half)", "192.168.1.0/24", "192.168.1.128/25", true},
		{"subnet outside network", "192.168.1.0/24", "192.168.2.0/24", false},
		{"subnet larger than network", "192.168.1.0/25", "192.168.1.0/24", false},
		{"subnet equal to network", "192.168.1.0/24", "192.168.1.0/24", true},
		{"default route covers subnet", "0.0.0.0/0", "192.168.1.0/24", true},
		{"default route is not within any subnet", "192.168.1.0/24", "0.0.0.0/0", false},
		{"IPv6 subnet within network", "2001:db8::/32", "2001:db8::/48", true},
		{"IPv6 subnet outside network", "2001:db8::/32", "2001:db9::/48", false},
		{"IPv6 subnet larger than network", "2001:db8::/48", "2001:db8::/32", false},
		{"nil network (default route)", "", "192.168.1.0/24", true},
		{"nil subnet", "192.168.1.0/24", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var network, subnet *net.IPNet
			if test.network != "" {
				_, network, _ = net.ParseCIDR(test.network)
			}
			if test.subnet != "" {
				_, subnet, _ = net.ParseCIDR(test.subnet)
			}
			assert.Equal(t, test.expected, isSubnet(network, subnet))
		})
	}
}

func rule(r netlink.Rule) netlink.Rule {
	newRule := netlink.NewRule()
	newRule.Table = r.Table
	newRule.Invert = r.Invert
	newRule.Src = r.Src
	if r.SuppressPrefixlen != 0 {
		newRule.SuppressPrefixlen = r.SuppressPrefixlen
	}
	if r.SuppressIfgroup != 0 {
		newRule.SuppressIfgroup = r.SuppressIfgroup
	}
	if r.Mark != 0 {
		newRule.Mark = r.Mark
	}
	return *newRule
}

// mustParseCIDR parses a CIDR string and returns a net.IPNet.
// It panics if there is an error during parsing.
func mustParseCIDR(cidr string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return ipNet
}
