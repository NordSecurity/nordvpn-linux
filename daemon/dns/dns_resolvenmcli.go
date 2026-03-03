package dns

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	bridgeType   = "bridge"
	cdmaType     = "cdma"
	ethernetType = "ethernet"
	gsmType      = "gsm"
	wirelessType = "wireless"

	nmCliConKey             = "con"
	nmCliExecutable         = "nmcli"
	nmCliIPv4DNSKey         = "ipv4.dns"
	nmCliIPIgnoreAutoDnsKey = "ipv4.ignore-auto-dns"
	nmCliSlaveBridgeType    = "connection.slave-type"
)

// connectionState holds the DNS configuration state for a connection
type connectionState struct {
	name          string
	ipv4DNS       string
	ignoreAutoDNS string
	isModified    bool
}

type connectionInfo struct {
	name     string
	isActive bool
}

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
		log.Println(internal.WarningPrefix, dnsPrefix, "Failed to get active connections upon SetDNS", err)
		return fmt.Errorf("failed to get active connection upon SetDNS: %w", err)
	}

	originalStates := make(map[string]*connectionState)
	for _, con := range connections {
		state, err := nmcli.getConnectionState(con.name)
		if err != nil {
			log.Println(internal.WarningPrefix, dnsPrefix, "Failed to get state for connection", con, ":", err)
			return fmt.Errorf("failed to get connection state: %w", err)
		}
		originalStates[con.name] = state
	}

	for _, con := range connections {
		args := []string{nmCliConKey, "modify", con.name, nmCliIPv4DNSKey}
		args = append(args, strings.Join(nameservers, ","))
		args = append(args, nmCliIPIgnoreAutoDnsKey, "yes")

		if _, err := nmcli.cmdExecutor(nmCliExecutable, args...); err != nil {
			log.Println(internal.WarningPrefix, dnsPrefix, "Failed to modify connection", con, ":", err)
			nmcli.rollback(originalStates)
			return fmt.Errorf("setting dns with nmcli failed: %w", err)
		}
		originalStates[con.name].isModified = true

		if con.isActive {
			if err := nmcli.reloadConnection(con.name); err != nil {
				log.Println(internal.WarningPrefix, dnsPrefix, "Failed to reload connection", con, ":", err)
				nmcli.rollback(originalStates)
				return fmt.Errorf("failed to reload connection upon SetDNS: %w", err)
			}
		}
	}
	return nil
}

// Unset removes the DNS configuration for the specified network interface using NetworkManager's nmcli tool.
// Parameters:
//   - iface: unused
//
// Returns an error if the nmcli command fails or if reloading the connection fails.
func (nmcli *NMCli) Unset(_ string) error {
	connections, err := nmcli.getConnectionFromPhysicalInterfaces()
	if err != nil {
		log.Println(internal.WarningPrefix, dnsPrefix, "Failed to get active connections upon UnsetDNS", err)
		return fmt.Errorf("failed to get active connection upon UnsetDNS: %w", err)
	}
	for _, con := range connections {
		args := []string{nmCliConKey, "modify", con.name, nmCliIPv4DNSKey, ""}
		args = append(args, nmCliIPIgnoreAutoDnsKey, "no")

		if _, err := nmcli.cmdExecutor(nmCliExecutable, args...); err != nil {
			return fmt.Errorf("setting dns with nmcli failed: %w", err)
		}
		if con.isActive {
			if err := nmcli.reloadConnection(con.name); err != nil {
				return fmt.Errorf("failed to reload connection upon UnsetDNS: %w", err)
			}
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
// Any malformed output  is disregarded.
//
// Returns a slice of connection names and an error if the nmcli command fails.
func (nmcli *NMCli) getConnectionFromPhysicalInterfaces() ([]connectionInfo, error) {
	activeConnectionsList, err := nmcli.cmdExecutor(nmCliExecutable, "-t", "-f", "NAME,TYPE", "con", "show", "--active")
	if err != nil {
		return []connectionInfo{}, fmt.Errorf("Failed to fetch active connections: %w", err)
	}

	allConnectionsList, err := nmcli.cmdExecutor(nmCliExecutable, "-t", "-f", "NAME,TYPE", "con", "show")
	if err != nil {
		return []connectionInfo{}, fmt.Errorf("Failed to fetch all connections: %w", err)
	}
	activeConns := nmcli.parsePhysicalConnections(string(activeConnectionsList))
	activeSet := make(map[string]bool)
	for _, conn := range activeConns {
		activeSet[conn] = true
	}
	allConns := nmcli.parsePhysicalConnections(string(allConnectionsList))

	var connections []connectionInfo
	for _, conn := range allConns {
		connections = append(connections, connectionInfo{name: conn, isActive: activeSet[conn]})
	}
	return connections, nil
}

// parsePhysicalConnections parses nmcli output and filters physical network connections
func (nmcli *NMCli) parsePhysicalConnections(connectionsOutput string) []string {
	var conns = []string{}
	lines := strings.SplitSeq(connectionsOutput, "\n")
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
			(strings.Contains(fields[1], ethernetType) && !nmcli.isSlaveBridgeConnection(fields[0])) ||
			strings.Contains(fields[1], bridgeType) ||
			strings.Contains(fields[1], gsmType) ||
			strings.Contains(fields[1], cdmaType) {
			// for tersed output (eg. the -t flag) nmcli provides escaped colons in the connection names
			// thus it is essential to unescape them here
			connName := strings.ReplaceAll(fields[0], "\\:", ":")
			conns = append(conns, strings.TrimSpace(connName))
		}
	}
	return conns
}

// isSlaveBridgeConnection checks if the connection is a slave bridge connection, which should be ignored for DNS configuration
// in case of an error while fetching type of the connection, it returns true (so the connection is anyway filtered out)
func (nmcli *NMCli) isSlaveBridgeConnection(connectionName string) bool {
	if out, err := nmcli.cmdExecutor(nmCliExecutable, "-t", "-f", nmCliSlaveBridgeType, nmCliConKey, "show", connectionName); err == nil {
		parts := strings.Split(string(out), ":")
		if len(parts) < 2 {
			// since output is malformed, we chose safe-path here, and assume this is a slave bridge connection
			log.Println(internal.DebugPrefix, dnsPrefix, "Received malformed output while getting bridge connection type, assuming", connectionName, " is a slave")
			return true
		}
		isSlaveBridge := strings.TrimSpace(parts[1])
		//for non-slave type bridges, this field is empty
		return len(isSlaveBridge) > 0
	}
	return true
}

// reloadConnection restarts the network connection for the specified connection name using nmcli tool.
func (nmcli *NMCli) reloadConnection(connectionName string) error {
	reloadArgs := []string{nmCliConKey, "reload"}
	if out, err := nmcli.cmdExecutor(nmCliExecutable, reloadArgs...); err != nil {
		log.Println(internal.WarningPrefix, dnsPrefix, ":", strings.TrimSpace(string(out)))
		return fmt.Errorf("reload connection failed: %w", err)
	}

	upArgs := []string{nmCliConKey, "up", connectionName}
	if _, err := nmcli.cmdExecutor(nmCliExecutable, upArgs...); err != nil {
		return fmt.Errorf("Setting %s UP failed with:%w", connectionName, err)
	}
	return nil
}

// getConnectionState retrieves current DNS configuration for a connection
func (nmcli *NMCli) getConnectionState(connectionName string) (*connectionState, error) {
	// Get ipv4.dns
	dnsOut, err := nmcli.cmdExecutor(nmCliExecutable, "-t", "-f", nmCliIPv4DNSKey, nmCliConKey, "show", connectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s for connection %s: %w", nmCliIPv4DNSKey, connectionName, err)
	}

	// Get ipv4.ignore-auto-dns
	ignoreOut, err := nmcli.cmdExecutor(nmCliExecutable, "-t", "-f", nmCliIPIgnoreAutoDnsKey, nmCliConKey, "show", connectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s for connection %s: %w", nmCliIPIgnoreAutoDnsKey, connectionName, err)
	}

	// Parse output (format: "property:value")
	dnsValue := strings.TrimSpace(strings.TrimPrefix(string(dnsOut), nmCliIPv4DNSKey+":"))
	ignoreValue := strings.TrimSpace(strings.TrimPrefix(string(ignoreOut), nmCliIPIgnoreAutoDnsKey+":"))

	return &connectionState{
		name:          connectionName,
		ipv4DNS:       dnsValue,
		ignoreAutoDNS: ignoreValue,
		isModified:    false,
	}, nil
}

// restoreConnectionState restores DNS configuration for a connection
func (nmcli *NMCli) restoreConnectionState(state *connectionState) error {
	args := []string{nmCliConKey, "modify", state.name, nmCliIPv4DNSKey, state.ipv4DNS, nmCliIPIgnoreAutoDnsKey, state.ignoreAutoDNS}
	if _, err := nmcli.cmdExecutor(nmCliExecutable, args...); err != nil {
		return fmt.Errorf("failed to restore connection %s: %w", state.name, err)
	}
	return nmcli.reloadConnection(state.name)
}

// rollback restores original DNS configuration for all modified connections
func (nmcli *NMCli) rollback(originalStates map[string]*connectionState) {
	log.Println(internal.WarningPrefix, dnsPrefix, "Rolling back DNS changes...")
	for con, state := range originalStates {
		if state.isModified {
			if err := nmcli.restoreConnectionState(state); err != nil {
				log.Println(internal.WarningPrefix, dnsPrefix, "Failed to rollback connection", con, ":", err)
			} else {
				log.Println(internal.InfoPrefix, dnsPrefix, "Successfully rolled back connection", con)
			}
		}
	}
}
