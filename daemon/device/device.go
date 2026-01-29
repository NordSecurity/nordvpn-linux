// Package device provides utilities for querying device information.
package device

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type ListFunc func() ([]net.Interface, error)

var sysDepsImpl SystemDeps = realSystemDeps{}

type SystemDeps interface {
	// netlink
	RouteList(link netlink.Link, family int) ([]netlink.Route, error)

	// net
	InterfaceByIndex(index int) (*net.Interface, error)
	InterfaceByName(name string) (*net.Interface, error)
	Interfaces() ([]net.Interface, error)

	// os
	ReadDir(name string) ([]os.DirEntry, error)
}

// realSystemDeps is the production implementation backed by real OS calls.
type realSystemDeps struct{}

func (realSystemDeps) RouteList(link netlink.Link, family int) ([]netlink.Route, error) {
	return netlink.RouteList(link, family)
}

func (realSystemDeps) InterfaceByIndex(index int) (*net.Interface, error) {
	return net.InterfaceByIndex(index)
}

func (realSystemDeps) InterfaceByName(name string) (*net.Interface, error) {
	return net.InterfaceByName(name)
}

func (realSystemDeps) Interfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

func (realSystemDeps) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func listVirtual() ([]net.Interface, error) {
	files, err := sysDepsImpl.ReadDir("/sys/devices/virtual/net/")
	if err != nil {
		return nil, fmt.Errorf("listing files in network interfaces dir: %w", err)
	}

	var devices []net.Interface
	for _, file := range files {
		dev, err := sysDepsImpl.InterfaceByName(file.Name())
		if err != nil {
			return nil, fmt.Errorf("retrieving network interface by name: %w", err)
		}

		devices = append(devices, *dev)
	}

	return devices, nil
}

// OutsideCapableTrafficInterfaces returns a list of interfaces that can send traffic outside.
// The list includes
// * all physical interfaces
// * plus interfaces that have default route
// * plus interfaces that have gateway, e.g. 1.1.1.1 via 192.168.0.1 dev br0
func OutsideCapableTrafficInterfaces() ([]net.Interface, error) {
	interfaces, err := sysDepsImpl.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("retrieving system network interfaces: %w", err)
	}
	vInterfaces, err := listVirtual()
	if err != nil {
		return nil, fmt.Errorf("retrieving virtual interfaces: %w", err)
	}

	var devices []net.Interface

	for _, iface := range interfaces {
		if !ifaceListContains(vInterfaces, iface) {
			devices = append(devices, iface)
		}
	}

	// add interfaces that are capable to route traffic to outside
	// for example 1.1.1.1 via 192.168.0.1 dev br0
	routeList, err := sysDepsImpl.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the routes %w", err)
	}
	for _, r := range routeList {
		if !isOutsideCapable(r) {
			continue
		}

		if iface, err := sysDepsImpl.InterfaceByIndex(r.LinkIndex); err == nil && iface != nil {
			if !ifaceListContains(devices, *iface) {
				devices = append(devices, *iface)
			}
		} else {
			log.Println(internal.WarningPrefix, "not found interface with index", r.LinkIndex, err)
		}
	}
	return devices, nil
}

// OutsideCapableTrafficIfNames is a helper function that returns same as OutsideCapableTrafficInterfaces
// but just the interfaces names
func OutsideCapableTrafficIfNames(ignore mapset.Set[string]) mapset.Set[string] {
	result := mapset.NewSet[string]()
	ifaces, err := OutsideCapableTrafficInterfaces()
	if err != nil {
		log.Println(internal.WarningPrefix, "netlink monitoring failed to get interfaces", err)
		return result
	}

	for _, iface := range ifaces {
		if ignore == nil || !ignore.Contains(iface.Name) {
			result.Add(iface.Name)
		}
	}

	return result
}

func ifaceListContains(list []net.Interface, device net.Interface) bool {
	for _, iface := range list {
		if iface.Name == device.Name {
			return true
		}
	}
	return false
}

// DefaultGateway returns network interface used as default gateway.
//
// Linux generally has only a single default gateway. Although it can
// have more than one default gateway by using routing tables, only one
// is allowed per routing table.
func DefaultGateway() (net.Interface, error) {
	cmd := exec.Command("ip", "-4", "route", "list", "default") // local table
	out, err := cmd.CombinedOutput()
	if err != nil {
		return net.Interface{}, fmt.Errorf("getting network interface used by default route: %w", err)
	}

	if string(out) == "" {
		return net.Interface{}, fmt.Errorf("default gateway does not exist")
	}

	var name string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name, err = interfaceNameFromIPRoute(line)
		if err != nil {
			return net.Interface{}, fmt.Errorf("looking up the name of default gateway: %w", err)
		}
		//nolint:staticcheck
		break
	}

	device, err := net.InterfaceByName(name)
	if err != nil {
		return net.Interface{}, fmt.Errorf("default gateway retrieving network interface by name: %w", err)
	}
	return *device, nil
}

func interfaceNameFromIPRoute(line string) (string, error) {
	words := strings.Split(line, " ")
	for i, word := range words {
		if word == "dev" { // next word is the name of an interface
			return words[i+1], nil
		}
	}

	return "", fmt.Errorf("malformed input")
}

func InterfacesAreEqual(a net.Interface, b net.Interface) bool {
	return a.Index == b.Index &&
		a.MTU == b.MTU &&
		a.Name == b.Name &&
		a.HardwareAddr.String() == b.HardwareAddr.String() &&
		a.Flags == b.Flags
}

// InterfacesWithDefaultRoute returns all the interfaces that have a default route, excluding the ones from ignoreSet
func InterfacesWithDefaultRoute(ignoreSet mapset.Set[string]) map[string]net.Interface {
	// get interface list from default routes
	interfacesList := make(map[string]net.Interface)

	routeList, err := sysDepsImpl.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get system routes", err)
		return interfacesList
	}
	for _, r := range routeList {
		if !isDefaultRoute(r) {
			continue
		}

		if iface, err := sysDepsImpl.InterfaceByIndex(r.LinkIndex); err == nil && iface != nil {
			if ignoreSet == nil || !ignoreSet.Contains(iface.Name) {
				interfacesList[iface.Name] = *iface
			}
		} else {
			log.Println(internal.WarningPrefix, "default route, not found interface with index", r.LinkIndex, err)
		}
	}

	return interfacesList
}

// isDefaultRoute checks if a route is a default route
func isDefaultRoute(r netlink.Route) bool {
	if r.Dst == nil {
		return true
	}

	ones, bits := r.Dst.Mask.Size()
	if ones != 0 || bits != 32 { // must be /0
		return false
	}

	return true
}

// isOutsideCapable detects if a route is capable to send traffic outside
func isOutsideCapable(r netlink.Route) bool {
	// Ignore non-forwarding routes
	if r.Type == unix.RTN_BLACKHOLE ||
		r.Type == unix.RTN_UNREACHABLE ||
		r.Type == unix.RTN_PROHIBIT {
		return false
	}

	// Default route (even without gateway)
	if isDefaultRoute(r) {
		return true
	}

	// Any route with a real gateway
	if r.Gw != nil && !r.Gw.IsUnspecified() {
		return true
	}

	return false
}
