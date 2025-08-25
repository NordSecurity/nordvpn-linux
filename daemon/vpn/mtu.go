package vpn

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

const defaultMTU = 1500

// SetMTU for an interface.
func SetMTU(iface net.Interface, headerSize int) error {
	mtu := retrieveAndCalculateMTU(headerSize)

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

func getDefaultGatewayMTU() (int, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return 0, fmt.Errorf("listing routes: %w", err)
	}

	for _, route := range routes {
		if route.Dst == nil {
			link, err := netlink.LinkByIndex(route.LinkIndex)
			if err != nil {
				return 0, fmt.Errorf("getting link attributes: %w", err)
			}
			return link.Attrs().MTU, nil
		}
	}

	return 0, fmt.Errorf("failed to find default route")
}

func retrieveAndCalculateMTU(headerSize int) int {
	defaultRouteMTU, err := getDefaultGatewayMTU()
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to retrieve default interface, will use default value of: %w",
			defaultMTU, err)
		return defaultMTU
	}

	return defaultRouteMTU - headerSize
}
