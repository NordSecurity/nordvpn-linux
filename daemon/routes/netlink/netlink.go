/*
Package netlink provides router implementation that uses netlink.
*/
package netlink

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/vishvananda/netlink"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

// Router uses netlink under the hood.
type Router struct {
	routes []routes.Route
	sync.Mutex
}

// Add appends route to a routes list via netlink if it does not exist yet.
func (r *Router) Add(route routes.Route) error {
	r.Lock()
	defer r.Unlock()
	if r.has(route) {
		return fmt.Errorf("route %+v already exists", route)
	}
	// check if such route does not exist in routing table
	netlinkRoute := toNetlinkRoute(route)
	err := netlink.RouteAdd(&netlinkRoute)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return routes.ErrRouteToOtherDestinationExists
		}
		return fmt.Errorf("adding route %+v to a routing table: %w", route, err)
	}

	r.routes = append(r.routes, route)
	return nil
}

// Flush deletes all previously added routes via netlink.
func (r *Router) Flush() error {
	r.Lock()
	defer r.Unlock()
	var errs []error
	for _, route := range r.routes {
		netlinkRoute := toNetlinkRoute(route)
		if err := netlink.RouteDel(&netlinkRoute); err != nil {
			errs = append(errs, fmt.Errorf("deleting route %+v: %w", route, err))
			continue
		}
	}
	r.routes = nil
	return errors.Join(errs...)
}

// has returns true if router contains a given route in its memory.
func (r *Router) has(route routes.Route) bool {
	return slices.ContainsFunc(r.routes, route.IsEqual)
}

// toNetlinkRoute converts from routes.Route to netlink.Route.
func toNetlinkRoute(route routes.Route) netlink.Route {
	isIPv6 := route.Gateway.Is6()
	bits := net.IPv4len * 8
	if isIPv6 {
		bits = net.IPv6len * 8
	}
	scope := netlink.SCOPE_UNIVERSE
	if !route.Gateway.IsValid() || route.Gateway.IsUnspecified() {
		scope = netlink.SCOPE_LINK
	}
	// Never insert routes to local table
	tableID := route.TableID
	if tableID == 0 {
		tableID = unix.RT_TABLE_MAIN
	}
	return netlink.Route{
		LinkIndex: route.Device.Index,
		Gw:        route.Gateway.AsSlice(),
		Dst: &net.IPNet{
			IP:   route.Subnet.Addr().AsSlice(),
			Mask: net.CIDRMask(route.Subnet.Bits(), bits),
		},
		Table: int(tableID),
		Scope: scope,
	}
}
