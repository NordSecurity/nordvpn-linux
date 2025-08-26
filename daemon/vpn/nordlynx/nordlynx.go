// Package nordlynx provides nordlynx vpn technology.
package nordlynx

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	// InterfaceName for various NordLynx implementations
	InterfaceName       = "nordlynx"
	defaultPort         = 51820
	WireguardHeaderSize = 80
)

var (
	errNoKernelModule = errors.New("interface of type wireguard not supported")
)

var DefaultPrefix = netip.MustParsePrefix("10.5.0.2/16")

func upWGInterface(iface string) error {
	debug("ip", "link", "add", iface, "type", "wireguard")
	err := addDevice(iface)
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
func addDevice(device string) error {
	_, err := exec.Command("ip", "link", "add", device, "type", "wireguard").CombinedOutput()
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
