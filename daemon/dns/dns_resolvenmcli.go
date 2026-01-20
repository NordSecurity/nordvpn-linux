package dns

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	execNmCli     = "nmcli"
	nmCliPrintTag = "[DNS][NMCLI]"
)

type NMCli struct {
	ignoreAutoDNS bool
}

// Set configures DNS nameservers for the specified network interface using nmcli tool.
//
// Parameters:
//   - iface: the network interface name to configure
//   - nameservers: a slice of DNS server addresses to set
//
// Returns an error if:
//   - iface is empty or nameservers slice is empty
//   - the nmcli command fails to execute
//   - the connection reload fails
func (nmcli *NMCli) Set(iface string, nameservers []string) error {
	//nmcli connection modify {} ipv4.dns {}
	if iface == "" || len(nameservers) == 0 {
		log.Println(internal.WarningPrefix, nmCliPrintTag, "Provided interface name or nameservers are empty")
		return fmt.Errorf("empty interface name or no nameservers provided")
	}
	args := []string{"connection", "modify", iface, "ipv4.dns"}
	args = append(args, nameservers...)
	args = append(args, "ipv4.ignore-auto-dns", "yes")

	if out, err := exec.Command(execNmCli, args...).CombinedOutput(); err != nil {
		log.Println(internal.WarningPrefix, nmCliPrintTag, "Setting DNS with nmcli failed:", strings.TrimSpace(string(out)))
		return fmt.Errorf("setting dns with nmcli failed: %w", err)
	}
	return nmcli.reloadConnection(iface)
}

// Unset removes the DNS configuration for the specified network interface using NetworkManager's nmcli tool.
// Parameters:
//   - iface: the network interface name to unset DNS for
//
// Returns an error if the nmcli command fails or if reloading the connection fails.
func (nmcli *NMCli) Unset(iface string) error {
	if iface == "" {
		return nil
	}
	args := []string{"connection", "modify", iface, "ipv4.dns", ""}
	if _, err := exec.Command(execNmCli, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("unsetting dns with nmcli failed: %w", err)
	}
	return nmcli.reloadConnection(iface)
}

// Name returns name of the DNS management method.
func (nmcli *NMCli) Name() string {
	return "nmcli"
}

// reloadConnection restarts the network connection for the specified interface using nmcli tool.
func (nmcli *NMCli) reloadConnection(iface string) error {
	disableInterfaceargs := []string{"connection", "down", iface}
	if out, err := exec.Command(execNmCli, disableInterfaceargs...).CombinedOutput(); err != nil {
		log.Println(internal.WarningPrefix, nmCliPrintTag, ":", strings.TrimSpace(string(out)))
		return fmt.Errorf("reload connection failed for DOWN request %w", err)
	}

	enableInterfaceargs := []string{"connection", "up", iface}
	if _, err := exec.Command(execNmCli, enableInterfaceargs...).CombinedOutput(); err != nil {
		return fmt.Errorf("reload connection failed for UP request %w", err)
	}
	return nil
}
