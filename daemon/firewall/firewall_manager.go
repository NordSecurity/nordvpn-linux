package firewall

import (
	"errors"
	"fmt"
	"net/netip"
	"os/exec"
	"sort"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
)

var ErrRuleAlreadyActive = errors.New("this rule is already active")

const (
	iptablesCommand  = "iptables"
	ip6tablesCommand = "ip6tables"
)

type IPVersion int

const (
	IPv4 = iota
	IPv6
	Both
)

// IptablesExecutor is an abstraction over iptables command. It should accept any input that is also viable for iptables
// command.
type IptablesExecutor interface {
	InsertRule(rule string, version IPVersion) error
	DeleteRule(rule string, version IPVersion) error
	// Enable changes the Executor state so that it performs the commands provided to ExecuteCommand and
	// ExecuteCommandIPv6.
	Enable()
	// Disable changes the Executor state so that it does not perform commands provided to ExecuteCommands and
	// ExecuteCommandIPv6.
	Disable()
}

type Iptables struct {
	ip6tablesSupported bool
	enabled            bool
}

func areIp6tablesSupported() bool {
	// #nosec G204 -- input is properly sanitized
	_, err := exec.Command(ip6tablesCommand, "-S").CombinedOutput()
	return err != nil
}

func NewIptables() Iptables {
	return Iptables{
		ip6tablesSupported: areIp6tablesSupported(),
	}
}

func (i Iptables) executeCommand(version IPVersion, args ...string) error {
	if !i.enabled {
		return nil
	}

	switch version {
	case IPv4:
		if _, err := exec.Command(iptablesCommand, args...).CombinedOutput(); err != nil {
			return err
		}
	case IPv6:
		if i.enabled {
			if _, err := exec.Command(ip6tablesCommand, args...).CombinedOutput(); err != nil {
				return err
			}
		}
	case Both:
		if _, err := exec.Command(iptablesCommand, args...).CombinedOutput(); err != nil {
			return err
		}
		if i.enabled {
			if _, err := exec.Command(ip6tablesCommand, args...).CombinedOutput(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (i Iptables) InsertRule(rule string, version IPVersion) error {
	commandArgs := strings.Split("-I "+rule, " ")

	return i.executeCommand(version, commandArgs...)
}

func (i Iptables) DeleteRule(rule string, version IPVersion) error {
	commandArgs := strings.Split("-D "+rule, " ")

	return i.executeCommand(version, commandArgs...)
}

type PortRange struct {
	min int
	max int
}

type AllowlistSubnet struct {
	rule      string
	ipVersion IPVersion
}

type FirewallManager struct {
	commandExecutor      IptablesExecutor
	devices              device.ListFunc // list network interfaces
	allowlistPortsRules  []string
	allowlistSubnetRules []AllowlistSubnet
	trafficBlockRules    []string
	apiAllowlistRules    []string
	connmark             uint32
}

func NewFirewallManager(devices device.ListFunc, commandExecutor IptablesExecutor, connmark uint32) FirewallManager {
	return FirewallManager{
		commandExecutor: commandExecutor,
		devices:         devices,
		connmark:        connmark,
	}
}

// BlocTraffic adds DROP rules for all the incoming traffic, for every viable network interface
func (f *FirewallManager) BlockTraffic() error {
	if f.trafficBlockRules != nil {
		return ErrRuleAlreadyActive
	}

	interfaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	// -I INPUT -i <iface> -m comment --comment nordvpn -j DROP
	// -I OUTPUT -o <iface> -m comment --comment nordvpn -j DROP
	for _, iface := range interfaces {
		inputCommand := fmt.Sprintf("INPUT -i %s -m comment --comment nordvpn -j DROP", iface.Name)
		if err := f.commandExecutor.InsertRule(inputCommand, Both); err != nil {
			return fmt.Errorf("blocking input traffic: %w", err)
		}
		f.trafficBlockRules = append(f.trafficBlockRules, inputCommand)

		outputCommand := fmt.Sprintf("OUTPUT -o %s -m comment --comment nordvpn -j DROP", iface.Name)
		if err := f.commandExecutor.InsertRule(outputCommand, Both); err != nil {
			return fmt.Errorf("blocking output traffic: %w", err)
		}
		f.trafficBlockRules = append(f.trafficBlockRules, outputCommand)
	}
	return nil
}

func (f *FirewallManager) removeBlockTrafficRules() error {
	// -D INPUT -i <iface> -m comment --comment nordvpn -j DROP
	// -D OUTPUT -o <iface> -m comment --comment nordvpn -j DROP
	for _, rule := range f.trafficBlockRules {
		if err := f.commandExecutor.DeleteRule(rule, Both); err != nil {
			return fmt.Errorf("unblocking input traffic: %w", err)
		}
	}

	return nil
}

// UnblockTraffic removes DROP rules added by BlockTraffic. Returns an error if BlockTraffic was not previously called.
func (f *FirewallManager) UnblockTraffic() error {
	if f.trafficBlockRules == nil {
		return ErrRuleAlreadyActive
	}

	if err := f.removeBlockTrafficRules(); err != nil {
		return fmt.Errorf("removing traffic block rules: %w", err)
	}

	f.trafficBlockRules = nil

	return nil
}

// portsToPortRanges groups ports into ranges
func portsToPortRanges(ports []int) []PortRange {
	if len(ports) == 0 {
		return nil
	}

	sort.Ints(ports)

	var ranges []PortRange
	pPort := ports[0]
	r := PortRange{min: pPort, max: pPort}
	for i, port := range ports[1:] {
		if port == ports[i]+1 {
			r.max = port
			continue
		}
		ranges = append(ranges, r)
		r = PortRange{min: port, max: port}
	}

	return append(ranges, r)
}

func (f *FirewallManager) allowlistPort(iface string, protocol string, portRange PortRange) error {
	// -A INPUT -i <interface> -p <protocol> -m <protocol> --dport <port> -m comment --comment nordvpn -j ACCEPT
	// -A INPUT -i <interface> -p <protocol> -m <protocol> --sport <port> -m comment --comment nordvpn -j ACCEPT
	// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --sport <port> -m comment --comment nordvpn -j ACCEPT
	// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --dport <port> -m comment --comment nordvpn -j ACCEPT

	inputDportRule := fmt.Sprintf("INPUT -i %s -p %s -m %s --dport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	if err := f.commandExecutor.InsertRule(inputDportRule, Both); err != nil {
		return fmt.Errorf("allowlisting input dport: %w", err)
	}
	f.allowlistPortsRules = append(f.allowlistPortsRules, inputDportRule)

	inputSportRule := fmt.Sprintf("INPUT -i %s -p %s -m %s --sport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	if err := f.commandExecutor.InsertRule(inputSportRule, Both); err != nil {
		return fmt.Errorf("allowlisting input sport: %w", err)
	}
	f.allowlistPortsRules = append(f.allowlistPortsRules, inputSportRule)

	outputDportRule := fmt.Sprintf("OUTPUT -o %s -p %s -m %s --dport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	if err := f.commandExecutor.InsertRule(outputDportRule, Both); err != nil {
		return fmt.Errorf("allowlisting output dport: %w", err)
	}
	f.allowlistPortsRules = append(f.allowlistPortsRules, outputDportRule)

	outputSportRule := fmt.Sprintf("OUTPUT -o %s -p %s -m %s --sport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	if err := f.commandExecutor.InsertRule(outputSportRule, Both); err != nil {
		return fmt.Errorf("allowlisting input dport: %w", err)
	}
	f.allowlistPortsRules = append(f.allowlistPortsRules, outputSportRule)

	return nil
}

// SetAllowlist adds allowlist rules for the given udpPorts, tcpPorts and subnets.
func (f *FirewallManager) SetAllowlist(udpPorts []int, tcpPorts []int, subnets []netip.Prefix) error {
	ifaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	for _, subnet := range subnets {
		for _, iface := range ifaces {
			version := IPVersion(IPv4)
			if subnet.Addr().Is6() {
				version = IPv6
			}

			inputRule := fmt.Sprintf("INPUT -s %s -i %s -m comment --comment nordvpn -j ACCEPT", subnet.String(), iface.Name)
			if err := f.commandExecutor.InsertRule(inputRule, version); err != nil {
				return fmt.Errorf("adding input accept rule for subnet: %w", err)
			}
			f.allowlistSubnetRules = append(f.allowlistSubnetRules, AllowlistSubnet{rule: inputRule, ipVersion: version})

			outputRule := fmt.Sprintf("OUTPUT -d %s -o %s -m comment --comment nordvpn -j ACCEPT", subnet.String(), iface.Name)
			if err := f.commandExecutor.InsertRule(outputRule, version); err != nil {
				return fmt.Errorf("adding output accept rule for subnet: %w", err)
			}
			f.allowlistSubnetRules = append(f.allowlistSubnetRules, AllowlistSubnet{rule: outputRule, ipVersion: version})
		}
	}

	udpPortRanges := portsToPortRanges(udpPorts)
	for _, portRange := range udpPortRanges {
		for _, iface := range ifaces {
			if err := f.allowlistPort(iface.Name, "udp", portRange); err != nil {
				return fmt.Errorf("allowlisting udp ports: %w", err)
			}
		}
	}

	tcpPortRanges := portsToPortRanges(tcpPorts)
	for _, portRange := range tcpPortRanges {
		for _, iface := range ifaces {
			if err := f.allowlistPort(iface.Name, "tcp", portRange); err != nil {
				return fmt.Errorf("allowlisting tcp ports: %w", err)
			}
		}
	}

	return nil
}

// UnsetAllowlist removes all the rules added by SetAllowlist.
func (f *FirewallManager) UnsetAllowlist() error {
	for _, rule := range f.allowlistSubnetRules {
		if err := f.commandExecutor.DeleteRule(rule.rule, rule.ipVersion); err != nil {
			return fmt.Errorf("removing allowlist rule: %w", err)
		}
	}
	f.allowlistSubnetRules = nil

	for _, rule := range f.allowlistPortsRules {
		if err := f.commandExecutor.DeleteRule(rule, Both); err != nil {
			return fmt.Errorf("removing allowlist rule: %w", err)
		}
	}
	f.allowlistPortsRules = nil

	return nil
}

// APIAllowlist adds ACCEPT rules for privileged traffic, for each interface.
func (f *FirewallManager) APIAllowlist() error {
	ifaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	for _, iface := range ifaces {
		inputRule := fmt.Sprintf("INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, f.connmark)
		if err := f.commandExecutor.InsertRule(inputRule, Both); err != nil {
			return fmt.Errorf("adding api allowlist INPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, inputRule)

		outputRule :=
			fmt.Sprintf("OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
				iface.Name, f.connmark)
		if err := f.commandExecutor.InsertRule(outputRule, Both); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, outputRule)

		outputConnmarkRule := fmt.Sprintf("OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, f.connmark)
		if err := f.commandExecutor.InsertRule(outputConnmarkRule, Both); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, outputConnmarkRule)
	}

	return nil
}

// ApiDenylis removes ACCEPT rules added by ApiAllowlist.
func (f *FirewallManager) APIDenylist() error {
	for _, rule := range f.apiAllowlistRules {
		if err := f.commandExecutor.DeleteRule(rule, Both); err != nil {
			return fmt.Errorf("removing api allowlist rule: %w", err)
		}
	}

	f.apiAllowlistRules = nil

	return nil
}
