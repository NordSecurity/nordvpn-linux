package dns

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	execNmCli               = "nmcli"
	wirelessType            = "wireless"
	ethernetType            = "ethernet"
	gsmType                 = "gsm"
	cdmaType                = "cdma"
	nmCliPrintTag           = "[DNS][NMCLI]"
	nmCliConKey             = "con"
	nmCliIPv4DNSKey         = "ipv4.dns"
	nmCliIPIgnoreAutoDnsKey = "ipv4.ignore-auto-dns"
)

type NMCli struct {
	cmdExecutor func(name string, arg ...string) ([]byte, error)
}

func newNMCli() *NMCli {
	return &NMCli{
		cmdExecutor: func(name string, arg ...string) ([]byte, error) {
			// #nosec G204: input is properly validated
			return exec.Command(name, arg...).CombinedOutput()
		},
	}
}

// Set configures DNS nameservers for the specified network interface using nmcli tool.
//
// Parameters:
//   - iface: unused
//   - nameservers: a set of DNS addresses to be used for the configuration
//
// Returns an error if:
//   - the nmcli command fails to fetch connections related to physical interfaces
//   - the nmcli command fails to execute
//   - the connection reload fails
func (nmcli *NMCli) Set(iface string, nameservers []string) error {
	connections, err := nmcli.getConnectionFromPhysicalInterfaces()
	if err != nil {
		log.Println(internal.WarningPrefix, nmCliPrintTag, "Failed to get active connections upon SetDNS", err)
		return fmt.Errorf("%w", err)
	}
	for _, con := range connections {
		args := []string{nmCliConKey, "modify", con, nmCliIPv4DNSKey}
		args = append(args, strings.Join(nameservers, ","))
		args = append(args, nmCliIPIgnoreAutoDnsKey, "yes")

		if out, err := nmcli.cmdExecutor(execNmCli, args...); err != nil {
			log.Println(internal.WarningPrefix, nmCliPrintTag, "Setting DNS with nmcli failed:", strings.TrimSpace(string(out)))
			return fmt.Errorf("setting dns with nmcli failed: %w", err)
		}

		if err := nmcli.reloadConnection(con); err != nil {
			return fmt.Errorf("failure %w", err)
		}
	}
	return nil
}

// Unset removes the DNS configuration for the specified network interface using NetworkManager's nmcli tool.
// Parameters:
//   - iface: unused
//
// Returns an error if the nmcli command fails or if reloading the connection fails.
func (nmcli *NMCli) Unset(iface string) error {
	connections, err := nmcli.getConnectionFromPhysicalInterfaces()
	if err != nil {
		log.Println(internal.WarningPrefix, nmCliPrintTag, "Failed to get active connections upon UnsetDNS", err)
		return fmt.Errorf("%w", err)
	}
	for _, con := range connections {
		args := []string{nmCliConKey, "modify", con, nmCliIPv4DNSKey, ""}
		args = append(args, nmCliIPIgnoreAutoDnsKey, "no")

		if out, err := nmcli.cmdExecutor(execNmCli, args...); err != nil {
			log.Println(internal.WarningPrefix, nmCliPrintTag, "Setting DNS with nmcli failed:", strings.TrimSpace(string(out)))
			return fmt.Errorf("setting dns with nmcli failed: %w", err)
		}
		if err := nmcli.reloadConnection(con); err != nil {
			return fmt.Errorf("failure %w", err)
		}
	}
	return nil
}

// Name returns name of the DNS management method.
func (nmcli *NMCli) Name() string {
	return "nmcli"
}

// getConnectionFromPhysicalInterfaces retrieves a list of active physical network connection names
// from NetworkManager using nmcli. It filters connections by type, including only wireless,
// ethernet, GSM, and CDMA connections (eg. physical ones).
// Returns a slice of connection names and an error if the nmcli command fails.
// Returns an empty slice upon any malformed output from the nmcli command.
func (nmcli *NMCli) getConnectionFromPhysicalInterfaces() ([]string, error) {
	cmd, err := nmcli.cmdExecutor(execNmCli, "-t", "-f", "NAME,TYPE", "con", "show", "--active")
	if err != nil {
		return []string{}, fmt.Errorf("Failed to fetch active devices: %w", err)
	}
	var conns = []string{}
	lines := strings.SplitSeq(string(cmd), "\n")
	for line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		// to correctly handle connection name with a colon in the name
		// combine it back to have always two fields: name, and a type
		fields := []string{
			strings.Join(parts[:len(parts)-1], ":"),
			parts[len(parts)-1],
		}
		if strings.Contains(fields[1], wirelessType) ||
			strings.Contains(fields[1], ethernetType) ||
			strings.Contains(fields[1], gsmType) ||
			strings.Contains(fields[1], cdmaType) {
			conns = append(conns, strings.TrimSpace(fields[0]))
		}
	}
	return conns, nil
}

// reloadConnection restarts the network connection for the specified interface using nmcli tool.
func (nmcli *NMCli) reloadConnection(iface string) error {
	reloadArgs := []string{nmCliConKey, "reload", iface}
	if out, err := nmcli.cmdExecutor(execNmCli, reloadArgs...); err != nil {
		log.Println(internal.WarningPrefix, nmCliPrintTag, ":", strings.TrimSpace(string(out)))
		return fmt.Errorf("reload connection failed: %w", err)
	}

	upArgs := []string{nmCliConKey, "up", iface}
	if _, err := nmcli.cmdExecutor(execNmCli, upArgs...); err != nil {
		//at this stage we can disregard the error, as the DNS configuration should be applied even if the connection is not reloaded properly. Log it for debugging purposes.
		log.Println(internal.WarningPrefix, nmCliPrintTag, "Setting ", iface, " UP failed with:", err)
	}
	return nil
}
