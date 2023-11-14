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
	iptables  = "iptables"
	ip6tables = "ip6tables"
)

type IptablesExecutor interface {
	ExecuteCommand(command string) error
	ExecuteCommandIPv6(command string) error
}

type Iptables struct {
	ip6tablesSupported bool
}

func AreIp6tablesSupported() bool {
	// #nosec G204 -- input is properly sanitized
	_, err := exec.Command(ip6tables, "-S").CombinedOutput()
	return err != nil
}

func NewIptables() Iptables {
	return Iptables{
		ip6tablesSupported: AreIp6tablesSupported(),
	}
}

func (i Iptables) ExecuteCommand(command string) error {
	commandArgs := strings.Split(command, " ")

	// #nosec G204 -- arg values are known before even running the program
	if _, err := exec.Command(iptables, commandArgs...).CombinedOutput(); err != nil {
		return err
	}

	return nil
}

func (i Iptables) ExecuteCommandIPv6(command string) error {
	if !i.ip6tablesSupported {
		return errors.New("ip6tables are not supported")
	}

	commandArgs := strings.Split(command, " ")

	// #nosec G204 -- arg values are known before even running the program
	if _, err := exec.Command(ip6tables, commandArgs...).CombinedOutput(); err != nil {
		return err
	}

	return nil
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
	connmark          uint32
	enabled           bool
}

func NewFirewallManager(devices device.ListFunc, commandExecutor IptablesExecutor, connmark uint32, enabled bool) FirewallManager {
	return FirewallManager{
		commandExecutor: commandExecutor,
		devices:         devices,
		connmark:        connmark,
		enabled:         enabled,
	}
}

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
		outputCommand := fmt.Sprintf("OUTPUT -o %s -m comment --comment nordvpn -j DROP", iface.Name)
		f.trafficBlockRules = append(f.trafficBlockRules, inputCommand)
		f.trafficBlockRules = append(f.trafficBlockRules, outputCommand)

		if f.enabled {
			if err := f.commandExecutor.ExecuteCommand("-I " + inputCommand); err != nil {
				return fmt.Errorf("blocking input traffic: %w", err)
			}

			if err := f.commandExecutor.ExecuteCommand("-I " + outputCommand); err != nil {
				return fmt.Errorf("blocking output traffic: %w", err)
			}
		}
	}
	return nil
}

func (f *FirewallManager) removeBlockTrafficRules() error {
	// -D INPUT -i <iface> -m comment --comment nordvpn -j DROP
	// -D OUTPUT -o <iface> -m comment --comment nordvpn -j DROP
	for _, rule := range f.trafficBlockRules {
		if err := f.commandExecutor.ExecuteCommand("-D " + rule); err != nil {
			return fmt.Errorf("unblocking input traffic: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) UnblockTraffic() error {
	if f.trafficBlockRules == nil {
		return ErrRuleAlreadyActive
	}

	if f.enabled {
		if err := f.removeBlockTrafficRules(); err != nil {
			return fmt.Errorf("removing traffic block rules: %w", err)
		}
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
	inputSportRule := fmt.Sprintf("INPUT -i %s -p %s -m %s --sport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputDportRule := fmt.Sprintf("OUTPUT -o %s -p %s -m %s --dport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputSportRule := fmt.Sprintf("OUTPUT -o %s -p %s -m %s --sport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)

	if f.enabled {
		if err := f.commandExecutor.ExecuteCommand("-I " + inputDportRule); err != nil {
			return fmt.Errorf("allowlisting input dport: %w", err)
		}

		if err := f.commandExecutor.ExecuteCommand("-I " + inputSportRule); err != nil {
			return fmt.Errorf("allowlisting input sport: %w", err)
		}

		if err := f.commandExecutor.ExecuteCommand("-I " + outputDportRule); err != nil {
			return fmt.Errorf("allowlisting output dport: %w", err)
		}

		if err := f.commandExecutor.ExecuteCommand("-I " + outputSportRule); err != nil {
			return fmt.Errorf("allowlisting input dport: %w", err)
		}
	}

	f.allowlistRules = append(f.allowlistRules, inputDportRule)
	f.allowlistRules = append(f.allowlistRules, inputSportRule)
	f.allowlistRules = append(f.allowlistRules, outputDportRule)
	f.allowlistRules = append(f.allowlistRules, outputSportRule)

	return nil
}

func (f *FirewallManager) SetAllowlist(udpPorts []int, tcpPorts []int, subnets []netip.Prefix) error {
	ifaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	for _, subnet := range subnets {
		for _, iface := range ifaces {
			inputRule := fmt.Sprintf("INPUT -s %s -i %s -m comment --comment nordvpn -j ACCEPT", subnet.String(), iface.Name)
			outputRule := fmt.Sprintf("OUTPUT -d %s -o %s -m comment --comment nordvpn -j ACCEPT", subnet.String(), iface.Name)

			if f.enabled {
				if err := f.commandExecutor.ExecuteCommand("-I " + inputRule); err != nil {
					return fmt.Errorf("adding input accept rule for subnet: %w", err)
				}
				if err := f.commandExecutor.ExecuteCommand("-I " + outputRule); err != nil {
					return fmt.Errorf("adding output accept rule for subnet: %w", err)
				}
			}

			f.allowlistRules = append(f.allowlistRules, inputRule)
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

func (f *FirewallManager) UnsetAllowlist() error {
	if f.enabled {
		for _, rule := range f.allowlistRules {
			if err := f.commandExecutor.ExecuteCommand("-D " + rule); err != nil {
				return fmt.Errorf("removing allowlist rule: %w", err)
			}
		}
	}

	f.allowlistRules = nil

	return nil
}

func (f *FirewallManager) manageApiAllowlist(allow bool) error {
	iptablesMode := "-I "
	if !allow {
		iptablesMode = "-D "
	}

	ifaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	for _, iface := range ifaces {
		inputRule := fmt.Sprintf("INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, f.connmark)
		if err := f.commandExecutor.ExecuteCommand(iptablesMode + inputRule); err != nil {
			return fmt.Errorf("adding api allowlist INPUT rule: %w", err)
		}

		outputRule :=
			fmt.Sprintf("OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
				iface.Name, f.connmark)
		if err := f.commandExecutor.ExecuteCommand(iptablesMode + outputRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}

		outputConnmarkRule := fmt.Sprintf("OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, f.connmark)
		if err := f.commandExecutor.ExecuteCommand(iptablesMode + outputConnmarkRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) ApiAllowlist() error {
	if !f.enabled {
		return nil
	}

	return f.manageApiAllowlist(true)
}

func (f *FirewallManager) ApiDenylist() error {
	if !f.enabled {
		return nil
	}

	return f.manageApiAllowlist(false)
}
