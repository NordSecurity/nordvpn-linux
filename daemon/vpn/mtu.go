package vpn

import (
	"errors"
	"log"
	"net"
	"os/exec"
	"strings"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/sys/unix"
)

var (
	errNoDefaultIpRoute          = errors.New("default gateway not found")
	errUnrecognizedIpRouteOutput = errors.New("unrecognized output of 'ip route show default'")
)

const defaultMTU = 1500

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

func retrieveAndCalculateMTU(headerSize int) int {
	c1 := exec.Command("ip", "route", "show", "default")
	out, err := c1.Output()

	if err != nil {
		log.Println(internal.ErrorPrefix, "ip route show default failed: ", err)
		out = nil
	}

	return calculateMTU(string(out), headerSize)
}

func calculateMTU(ipRouteOutput string, headerSize int) int {
	defaultGatewayMTU := func() (int, error) {
		defaultGatewayName, err := getDefaultIpRouteInterface(ipRouteOutput)

		if err != nil {
			return 0, err
		}

		defaultGateway, err := net.InterfaceByName(defaultGatewayName)
		if err != nil {
			return 0, err
		}

		// wireguard-quick does this
		mtu := defaultGateway.MTU - headerSize
		return mtu, nil
	}

	if ipRouteOutput != "" {
		mtu, err := defaultGatewayMTU()
		if err == nil {
			return mtu
		}

		log.Println(internal.WarningPrefix, "using default MTU, failed to get default gateway MTU:", err)
	}

	return defaultMTU - headerSize
}
