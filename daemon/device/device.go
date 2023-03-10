// Package device provides utilities for querying device information.
package device

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

type ListFunc func() ([]net.Interface, error)

func listVirtual() ([]net.Interface, error) {
	files, err := os.ReadDir("/sys/devices/virtual/net/")
	if err != nil {
		return nil, fmt.Errorf("listing files in network interfaces dir: %w", err)
	}

	var devices []net.Interface
	for _, file := range files {
		dev, err := net.InterfaceByName(file.Name())
		if err != nil {
			return nil, fmt.Errorf("retrieving network interface by name: %w", err)
		}

		devices = append(devices, *dev)
	}

	return devices, nil
}

// ListPhysical network interfaces found on the system.
//
// All Linux systems have physical interfaces with one exception - containers.
// When system is properly virtualized, guest does not know that it is virtual
// so even though those interfaces are mapped to virtual interfaces on the host,
// guest does not know this, but containers know. This is because the kernel is
// shared between the host and the guest.
//
// If the system has only virtual interfaces, return a virtual interface which is used as
// default gateway.
func ListPhysical() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("retrieving system network interfaces: %w", err)
	}
	vInterfaces, err := listVirtual()
	if err != nil {
		return nil, fmt.Errorf("retrieving virtual interfaces: %w", err)
	}

	if len(interfaces) == len(vInterfaces) {
		gateway4, err4 := DefaultGateway(false)
		gateway6, err6 := DefaultGateway(true)

		if err4 == nil && err6 == nil && !interfacesAreEqual(gateway4, gateway6) {
			return []net.Interface{gateway4, gateway6}, nil
		}
		if err4 == nil {
			return []net.Interface{gateway4}, nil
		}
		if err6 == nil {
			return []net.Interface{gateway6}, nil
		}

		if err4 != nil {
			return nil, err4
		}
		if err6 != nil {
			return nil, err6
		}
		return nil, errors.New("unable to retrieve default gateway")
	}

	var devices []net.Interface
	for _, iface := range interfaces {
		if !ifaceListContains(vInterfaces, iface) {
			devices = append(devices, iface)
		}
	}
	return devices, nil
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
func DefaultGateway(ipv6 bool) (net.Interface, error) {
	version := "-4"
	if ipv6 {
		version = "-6"
	}
	cmd := exec.Command("ip", version, "route", "list", "default") // local table
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
		break
	}

	device, err := net.InterfaceByName(name)
	if err != nil {
		return net.Interface{}, fmt.Errorf("retrieving network interface by name: %w", err)
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

func interfacesAreEqual(a net.Interface, b net.Interface) bool {
	return a.Index == b.Index &&
		a.MTU == b.MTU &&
		a.Name == b.Name &&
		a.HardwareAddr.String() == b.HardwareAddr.String() &&
		a.Flags == b.Flags
}
