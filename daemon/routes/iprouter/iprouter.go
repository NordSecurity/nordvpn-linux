/*
Package iprouter provides Go API for interacting with ip route.
*/
package iprouter

import (
	"bytes"
	"errors"
	"fmt"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/network"
)

// Router uses `ip route` under the hood.
type Router struct {
	routes []routes.Route
	sync.Mutex
}

// Add calls ip route add command and appends route to a routes list if it does not exist yet
func (r *Router) Add(route routes.Route) error {
	r.Lock()
	defer r.Unlock()
	if r.has(route) {
		return fmt.Errorf("route %+v already exists", route)
	}
	// check if such route does not exist in routing table
	exists, err := existsInRoutingTable(route)
	if err != nil {
		return fmt.Errorf("checking if route exists: %w", err)
	}
	// If route already existed in the system, do not add it to the router so it would not be flushed
	if exists {
		return nil
	}

	ipArgs, err := getRouteArgs(route, "add")
	if err != nil {
		return fmt.Errorf("building 'ip route' command: %w", err)
	}

	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command("ip", ipArgs...).CombinedOutput()
	if strings.HasSuffix(strings.TrimSpace(string(out)), "File exists") {
		return routes.ErrRouteToOtherDestinationExists
	}
	if err != nil {
		return fmt.Errorf("executing 'ip %s' command: %w: %s %v", strings.Join(ipArgs, " "), err, string(out), route)
	}
	r.routes = append(r.routes, route)
	return nil
}

// Flush calls ip route delete command for all existing routes
func (r *Router) Flush() error {
	r.Lock()
	defer r.Unlock()
	for _, route := range r.routes {
		if !r.has(route) {
			return fmt.Errorf("route %+v does not exist", route)
		}

		ipArgs, err := getRouteArgs(route, "delete")
		if err != nil {
			return fmt.Errorf("building 'ip route' command: %w", err)
		}

		// #nosec G204 -- input is properly sanitized
		out, err := exec.Command("ip", ipArgs...).CombinedOutput()
		if err != nil && !strings.HasSuffix(strings.TrimSpace(string(out)), "No such process") {
			return fmt.Errorf("executing 'ip route delete %s' command: %w: %s", route.Subnet.String(), err, string(out))
		}
	}
	r.routes = nil
	return nil
}

func (r *Router) has(route routes.Route) bool {
	for _, ro := range r.routes {
		if route.IsEqual(ro) {
			return true
		}
	}
	return false
}

func existsInRoutingTable(route routes.Route) (bool, error) {
	version := "-4"
	if route.Subnet.Addr().Is6() {
		version = "-6"
	}

	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command("ip", version, "route").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("executing 'ip %s route': %w: %s", version, err, string(out))
	}
	return existsInRoutingTableOutput(out, route), nil
}

func existsInRoutingTableOutput(out []byte, route routes.Route) bool {
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		networkStr := network.ToRouteString(route.Subnet)
		if bytes.HasPrefix(line, []byte(networkStr)) {
			if bytes.Contains(line, []byte(route.Gateway.String())) {
				return true
			} else if route.Device.Name != "" && bytes.Contains(line, []byte(route.Device.Name)) {
				return true
			}
		}
	}
	return false
}

func getRouteArgs(route routes.Route, operation string) ([]string, error) {
	if route.Device.Name == "" {
		return nil, errors.New("dev is empty")
	}
	version := "-4"
	if route.Subnet.Addr().Is6() {
		version = "-6"
	}

	args := []string{
		version,
		"route",
		operation,
	}

	if route.TableID != 0 {
		args = append(
			args,
			"table",
			strconv.Itoa(int(route.TableID)),
		)
	}

	if route.Subnet.Addr() != (netip.Addr{}) && route.Gateway != (netip.Addr{}) {
		return append(
			args,
			route.Subnet.String(),
			"via",
			route.Gateway.String(),
			"dev",
			route.Device.Name,
		), nil
	}

	if route.Gateway != (netip.Addr{}) {
		return append(
			args,
			route.Gateway.String(),
			"dev",
			route.Device.Name,
		), nil
	}

	if route.Subnet.Addr() != (netip.Addr{}) {
		return append(
			args,
			route.Subnet.String(),
			"dev",
			route.Device.Name,
		), nil
	}

	return append(
		args,
		"dev",
		route.Device.Name,
	), nil
}
