package firewall

import (
	"errors"
	"fmt"
	"net/netip"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
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
	IPv4 IPVersion = iota
	IPv6
	Both
)

type RulePriority int

const (
	TrafficBlock RulePriority = iota
	ApiAllowlistMark
	ApiAllowlistOutputConnmark
	UserAllowlist
)

func (r RulePriority) toCommentArgs() string {
	return fmt.Sprintf("-m comment --comment nordvpn-%d", r)
}

func (r RulePriority) toComment() string {
	return fmt.Sprintf("nordvpn-%d", r)
}

type Chain int

const (
	INPUT  = iota
	OUTPUT = iota
)

func (c Chain) String() string {
	switch c {
	case INPUT:
		return "INPUT"
	case OUTPUT:
		return "OUTPUT"
	}
	return ""
}

type CommandRunner interface {
	RunCommand(string, string) (string, error)
}

type ExecCommandRunner struct {
}

func (ExecCommandRunner) RunCommand(command string, args string) (string, error) {
	output, err := exec.Command(args, strings.Split(args, " ")...).CombinedOutput()
	return string(output), err
}

type iptablesManager struct {
	ip6tablesSupported bool
	enabled            bool
	commandRunner      CommandRunner
}

func AreIp6tablesSupported() bool {
	// #nosec G204 -- input is properly sanitized
	_, err := exec.Command(ip6tablesCommand, "-S").CombinedOutput()
	return err != nil
}

func newIptablesManager(commandRunner CommandRunner, enabled bool, ip6tablesSupported bool) iptablesManager {
	return iptablesManager{
		commandRunner:      commandRunner,
		enabled:            enabled,
		ip6tablesSupported: ip6tablesSupported,
	}
}

func (i iptablesManager) executeCommand(insert bool, rule FwRule) error {
	if !i.enabled {
		return nil
	}

	command := rule.ToDeleteCommand()

	if rule.ipVersion == IPv4 || rule.ipVersion == Both {
		if insert {
			index, err := i.getRuleLine(iptablesCommand, rule.chain, rule.priority)
			if err != nil {
				return fmt.Errorf("calculating rule index: %w", err)
			}
			command = rule.ToInsertAppendCommand(index)
		}

		if _, err := i.commandRunner.RunCommand(iptablesCommand, command); err != nil {
			return err
		}
	}

	if rule.ipVersion == IPv4 || !i.ip6tablesSupported {
		return nil
	}

	if insert {
		index, err := i.getRuleLine(ip6tablesCommand, rule.chain, rule.priority)
		if err != nil {
			return fmt.Errorf("calculating rule index: %w", err)
		}
		command = rule.ToInsertAppendCommand(index)
	}

	if _, err := i.commandRunner.RunCommand(ip6tablesCommand, command); err != nil {
		return err
	}

	return nil
}

// getRuleLine returns rules line number, based on given priority. It return line in iptables where the rule should be
// inserted to satisfy that priority.
//
// Line numbers in iptables are inserted in ascending order, with lower line number having priority over higher line
// number. Higher priority will generally result in lower line number. Line number 1 will be returned for the highest
// priority.
//
// Assumptions/desired behavior:
//  1. nordvpn rules(i.e rules that contain 'nordvpn-<priority> comment) should always have priority over non-nordvpn
//     rules.
//  2. Last nordvpn rule in the chain is our boundary. New rule should be inserted either above it, or at most at one
//     line bellow it. This enforces priority of nordvpn rules over non-nordvpn rules.
//  3. Non-nordvpn rules located between last nordvpn rule in the chain and the first rule in the chain are ignored.
func (i iptablesManager) getRuleLine(command string, chain Chain, priority RulePriority) (int, error) {
	// Run command with --numeric to avoid reverse DNS lookup. This takes a long time and is unecessary for the purpose
	// of line number calculation(we ignore everything but the 'nordvpn-<priority>' comment or the lack of thereof).
	args := "-L " + chain.String() + " --numeric"

	output, err := i.commandRunner.RunCommand(command, args)
	if err != nil {
		return 0, fmt.Errorf("listing iptables rules: %w", err)
	}

	// Skip first two lines of output they are the chain name and table values name.
	outputLines := strings.Split(string(output), "\n")[2:]

	if len(outputLines) == 0 {
		return 1, nil
	}

	nordvpnCommentPattern := regexp.MustCompile(`nordvpn-(\d+)`)

	lastNordvpnRuleLine := 0
	for ruleIndex, line := range outputLines {
		matches := nordvpnCommentPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			// Offset by 1 because iptables rules are 1 based
			lastNordvpnRuleLine = ruleIndex + 1
		}
	}

	// Rules are present in iptables, but there are no nordvpn rules. In this case we want to insert the rule at the top
	// so that it will be prioritized over other rules.
	if len(outputLines) > 0 && lastNordvpnRuleLine == 0 {
		return 1, nil
	}

	// get next lowest index
	for lineIndex, line := range outputLines[:lastNordvpnRuleLine] {
		if strings.Contains(line, priority.toComment()) {
			return lineIndex + 1, nil
		}

		matches := nordvpnCommentPattern.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}

		rulePriority, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, fmt.Errorf("converting priority to int: %w", err)
		}

		if rulePriority < int(priority) {
			// Array indexes are 0 based, iptables line numbers are 1 based.
			return lineIndex + 1, nil
		}
	}

	// Offset by 1 because we want to insert the rule after the last discovered nordvpn rule.
	return lastNordvpnRuleLine + 1, nil
}

func (i iptablesManager) InsertRule(rule FwRule) error {
	return i.executeCommand(true, rule)
}

func (i iptablesManager) DeleteRule(rule FwRule) error {
	return i.executeCommand(false, rule)
}

type PortRange struct {
	min int
	max int
}

type FwRule struct {
	chain     Chain
	ipVersion IPVersion
	params    string
	priority  RulePriority
}

func NewFwRule(chain Chain, ipVersion IPVersion, params string, priority RulePriority) FwRule {
	return FwRule{
		chain:     chain,
		ipVersion: ipVersion,
		params:    params,
		priority:  priority,
	}
}

// ToInsertAppendCommand returns the FwRule converted to insert command(-I <CHAIN> <ARGS>) or append command if index is
// -1.
func (f FwRule) ToInsertAppendCommand(index int) string {
	return fmt.Sprintf("-I %s %d %s %s", f.chain, index, f.params, f.priority.toCommentArgs())
}

func (f FwRule) ToDeleteCommand() string {
	return fmt.Sprintf("-D %s %s %s", f.chain, f.params, f.priority.toCommentArgs())
}

type FirewallManager struct {
	iptablesManager   iptablesManager
	devices           device.ListFunc // list network interfaces
	allowlistRules    []FwRule
	trafficBlockRules []FwRule
	apiAllowlistRules []FwRule
	connmark          uint32
}

func NewFirewallManager(devices device.ListFunc,
	commandRunner CommandRunner,
	connmark uint32,
	ip6TablesSupported bool,
	enabled bool) FirewallManager {
	return FirewallManager{
		iptablesManager: newIptablesManager(commandRunner, ip6TablesSupported, enabled),
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
		inputRule := NewFwRule(INPUT, Both, inputParams, TrafficBlock)
		if err := f.iptablesManager.InsertRule(inputRule); err != nil {
			return fmt.Errorf("blocking input traffic: %w", err)
		}
		f.trafficBlockRules = append(f.trafficBlockRules, inputRule)

		outputParams := fmt.Sprintf("-o %s -j DROP", iface.Name)
		outputRule := NewFwRule(OUTPUT, Both, outputParams, TrafficBlock)
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

func (f *FirewallManager) allowlistPort(iface string, protocol string, portRange PortRange) error {
	// -A INPUT -i <interface> -p <protocol> -m <protocol> --dport <port> -j ACCEPT
	// -A INPUT -i <interface> -p <protocol> -m <protocol> --sport <port> -j ACCEPT
	// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --sport <port> -j ACCEPT
	// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --dport <port> -j ACCEPT

	inputDportParams := fmt.Sprintf("-i %s -p %s -m %s --dport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	inputDportRule := NewFwRule(INPUT, Both, inputDportParams, UserAllowlist)
	if err := f.iptablesManager.InsertRule(inputDportRule); err != nil {
		return fmt.Errorf("allowlisting input dport: %w", err)
	}
	f.allowlistRules = append(f.allowlistRules, inputDportRule)

	inputSportParams := fmt.Sprintf("-i %s -p %s -m %s --sport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	inputSportRule := NewFwRule(INPUT, Both, inputSportParams, UserAllowlist)
	if err := f.iptablesManager.InsertRule(inputSportRule); err != nil {
		return fmt.Errorf("allowlisting input sport: %w", err)
	}
	f.allowlistRules = append(f.allowlistRules, inputSportRule)

	outputDportParams := fmt.Sprintf("-o %s -p %s -m %s --dport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputDportRule := NewFwRule(OUTPUT, Both, outputDportParams, UserAllowlist)
	if err := f.iptablesManager.InsertRule(outputDportRule); err != nil {
		return fmt.Errorf("allowlisting output dport: %w", err)
	}
	f.allowlistRules = append(f.allowlistRules, outputDportRule)

	outputSportParams := fmt.Sprintf("-o %s -p %s -m %s --sport %d:%d -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputSportRule := NewFwRule(OUTPUT, Both, outputSportParams, UserAllowlist)
	if err := f.iptablesManager.InsertRule(outputSportRule); err != nil {
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
			version := IPVersion(IPv4)
			if subnet.Addr().Is6() {
				version = IPv6
			}

			inputParams := fmt.Sprintf("-s %s -i %s -j ACCEPT", subnet.String(), iface.Name)
			inputRule := NewFwRule(INPUT, version, inputParams, UserAllowlist)
			if err := f.iptablesManager.InsertRule(inputRule); err != nil {
				return fmt.Errorf("adding input accept rule for subnet: %w", err)
			}
			f.allowlistRules = append(f.allowlistRules, inputRule)

			outputParams := fmt.Sprintf("-d %s -o %s -j ACCEPT", subnet.String(), iface.Name)
			outputRule := NewFwRule(OUTPUT, version, outputParams, UserAllowlist)
			if err := f.iptablesManager.InsertRule(outputRule); err != nil {
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
		inputRule := NewFwRule(INPUT, Both, inputParams, ApiAllowlistMark)
		if err := f.iptablesManager.InsertRule(inputRule); err != nil {
			return fmt.Errorf("adding api allowlist INPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, inputRule)

		outputConnmarkParams := fmt.Sprintf("-o %s -m connmark --mark %d -j ACCEPT", iface.Name, f.connmark)
		outputConnmarkRule := NewFwRule(OUTPUT, Both, outputConnmarkParams, ApiAllowlistOutputConnmark)
		if err := f.iptablesManager.InsertRule(outputConnmarkRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
		f.apiAllowlistRules = append(f.apiAllowlistRules, outputConnmarkRule)

		outputParams :=
			fmt.Sprintf("-o %s -m mark --mark %d -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
				iface.Name, f.connmark)
		outputRule := NewFwRule(OUTPUT, Both, outputParams, ApiAllowlistMark)
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
