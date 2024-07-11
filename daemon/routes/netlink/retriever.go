package netlink

import (
	"fmt"
	"net"
	"net/netip"
	"slices"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/vishvananda/netlink"
)

// Retriever is a routes.GatewayRetriever implementation that is using netlink
type Retriever struct{}

// Retrieve a gateway to a given address while ignoring the given routing table. The mechanism for
// determining a gateway:
//  1. All routes are queried in the system;
//  2. Routes are filtered so they contain only those routes which contain the subnet;
//  3. Filtered routes are sorted in priority order as it would be on route selection for a packet;
//  4. All rules are listed and ordered by priority by default;
//  5. Routes are re-ordered by the ip rule that applies to the route. Routes for which same route
//     can be applied, maintain the same order as defined in 3;
//  6. First route in the list is chosen as the best match and used to determine a gateway.
func (Retriever) Retrieve(prefix netip.Prefix, ignoreTable uint) (netip.Addr, net.Interface, error) {
	routeList, err := listRoutesForSubnet(prefixToIPNet(prefix), int(ignoreTable))
	if err != nil {
		return netip.Addr{},
			net.Interface{},
			fmt.Errorf("listing routes for subnet: %w", err)
	}
	if len(routeList) == 0 {
		return netip.Addr{}, net.Interface{}, routes.ErrNotFound
	}
	var route *netlink.Route
	for _, rt := range routeList {
		if rt.Gw != nil {
			route = &rt
			break
		}
	}
	if route == nil {
		return netip.Addr{},
			net.Interface{},
			fmt.Errorf("retrieving route with gateway: %w", err)
	}
	iface, err := net.InterfaceByIndex(route.LinkIndex)
	if err != nil || iface == nil {
		return netip.Addr{},
			net.Interface{},
			fmt.Errorf("retrieving interface by index %d: %w", route.LinkIndex, err)
	}

	// If not ok, Gw is likely not set
	ip, ok := netip.AddrFromSlice(route.Gw)
	if !ok {
		return netip.Addr{}, net.Interface{}, fmt.Errorf("failed retrieving gateway ip")
	}
	return ip, *iface, nil
}

// listRoutesForSubnet implements a route listing and sorting mechanism for the Retriever.
func listRoutesForSubnet(subnet *net.IPNet, ignoreTable int) ([]netlink.Route, error) {
	family := toNetlinkFamily(subnet.IP)
	routes, err := netlink.RouteListFiltered(family, &netlink.Route{}, netlink.RT_FILTER_TABLE)
	if err != nil {
		return nil, fmt.Errorf("listing routes: %w", err)
	}

	routes = filterRoutes(routes, subnet, ignoreTable)

	// Best route already found or it does not exist
	if len(routes) <= 1 {
		return routes, nil
	}

	// Sort routes as how netlink would do
	slices.SortFunc(routes, routeCmp)

	rules, err := netlink.RuleList(family)
	if err != nil {
		return nil, fmt.Errorf("listing rules: %w", err)
	}

	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("listing links: %w", err)
	}

	return orderRoutesByRules(rules, routes, links), nil
}

// orderRoutesByRules groups routes by the rules that apply to the route and orders those groups by
// the rules the order of rules. Routes in the same group maintain the same order.
func orderRoutesByRules(
	rules []netlink.Rule,
	routes []netlink.Route,
	links []netlink.Link,
) []netlink.Route {
	linksGroupMap := map[int]uint32{}
	for _, link := range links {
		if attrs := link.Attrs(); attrs != nil {
			linksGroupMap[attrs.Index] = attrs.Group
		}
	}
	used := make([]bool, len(routes))
	var out []netlink.Route
	for _, rule := range rules {
		for i, route := range routes {
			ifgroup := linksGroupMap[route.LinkIndex]
			if !used[i] && ruleAppliesForRoute(rule, route, ifgroup) {
				used[i] = true
				out = append(out, route)
			}
		}
	}
	return out
}

// filterRoutes that don't belong to the ignored table and contain the given subnet.
func filterRoutes(routes []netlink.Route, subnet *net.IPNet, ignoreTable int) []netlink.Route {
	return slices.DeleteFunc(routes, func(r netlink.Route) bool {
		return r.Table == ignoreTable || !isSubnet(r.Dst, subnet)
	})
}

// ruleAppliesForRoute determines if rule applies to a given route.
func ruleAppliesForRoute(rule netlink.Rule, route netlink.Route, ifgroup uint32) bool {
	routeDstPrefixLen := 0
	if route.Dst != nil {
		routeDstPrefixLen, _ = route.Dst.Mask.Size()
	}
	// Cannot make any assumptions about fwmarks as route has no information about them
	return rule.Invert != (rule.Mark < 0 &&
		rule.Table == route.Table &&
		isSubnet(rule.Src, route.Dst) &&
		(rule.SuppressPrefixlen < 0 || rule.SuppressPrefixlen < routeDstPrefixLen) &&
		(rule.SuppressIfgroup < 0 || rule.SuppressIfgroup != int(ifgroup)))
}

// routeCmp compares which of the routes is more specific.
func routeCmp(a netlink.Route, b netlink.Route) int {
	aPrefixLen := 0
	bPrefixLen := 0
	if a.Dst != nil {
		aPrefixLen, _ = a.Dst.Mask.Size()
	}
	if b.Dst != nil {
		bPrefixLen, _ = b.Dst.Mask.Size()
	}
	if aPrefixLen > bPrefixLen {
		return -1
	} else if aPrefixLen < bPrefixLen {
		return 1
	} else if a.Priority < b.Priority {
		return -1
	} else if a.Priority > b.Priority {
		return 1
	}
	return 0
}

// isSubnet returns true if network contains whole range of subnet. Otherwise, returns false.
func isSubnet(network, subnet *net.IPNet) bool {
	// nil is treated as an IP with 0 prefix
	if network == nil {
		return true
	}
	if subnet == nil {
		return false
	}
	nMaskSize, _ := network.Mask.Size()
	sMaskSize, _ := subnet.Mask.Size()
	return network.Contains(subnet.IP) && nMaskSize <= sMaskSize
}

func toNetlinkFamily(ip net.IP) int {
	if len(ip) == net.IPv6len {
		return netlink.FAMILY_V6
	}
	return netlink.FAMILY_V4
}
