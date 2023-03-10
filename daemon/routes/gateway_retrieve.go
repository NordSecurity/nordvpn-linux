package routes

import (
	"bytes"
	"fmt"
	"net"
	"net/netip"
	"os/exec"
	"regexp"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
)

// GatewayRetriever is responsible for retrieving default gateway in current system
type GatewayRetriever interface {
	// Default retrieves a default gateway
	Default(ipv6 bool) (netip.Addr, net.Interface, error)
}

// IPGatewayRetriever retrieves default gateway using ip command
type IPGatewayRetriever struct{}

// Default retrieves a default gateway using ip route command
func (r IPGatewayRetriever) Default(ipv6 bool) (netip.Addr, net.Interface, error) {
	version := "-4"
	if ipv6 {
		version = "-6"
	}

	out, err := exec.Command("ip", version, "route").CombinedOutput()
	if err != nil {
		return netip.Addr{}, net.Interface{}, fmt.Errorf("executing 'ip %s route' command: %w: %s", version, err, string(out))
	}

	dev, err := device.DefaultGateway(ipv6)
	if err != nil {
		return netip.Addr{}, net.Interface{}, err
	}

	ip, err := netip.ParseAddr(string(grepDefaultGatewayIPFromOutput(out, []string{dev.Name})))
	if err != nil {
		return netip.Addr{}, net.Interface{}, fmt.Errorf("default gateway was not found: %w", err)
	}
	return ip, dev, nil
}

// grepDefaultGatewayIPFromOutput finds and returns default gateway from ip route output
// if gateway was not found, returns an empty slice
func grepDefaultGatewayIPFromOutput(output []byte, devNames []string) []byte {
	reg := regexp.MustCompile(`via ([a-fA-F0-9:.]+?) dev`)
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		// if no dev names were provided, choose any default gateway
		if (strContainsAny(string(line), devNames)) && bytes.HasPrefix(line, []byte("default")) {
			matches := reg.FindSubmatch(line)
			if len(matches) > 1 {
				if net.ParseIP(string(matches[1])) != nil {
					return matches[1]
				}
			}
			return nil
		}
	}
	return nil
}

func strContainsAny(str string, list []string) bool {
	for _, substr := range list {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}
