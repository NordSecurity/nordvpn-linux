package firewall

import (
	"errors"
	"fmt"
	"net/netip"
	"sort"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	iptablesmanager "github.com/NordSecurity/nordvpn-linux/daemon/firewall/iptables_manager"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
)

var (
	ErrRuleAlreadyActive = errors.New("rule is already active")
	ErrRuleNotActive     = errors.New("rule does not exist")
)

type PortRange struct {
	min int
	max int
}

const (
	TrafficBlock iptablesmanager.RulePriority = iota
	ApiAllowlistMark
	ApiAllowlistOutputConnmark
	UserAllowlist
	MeshnetFileshare
	MeshnetIncoming
	MeshnetBlockIncomingLAN
)

type meshIncomingRule struct {
	allowIncomingRule iptablesmanager.FwRule
	blockLocalRules   []iptablesmanager.FwRule
}

type FirewallManager struct {
	iptablesManager iptablesmanager.IPTablesManager
	// list network interfaces
	devices           device.ListFunc
	allowlistRules    []iptablesmanager.FwRule
	trafficBlockRules []iptablesmanager.FwRule
	apiAllowlistRules []iptablesmanager.FwRule
	// maps peer UID to rules related to allowing incoming traffic
	allowIncomingRules map[string]meshIncomingRule
	// maps peer UID to rules that allow fileshare
	fileshareRules map[string]iptablesmanager.FwRule
	connmark       uint32
}

func NewFirewallManager(devices device.ListFunc,
	cmdRunner iptablesmanager.CommandRunner,
	connmark uint32,
	enabled bool) FirewallManager {
	return FirewallManager{
		iptablesManager:    iptablesmanager.NewIPTablesManager(cmdRunner, enabled),
		devices:            devices,
		allowIncomingRules: make(map[string]meshIncomingRule),
		fileshareRules:     make(map[string]iptablesmanager.FwRule),
		connmark:           connmark,
	}
}

func (f *FirewallManager) AllowIncoming(peer meshnet.UniqueAddress, allowLocal bool) error {
	if _, ok := f.allowIncomingRules[peer.UID]; ok {
		return ErrRuleAlreadyActive
	}

	blockLANRules := []iptablesmanager.FwRule{}
	if !allowLocal {
		lans := []string{
			"169.254.0.0/16",
			"192.168.0.0/16",
			"172.16.0.0/12",
			"10.0.0.0/8",
		}

		for _, lan := range lans {
			rule := iptablesmanager.NewFwRule(
				iptablesmanager.Input,
				iptablesmanager.IPv4,
				fmt.Sprintf("-s %s/32 -d %s -j DROP", peer.Address, lan),
				MeshnetBlockIncomingLAN)
			blockLANRules = append(blockLANRules, rule)

			if err := f.iptablesManager.InsertRule(rule); err != nil {
				return fmt.Errorf("blocking mesh peer from LAN access: %w", err)
			}
		}
	}

	rule := iptablesmanager.NewFwRule(
		iptablesmanager.Input,
		iptablesmanager.IPv4,
		fmt.Sprintf("-s %s/32 -j ACCEPT", peer.Address),
		MeshnetIncoming)

	if err := f.iptablesManager.InsertRule(rule); err != nil {
		return fmt.Errorf("allowing incoming traffic for peer: %w", err)
	}

	f.allowIncomingRules[peer.UID] = meshIncomingRule{
		allowIncomingRule: rule,
		blockLocalRules:   blockLANRules,
	}

	return nil
}

func (f *FirewallManager) DenyIncoming(peerUID string) error {
	rule, ok := f.allowIncomingRules[peerUID]

	if !ok {
		return ErrRuleNotFound
	}

	if err := f.iptablesManager.DeleteRule(rule.allowIncomingRule); err != nil {
		return fmt.Errorf("removing allow incoming rule: %w", err)
	}

	for _, blockLANRule := range rule.blockLocalRules {
		if err := f.iptablesManager.DeleteRule(blockLANRule); err != nil {
			return fmt.Errorf("removing block LAN rule: %w", err)
		}
	}

	delete(f.allowIncomingRules, peerUID)

	return nil
}

// AllowFileshare adds ACCEPT rule for all incoming connections to tcp port 49111 from the peer with given UniqueAddress.
func (f *FirewallManager) AllowFileshare(peer meshnet.UniqueAddress) error {
	if _, ok := f.fileshareRules[peer.UID]; ok {
		return ErrRuleAlreadyActive
	}

	args := fmt.Sprintf("-s %s/32 -p tcp -m tcp --dport 49111 -j ACCEPT", peer.Address.String())
	rule := iptablesmanager.NewFwRule(iptablesmanager.Input, iptablesmanager.IPv4, args, MeshnetFileshare)
	if err := f.iptablesManager.InsertRule(rule); err != nil {
		return fmt.Errorf("adding allow fileshare rule: %w", err)
	}

	f.fileshareRules[peer.UID] = rule
	return nil
}

// DenyFileshare removes ACCEPT rule for all incoming connections to tcp port 49111 from the peer with given UniqueAddress.
func (f *FirewallManager) DenyFileshare(peerUID string) error {
	rule, ok := f.fileshareRules[peerUID]
	if !ok {
		return ErrRuleNotActive
	}

	if err := f.iptablesManager.DeleteRule(rule); err != nil {
		return fmt.Errorf("deleting fileshare rule: %w", err)
	}

	delete(f.fileshareRules, peerUID)
	return nil
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
			iptablesmanager.IPv4,
			inputParams,
			TrafficBlock)
		if err := f.iptablesManager.InsertRule(inputRule); err != nil {
			return fmt.Errorf("blocking input traffic: %w", err)
		}
		f.trafficBlockRules = append(f.trafficBlockRules, inputRule)

		outputParams := fmt.Sprintf("-o %s -j DROP", iface.Name)
		outputRule := iptablesmanager.NewFwRule(
			iptablesmanager.Output,
			iptablesmanager.IPv4,
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
		iptablesmanager.IPv4,
		inputDportParams,
		UserAllowlist)
	if err := f.allowlistPort(inputDportRule); err != nil {
		return fmt.Errorf("allowlisting input dport: %w", err)
	}

	inputSportParams := fmt.Sprintf("-i %s -p %s -m %s --sport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	inputSportRule := iptablesmanager.NewFwRule(
		iptablesmanager.Input,
		iptablesmanager.IPv4,
		inputSportParams,
		UserAllowlist)
	if err := f.allowlistPort(inputSportRule); err != nil {
		return fmt.Errorf("allowlisting input sport: %w", err)
	}

	outputDportParams := fmt.Sprintf("-o %s -p %s -m %s --dport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputDportRule := iptablesmanager.NewFwRule(
		iptablesmanager.Output,
		iptablesmanager.IPv4,
		outputDportParams,
		UserAllowlist)
	if err := f.allowlistPort(outputDportRule); err != nil {
		return fmt.Errorf("allowlisting output dport: %w", err)
	}

	outputSportParams := fmt.Sprintf("-o %s -p %s -m %s --sport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputSportRule := iptablesmanager.NewFwRule(
		iptablesmanager.Output,
		iptablesmanager.IPv4,
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
				return errors.New("IPv6 not supported")
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
			iptablesmanager.IPv4,
			inputParams,
			ApiAllowlistMark)
		if err := f.iptablesManager.InsertRule(inputRule); err != nil {
			return fmt.Errorf("adding api allowlist INPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, inputRule)

		outputConnmarkParams := fmt.Sprintf("-o %s -m connmark --mark %d -j ACCEPT", iface.Name, f.connmark)
		outputConnmarkRule := iptablesmanager.NewFwRule(
			iptablesmanager.Output,
			iptablesmanager.IPv4,
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
			iptablesmanager.IPv4,
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
