package exitnode

import (
	"fmt"
	"math"
	"net/netip"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/iptables"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type operation bool

const (
	meshAllowlistRuleComment = "nordvpn-exitnode-allowlist"
)

// Used when adding or removing rules from firewall
var (
	REMOVE operation = false
	ADD    operation = true
)

// allowlistManager manages forwarding rules based on the configured allowlist
// This is necessary when when user doesn't have full access to their own LAN ((killswitch || VPN) && !lanDiscovery),
// but have allowlisted local IPs or ports. We enable access to allowlisted destinations for peers which are granted
// local network access meshnet permission.
type allowlistManager struct {
	allowlist      config.Allowlist
	peers          mesh.MachinePeers
	runCommandFunc runCommandFunc
}

func newAllowlist(runCommandFunc runCommandFunc, allowlist config.Allowlist) allowlistManager {
	return allowlistManager{
		allowlist:      allowlist,
		runCommandFunc: runCommandFunc,
	}
}

// allowlistToFirewall adds forwarding rules for subnets and ports in the allowlist
func allowlistToFirewall(allowlist config.Allowlist, peers mesh.MachinePeers, op operation, commandFunc runCommandFunc) error {
	for subnet := range allowlist.Subnets {
		parsedSubnet, err := netip.ParsePrefix(subnet)
		if err != nil {
			return fmt.Errorf("failed to parse subnet: %w", err)
		}
		if !parsedSubnet.Addr().IsPrivate() && !parsedSubnet.Addr().IsLinkLocalUnicast() {
			continue
		}

		err = allowlistRuleToFirewall(peers, op, subnet, "", commandFunc)
		if err != nil && op == ADD {
			return fmt.Errorf("adding allowlist rule to firewall: %w", err)
		}
	}

	for _, pair := range []struct {
		name  string
		ports map[int64]bool
	}{
		{name: "tcp", ports: allowlist.Ports.TCP},
		{name: "udp", ports: allowlist.Ports.UDP},
	} {
		var ports []int
		for port := range pair.ports {
			if port > math.MaxUint16 {
				continue
			}
			ports = append(ports, int(port))
		}

		for _, portRange := range iptables.PortsToPortRanges(ports) {
			destination := fmt.Sprintf("%d:%d", portRange.Min, portRange.Max)
			if portRange.Min == portRange.Max {
				destination = fmt.Sprintf("%d", portRange.Min)
			}

			err := allowlistRuleToFirewall(peers, op, destination, pair.name, commandFunc)
			if err != nil {
				return fmt.Errorf("adding allowlist rule to firewall: %w", err)
			}
		}
	}

	return nil
}

func allowlistRuleToFirewall(
	peers mesh.MachinePeers,
	op operation,
	destination string, // port, port range or subnet
	portType string, // if empty then destination is subnet
	commandFunc runCommandFunc,
) error {
	flag := "-I"
	if op == REMOVE {
		flag = "-D"
	}

	for _, peer := range peers {
		if !peer.DoIAllowLocalNetwork || !peer.Address.IsValid() {
			continue
		}
		if portType != "" && !peer.DoIAllowRouting {
			// Don't insert port rules if routing is not permitted as that would basically allow routing to outside world on those ports
			continue
		}

		command := []string{
			flag,
			"FORWARD",
			"-s",
			peer.Address.String(),
			"-j",
			"ACCEPT",
			"-m",
			"comment",
			"--comment",
			meshAllowlistRuleComment,
			"-w",
			internal.SecondsToWaitForIptablesLock,
		}

		if portType == "" {
			command = append(command, "-d", destination)
		} else {
			command = append(command, "-p", portType, "-m", portType, "--dport", destination)
		}

		output, err := commandFunc(iptablesCmd, command...)
		if err != nil && !strings.Contains(string(output), missingRuleMessage) {
			return fmt.Errorf("calling iptables: %w, %s", err, output)
		}
	}

	return nil
}

func (a *allowlistManager) setAllowlist(allowlist config.Allowlist) {
	a.allowlist = allowlist
}

func (a *allowlistManager) setPeers(peers mesh.MachinePeers) {
	a.peers = peers
}

func (a *allowlistManager) enableAllowlist() error {
	return allowlistToFirewall(a.allowlist, a.peers, ADD, a.runCommandFunc)
}

func (a *allowlistManager) disableAllowlist() error {
	return allowlistToFirewall(a.allowlist, a.peers, REMOVE, a.runCommandFunc)
}
