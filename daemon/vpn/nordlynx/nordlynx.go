// Package nordlynx provides nordlynx vpn technology.
package nordlynx

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/internal"

	"golang.org/x/sys/unix"
)

const (
	// InterfaceName for various NordLynx implementations
	InterfaceName = "nordlynx"
	defaultPort   = 51820
	defaultMTU    = 1500
)

var (
	errNoKernelModule            = errors.New("interface of type wireguard not supported")
	errNoDefaultIpRoute          = errors.New("default gateway not found")
	errUnrecognizedIpRouteOutput = errors.New("unrecognized output of 'ip route show default'")
)

// nordlynx client ipv6 address interface id (second portion of the address)
// nordlynx requires interface id to end with 2
// firewall rules depend on it
// agree with infra before changing it
func interfaceID() [8]byte {
	return [8]byte{0x0, 0x0, 0x0, 0x11, 0x0, 0x5, 0x0, 0x2}
}

// getDefaultIpRouteInterface takes output of the `ip route show default` command and returns the
// interface/device name. If there are multiple default routes in the output, first one will be returned
func getDefaultIpRouteInterface(ipRouteOutput string) (string, error) {
	outputRows := strings.Split(ipRouteOutput, "\n")

	if len(outputRows) < 1 || outputRows[0] == "" {
		return "", errNoDefaultIpRoute
	}

	outputColumns := strings.Split(strings.Trim(outputRows[0], "\n"), " ")

	if len(outputColumns) < 5 {
		log.Printf("unexpected output of 'ip route show default': %s, dev value not found", outputRows[0])
		return "", errUnrecognizedIpRouteOutput
	}

	return outputColumns[4], nil
}

// SetMTU for an interface.
func SetMTU(iface net.Interface) error {
	var err error

	c1 := exec.Command("ip", "route", "show", "default")
	out, err := c1.Output()

	if err != nil {
		return fmt.Errorf("ip route show default failed: %s", err)
	}

	defaultGatewayName, err := getDefaultIpRouteInterface(string(out))

	if err != nil {
		return err
	}

	defaultGateway, err := net.InterfaceByName(defaultGatewayName)
	if err != nil {
		return err
	}

	// wireguard-quick does this
	mtu := defaultGateway.MTU - 80

	fd, err := unix.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_IP)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	req, err := unix.NewIfreq(iface.Name)
	if err != nil {
		return err
	}
	req.SetUint32(uint32(mtu))

	return unix.IoctlIfreq(fd, unix.SIOCSIFMTU, req)
}

func upWGInterface(iface string) error {
	debug("ip", "link", "add", iface, "type", "wireguard")
	err := addDevice(iface, "wireguard")
	// there are only 2 cases when this can fail:
	// 1. Either kernel module is not found or the kernel was
	// recently updated, but the system is yet to be rebooted.
	// 2. wg command not found in path. (valid while we still rely on wg-tools)
	if err != nil {
		if internal.IsCommandAvailable("wg") {
			return err
		}
		return errNoKernelModule
	}
	return nil
}

func deleteInterface(iface net.Interface) error {
	debug("ip", "link", "delete", iface.Name)
	out, err := removeDevice(iface.Name)
	if err != nil {
		return errors.New(strings.Trim(string(out), "\n"))
	}
	return nil
}

// addDevice creates a new device with a given
// name and specified device type.
func addDevice(device, devType string) error {
	_, err := exec.Command("ip", "link", "add", device, "type", devType).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add device %w", err)
	}

	return nil
}

// removeDevice deletes the specified device.
func removeDevice(device string) ([]byte, error) {
	out, err := exec.Command("ip", "link", "delete", device).CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("failed to remove device %w", err)
	}
	return out, nil
}

func debug(data ...string) {
	log.Println("[nordlynx]", strings.Join(data, " "))
}
