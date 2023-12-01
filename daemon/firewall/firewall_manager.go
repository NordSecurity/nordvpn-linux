package firewall

import (
	"errors"
	"fmt"
	"net/netip"
	"sort"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	iptablesmanager "github.com/NordSecurity/nordvpn-linux/daemon/firewall/iptables_manager"
)

var ErrRuleAlreadyActive = errors.New("this rule is already active")

type PortRange struct {
	min int
	max int
}

const (
	TrafficBlock iptablesmanager.RulePriority = iota
	ApiAllowlistMark
	ApiAllowlistOutputConnmark
	UserAllowlist
)

type FirewallManager struct {
	iptablesManager   iptablesmanager.IPTablesManager
	devices           device.ListFunc // list network interfaces
	allowlistRules    []iptablesmanager.FwRule
	trafficBlockRules []iptablesmanager.FwRule
	apiAllowlistRules []iptablesmanager.FwRule
	connmark          uint32
}

func NewFirewallManager(devices device.ListFunc,
	cmdRunner iptablesmanager.CommandRunner,
	connmark uint32,
	ip6TablesSupported bool,
	enabled bool) FirewallManager {
	return FirewallManager{
		iptablesManager: iptablesmanager.NewIPTablesManager(cmdRunner, ip6TablesSupported, enabled),
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

	// -I INPUT -i <iface> -j DROP
	// -I OUTPUT -o <iface> -j DROP
	for _, iface := range interfaces {
		inputParams := fmt.Sprintf("-i %s -j DROP", iface.Name)
		inputRule := iptablesmanager.NewFwRule(
			iptablesmanager.Input,
			iptablesmanager.Both,
			inputParams,
			TrafficBlock)
		if err := f.iptablesManager.InsertRule(inputRule); err != nil {
			return fmt.Errorf("blocking input traffic: %w", err)
		}
		f.trafficBlockRules = append(f.trafficBlockRules, inputRule)

		outputParams := fmt.Sprintf("-o %s -j DROP", iface.Name)
		outputRule := iptablesmanager.NewFwRule(
			iptablesmanager.Output,
			iptablesmanager.Both,
			outputParams,
			TrafficBlock)
		if err := f.iptablesManager.InsertRule(outputRule); err != nil {
			return fmt.Errorf("blocking output traffic: %w", err)
		}
		f.trafficBlockRules = append(f.trafficBlockRules, outputRule)
	}
	return nil
}

func (f *FirewallManager) removeBlockTrafficRules() error {
	// -D INPUT -i <iface> -j DROP
	// -D OUTPUT -o <iface> -j DROP
	for _, rule := range f.trafficBlockRules {
		if err := f.iptablesManager.DeleteRule(rule); err != nil {
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

func (f *FirewallManager) allowlistPort(rule iptablesmanager.FwRule) error {
	if err := f.iptablesManager.InsertRule(rule); err != nil {
		return fmt.Errorf("allowlisting port: %w", err)
	}
	f.allowlistRules = append(f.allowlistRules, rule)
	return nil
}

func (f *FirewallManager) allowlistPorts(iface string, protocol string, portRange PortRange) error {
	// -A INPUT -i <interface> -p <protocol> -m <protocol> --dport <port> -j ACCEPT
	// -A INPUT -i <interface> -p <protocol> -m <protocol> --sport <port> -j ACCEPT
	// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --sport <port> -j ACCEPT
	// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --dport <port> -j ACCEPT

	inputDportParams := fmt.Sprintf("-i %s -p %s -m %s --dport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	inputDportRule := iptablesmanager.NewFwRule(
		iptablesmanager.Input,
		iptablesmanager.Both,
		inputDportParams,
		UserAllowlist)
	if err := f.allowlistPort(inputDportRule); err != nil {
		return fmt.Errorf("allowlisting input dport: %w", err)
	}

	inputSportParams := fmt.Sprintf("-i %s -p %s -m %s --sport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	inputSportRule := iptablesmanager.NewFwRule(
		iptablesmanager.Input,
		iptablesmanager.Both,
		inputSportParams,
		UserAllowlist)
	if err := f.allowlistPort(inputSportRule); err != nil {
		return fmt.Errorf("allowlisting input sport: %w", err)
	}

	outputDportParams := fmt.Sprintf("-o %s -p %s -m %s --dport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputDportRule := iptablesmanager.NewFwRule(
		iptablesmanager.Output,
		iptablesmanager.Both,
		outputDportParams,
		UserAllowlist)
	if err := f.allowlistPort(outputDportRule); err != nil {
		return fmt.Errorf("allowlisting output dport: %w", err)
	}

	outputSportParams := fmt.Sprintf("-o %s -p %s -m %s --sport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputSportRule := iptablesmanager.NewFwRule(
		iptablesmanager.Output,
		iptablesmanager.Both,
		outputSportParams,
		UserAllowlist)
	if err := f.allowlistPort(outputSportRule); err != nil {
		return fmt.Errorf("allowlisting output sport: %w", err)
	}

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
			version := iptablesmanager.IPv4
			if subnet.Addr().Is6() {
				version = iptablesmanager.IPv6
			}

			inputParams := fmt.Sprintf("-s %s -i %s -j ACCEPT", subnet.String(), iface.Name)
			inputRule := iptablesmanager.NewFwRule(iptablesmanager.Input, version, inputParams, UserAllowlist)
			if err := f.iptablesManager.InsertRule(inputRule); err != nil {
				return fmt.Errorf("adding input accept rule for subnet: %w", err)
			}
			f.allowlistRules = append(f.allowlistRules, inputRule)

			outputParams := fmt.Sprintf("-d %s -o %s -j ACCEPT", subnet.String(), iface.Name)
			outputRule := iptablesmanager.NewFwRule(iptablesmanager.Output, version, outputParams, UserAllowlist)
			if err := f.iptablesManager.InsertRule(outputRule); err != nil {
				return fmt.Errorf("adding output accept rule for subnet: %w", err)
			}
			f.allowlistRules = append(f.allowlistRules, outputRule)
		}
	}

	udpPortRanges := portsToPortRanges(udpPorts)
	for _, portRange := range udpPortRanges {
		for _, iface := range ifaces {
			if err := f.allowlistPorts(iface.Name, "udp", portRange); err != nil {
				return fmt.Errorf("allowlisting udp ports: %w", err)
			}
		}
	}

	tcpPortRanges := portsToPortRanges(tcpPorts)
	for _, portRange := range tcpPortRanges {
		for _, iface := range ifaces {
			if err := f.allowlistPorts(iface.Name, "tcp", portRange); err != nil {
				return fmt.Errorf("allowlisting tcp ports: %w", err)
			}
		}
	}

	return nil
}

// UnsetAllowlist removes all the rules added by SetAllowlist.
func (f *FirewallManager) UnsetAllowlist() error {
	for _, rule := range f.allowlistRules {
		if err := f.iptablesManager.DeleteRule(rule); err != nil {
			return fmt.Errorf("removing allowlist rule: %w", err)
		}
	}

	f.allowlistRules = nil

	return nil
}

// APIAllowlist adds ACCEPT rules for privileged traffic, for each interface.
func (f *FirewallManager) APIAllowlist() error {
	ifaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	for _, iface := range ifaces {
		inputParams := fmt.Sprintf("-i %s -m connmark --mark %d -j ACCEPT", iface.Name, f.connmark)
		inputRule := iptablesmanager.NewFwRule(
			iptablesmanager.Input,
			iptablesmanager.Both,
			inputParams,
			ApiAllowlistMark)
		if err := f.iptablesManager.InsertRule(inputRule); err != nil {
			return fmt.Errorf("adding api allowlist INPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, inputRule)

		outputConnmarkParams := fmt.Sprintf("-o %s -m connmark --mark %d -j ACCEPT", iface.Name, f.connmark)
		outputConnmarkRule := iptablesmanager.NewFwRule(
			iptablesmanager.Output,
			iptablesmanager.Both,
			outputConnmarkParams,
			ApiAllowlistOutputConnmark)
		if err := f.iptablesManager.InsertRule(outputConnmarkRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, outputConnmarkRule)

		outputParams :=
			fmt.Sprintf("-o %s -m mark --mark %d -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
				iface.Name, f.connmark)
		outputRule := iptablesmanager.NewFwRule(
			iptablesmanager.Output,
			iptablesmanager.Both,
			outputParams,
			ApiAllowlistMark)
		if err := f.iptablesManager.InsertRule(outputRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, outputRule)
	}

	return nil
}

// ApiDenylis removes ACCEPT rules added by ApiAllowlist.
func (f *FirewallManager) APIDenylist() error {
	for _, rule := range f.apiAllowlistRules {
		if err := f.iptablesManager.DeleteRule(rule); err != nil {
			return fmt.Errorf("removing api allowlist rule: %w", err)
		}
	}

	f.apiAllowlistRules = nil

	return nil
}
