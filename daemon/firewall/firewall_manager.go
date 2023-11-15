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

// IptablesExecutor is an abstraction over iptables command. It should accept any input that is also viable for iptables
// command.
type IptablesExecutor interface {
	InsertRule(command string) error
	DeleteRule(command string) error
	InsertRuleIPv6(command string) error
	DeleteRuleIPv6(command string) error
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

func (i Iptables) executeCommand(ipv6 bool, args ...string) error {
	if !i.enabled {
		return nil
	}

	command := iptablesCommand
	if ipv6 {
		if !i.ip6tablesSupported {
			return errors.New("ip6tables are not supported")
		}

		command = ip6tablesCommand
	}

	// #nosec G204 -- arg values are known before even running the program
	if _, err := exec.Command(command, args...).CombinedOutput(); err != nil {
		return err
	}

	return nil
}

func (i Iptables) InsertRule(command string) error {
	commandArgs := strings.Split("-I "+command, " ")

	return i.executeCommand(false, commandArgs...)
}

func (i Iptables) DeleteRule(command string) error {
	commandArgs := strings.Split("-D "+command, " ")

	return i.executeCommand(false, commandArgs...)
}

func (i Iptables) InsertRuleIPv6(command string) error {
	commandArgs := strings.Split("-I "+command, " ")

	return i.executeCommand(true, commandArgs...)
}

func (i Iptables) DeleteRuleIPv6(command string) error {
	commandArgs := strings.Split("-D "+command, " ")

	return i.executeCommand(true, commandArgs...)
}

type PortRange struct {
	min int
	max int
}

type FirewallManager struct {
	commandExecutor   IptablesExecutor
	devices           device.ListFunc // list network interfaces
	allowlistRules    []string
	trafficBlockRules []string
	apiAllowlistRules []string
	connmark          uint32
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
		if err := f.commandExecutor.InsertRule(inputCommand); err != nil {
			return fmt.Errorf("blocking input traffic: %w", err)
		}
		f.trafficBlockRules = append(f.trafficBlockRules, inputCommand)

		outputCommand := fmt.Sprintf("OUTPUT -o %s -m comment --comment nordvpn -j DROP", iface.Name)
		if err := f.commandExecutor.InsertRule(outputCommand); err != nil {
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
		if err := f.commandExecutor.DeleteRule(rule); err != nil {
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
	if err := f.commandExecutor.InsertRule(inputDportRule); err != nil {
		return fmt.Errorf("allowlisting input dport: %w", err)
	}
	f.allowlistRules = append(f.allowlistRules, inputDportRule)

	inputSportRule := fmt.Sprintf("INPUT -i %s -p %s -m %s --sport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	if err := f.commandExecutor.InsertRule(inputSportRule); err != nil {
		return fmt.Errorf("allowlisting input sport: %w", err)
	}
	f.allowlistRules = append(f.allowlistRules, inputSportRule)

	outputDportRule := fmt.Sprintf("OUTPUT -o %s -p %s -m %s --dport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	if err := f.commandExecutor.InsertRule(outputDportRule); err != nil {
		return fmt.Errorf("allowlisting output dport: %w", err)
	}
	f.allowlistRules = append(f.allowlistRules, outputDportRule)

	outputSportRule := fmt.Sprintf("OUTPUT -o %s -p %s -m %s --sport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	if err := f.commandExecutor.InsertRule(outputSportRule); err != nil {
		return fmt.Errorf("allowlisting input dport: %w", err)
	}
	f.allowlistRules = append(f.allowlistRules, outputSportRule)

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
			inputRule := fmt.Sprintf("INPUT -s %s -i %s -m comment --comment nordvpn -j ACCEPT", subnet.String(), iface.Name)
			if err := f.commandExecutor.InsertRule(inputRule); err != nil {
				return fmt.Errorf("adding input accept rule for subnet: %w", err)
			}
			f.allowlistRules = append(f.allowlistRules, inputRule)

			outputRule := fmt.Sprintf("OUTPUT -d %s -o %s -m comment --comment nordvpn -j ACCEPT", subnet.String(), iface.Name)
			if err := f.commandExecutor.InsertRule(outputRule); err != nil {
				return fmt.Errorf("adding output accept rule for subnet: %w", err)
			}
			f.allowlistRules = append(f.allowlistRules, outputRule)
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
	for _, rule := range f.allowlistRules {
		if err := f.commandExecutor.DeleteRule(rule); err != nil {
			return fmt.Errorf("removing allowlist rule: %w", err)
		}
	}

	f.allowlistRules = nil

	return nil
}

// ApiAllowlist adds ACCEPT rules for privileged traffic, for each interface.
func (f *FirewallManager) ApiAllowlist() error {
	ifaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	for _, iface := range ifaces {
		inputRule := fmt.Sprintf("INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, f.connmark)
		if err := f.commandExecutor.InsertRule(inputRule); err != nil {
			return fmt.Errorf("adding api allowlist INPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, inputRule)

		outputRule :=
			fmt.Sprintf("OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
				iface.Name, f.connmark)
		if err := f.commandExecutor.InsertRule(outputRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, outputRule)

		outputConnmarkRule := fmt.Sprintf("OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, f.connmark)
		if err := f.commandExecutor.InsertRule(outputConnmarkRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, outputConnmarkRule)
	}

	return nil
}

// ApiDenylis removes ACCEPT rules added by ApiAllowlist.
func (f *FirewallManager) ApiDenylist() error {
	for _, rule := range f.apiAllowlistRules {
		if err := f.commandExecutor.DeleteRule(rule); err != nil {
			return fmt.Errorf("removing api allowlist rule: %w", err)
		}
	}

	f.apiAllowlistRules = nil

	return nil
}
