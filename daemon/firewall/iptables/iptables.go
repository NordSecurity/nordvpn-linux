// Package iptables implements iptables firewall agent.
package iptables

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	ipv4Table      = "iptables"
	ipv6Table      = "ip6tables"
	defaultComment = "nordvpn"
)

var usedIPTables = [...]string{"mangle", "filter"}

const (
	accept   ruleTarget = "ACCEPT"
	drop     ruleTarget = "DROP"
	connmark ruleTarget = "CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff"
)

// ruleTarget specifies what can be passed as an argument to `-j`
type ruleTarget string

type ruleChain int

const (
	chainInput ruleChain = iota
	chainOutput
	chainForward
	chainPrerouting
	chainPostrouting
)

type PortRange struct {
	Min int
	Max int
}

// @TODO upgrade to netfilter library.
// IPTables handles all firewall changes with iptables
type IPTables struct {
	stateModule       string
	stateFlag         string
	chainPrefix       string
	originalInput     map[string]*bool
	originalOutput    map[string]*bool
	supportedIPTables []string
	sync.Mutex
}

// New is a default constructor for IPTables firewall
func New(stateModule string, stateFlag string, chainPrefix string, supportedIPTables []string) *IPTables {
	originalInput := make(map[string]*bool)
	originalOutput := make(map[string]*bool)
	return &IPTables{
		stateModule:       stateModule,
		stateFlag:         stateFlag,
		chainPrefix:       chainPrefix,
		originalInput:     originalInput,
		originalOutput:    originalOutput,
		supportedIPTables: supportedIPTables,
	}
}

func (ipt *IPTables) Add(rule firewall.Rule) error {
	ipt.Lock()
	defer ipt.Unlock()
	return ipt.applyRule(rule, true)
}

func (ipt *IPTables) getStateModule(rule firewall.Rule) (module string, flag string) {
	if rule.ConnectionStates.States != nil {
		module = ipt.stateModule
		flag = ipt.stateFlag
	}
	return module, flag
}

func (ipt *IPTables) Delete(rule firewall.Rule) error {
	ipt.Lock()
	defer ipt.Unlock()
	return ipt.applyRule(rule, false)
}

func (ipt *IPTables) applyRule(rule firewall.Rule, add bool) error {
	tableFlag := "filter"
	flag := "-D"
	errStr := "deleting"
	if add {
		flag = "-I"
		errStr = "adding"
	}
	module, stateFlag := ipt.getStateModule(rule)
	allRules := ruleToIPTables(rule, module, stateFlag, ipt.chainPrefix)
	if rule.Physical {
		tableFlag = "mangle"
	}

	for _, iptableVersion := range ipt.supportedIPTables {
		ipTablesRules, ok := allRules[iptableVersion]
		if !ok {
			continue
		}
		for _, ipTableRule := range ipTablesRules {
			// -w does not accept arguments on older iptables versions
			args := fmt.Sprintf("-t %s %s %s -w "+internal.SecondsToWaitForIptablesLock, tableFlag, flag, ipTableRule)
			// #nosec G204 -- input is properly sanitized
			out, err := exec.Command(iptableVersion, strings.Split(args, " ")...).CombinedOutput()
			if err != nil {
				if flag == "-D" && strings.Contains(string(out), "does a matching rule exist in that chain") {
					return nil
				}
				return fmt.Errorf("%s %s rule '%s': %w: %s", errStr, iptableVersion, ipTableRule, err, string(out))
			}
		}
	}
	return nil
}

func generateFlushRules(rules string, table string) []string {
	re := regexp.MustCompile(fmt.Sprintf(`--comment\s+%s(?:\s|$)`, regexp.QuoteMeta(defaultComment)))
	flushRules := []string{}
	for _, rule := range strings.Split(rules, "\n") {
		if re.MatchString(rule) {
			newRule := fmt.Sprintf("-t %s %s", table, strings.Replace(rule, "-A", "-D", 1))
			flushRules = append(flushRules, newRule)
		}
	}

	return flushRules
}

func (ipt *IPTables) Flush() error {
	var finalErr error = nil
	for _, table := range usedIPTables {
		for _, iptableVersion := range ipt.supportedIPTables {
			out, err := getRuleOutput(iptableVersion, table)
			if err != nil {
				return err
			}
			rules := string(out)
			for _, rule := range generateFlushRules(rules, table) {
				err := exec.Command(iptableVersion, strings.Split(rule, " ")...).Run()
				if err != nil {
					log.Printf("%s failed to delete rule %s: %s", internal.ErrorPrefix, rule, err)
					finalErr = fmt.Errorf("failed to delete all rules")
				}
			}
		}
	}

	return finalErr
}

// FilterSupportedIPTables filter supported versions based on what exists in the system
func FilterSupportedIPTables(supportedIPTables []string) []string {
	var supported []string
	for _, cmd := range supportedIPTables {
		// #nosec G204 -- input is properly sanitized
		_, err := exec.Command(cmd, "-S", "-w", internal.SecondsToWaitForIptablesLock).CombinedOutput()
		if err != nil {
			continue
		}
		supported = append(supported, cmd)
	}
	return supported
}

func trimPrefixes(str string, prefixes ...string) string {
	for _, prefix := range prefixes {
		str = strings.TrimSpace(strings.TrimPrefix(str, prefix))
	}
	return str
}

func portsDirectionToPortsFlag(direction firewall.PortsDirection) []string {
	switch direction {
	case firewall.SourceAndDestination:
		return []string{"--sport", "--dport"}
	case firewall.Destination:
		return []string{"--dport"}
	case firewall.Source:
		return []string{"--sport"}
	default:
		return []string{"--sport", "--dport"}
	}
}

// This is here for historical reasons. Please don't judge us
func ruleToIPTables(rule firewall.Rule, module string, stateFlag string, chainPrefix string) map[string][]string {
	// fill nil fields with elements of nil values, so each slice has at least one element and at least 1 rule is generated
	rule = generateNonEmptyRule(rule)
	var ipv4TableRules []string
	var ipv6TableRules []string
	nameComment := rule.Name
	if len(rule.SimplifiedName) > 0 {
		nameComment = rule.SimplifiedName
	}
	// n-fold Cartesian Product, where n stands for the level of for loop nesting
	for _, iface := range rule.Interfaces {
		for _, remoteNetwork := range rule.RemoteNetworks {
			for _, localNetwork := range rule.LocalNetworks {
				for _, pRange := range PortsToPortRanges(rule.Ports) {
					for _, protocol := range rule.Protocols {
						for _, chain := range toChainSlice(rule.Direction, rule.Physical) {
							for _, target := range toTargetSlice(rule.Allow, chain, rule.Marks) {
								for _, mark := range rule.Marks {
									if pRange.Min != 0 {
										for _, portFlag := range portsDirectionToPortsFlag(rule.PortsDirection) {
											newRule := generateIPTablesRule(
												chain, target, iface, remoteNetwork, localNetwork, protocol, pRange,
												module, stateFlag, rule.ConnectionStates, chainPrefix, portFlag,
												rule.HopLimit, nil, nil, rule.Comment, mark, nameComment,
											)
											if rule.Ipv6Only {
												// We should have just our block rule here
												ipv6TableRules = append(ipv6TableRules, newRule)
											} else {
												ipv4TableRules = append(ipv4TableRules, newRule)
											}
										}
									} else {
										newRule := generateIPTablesRule(
											chain, target, iface, remoteNetwork, localNetwork, protocol, pRange,
											module, stateFlag, rule.ConnectionStates, chainPrefix, "",
											rule.HopLimit, rule.SourcePorts, rule.DestinationPorts,
											rule.Comment, mark, nameComment,
										)
										if rule.Ipv6Only {
											// We should have just our block rule here
											ipv6TableRules = append(ipv6TableRules, newRule)
										} else {
											ipv4TableRules = append(ipv4TableRules, newRule)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return map[string][]string{ipv4Table: ipv4TableRules, ipv6Table: ipv6TableRules}
}

// PortsToPortRanges groups ports into ranges
func PortsToPortRanges(ports []int) []PortRange {
	if len(ports) == 0 {
		return nil
	}
	sort.Ints(ports)

	var ranges []PortRange
	pPort := ports[0]
	r := PortRange{Min: pPort, Max: pPort}
	for i, port := range ports[1:] {
		if port == ports[i]+1 {
			r.Max = port
			continue
		}
		ranges = append(ranges, r)
		r = PortRange{Min: port, Max: port}
	}
	return append(ranges, r)
}

// toChainSlice returns a slice of which iptables have to be created.
// E. g. for inbound rule we create 1 rule in INPUT chain, for outbound - 1 rule in OUTPUT chain, For TwoWay - rule per both chains
func toChainSlice(direction firewall.Direction, physical bool) []ruleChain {
	incomingChain := chainInput
	outgoingChain := chainOutput
	if physical {
		incomingChain = chainPrerouting
		outgoingChain = chainPostrouting
	}
	switch direction {
	case firewall.Inbound:
		return []ruleChain{incomingChain}
	case firewall.Outbound:
		return []ruleChain{outgoingChain}
	case firewall.TwoWay:
		return []ruleChain{incomingChain, outgoingChain}
	case firewall.Forward:
		return []ruleChain{chainForward}
	}
	return nil
}

func toTargetSlice(allowPackets bool, chain ruleChain, marks []uint32) []ruleTarget {
	var targets []ruleTarget
	if allowPackets {
		targets = append(targets, accept)
	} else {
		targets = append(targets, drop)
	}
	// connmark is meant for OUTPUT and POSTROUTING chains only
	if (chain == chainOutput || chain == chainPostrouting) && (len(marks) > 0 && marks[0] != 0) {
		targets = append(targets, connmark)
	}
	return targets
}

// generateNonEmptyRule fills nil fields with values of one element with nil value
func generateNonEmptyRule(rule firewall.Rule) firewall.Rule {
	if rule.RemoteNetworks == nil {
		rule.RemoteNetworks = append(rule.RemoteNetworks, netip.Prefix{})
	}
	if rule.LocalNetworks == nil {
		rule.LocalNetworks = append(rule.LocalNetworks, netip.Prefix{})
	}
	if rule.Protocols == nil {
		rule.Protocols = append(rule.Protocols, "")
	}
	if rule.Interfaces == nil {
		rule.Interfaces = append(rule.Interfaces, net.Interface{})
	}
	if rule.Ports == nil {
		rule.Ports = append(rule.Ports, 0)
	}
	if rule.Marks == nil {
		rule.Marks = append(rule.Marks, 0)
	}
	return rule
}

// generateIPTablesRule converts input fields to a single IPTables rule string
func generateIPTablesRule(
	direction ruleChain,
	target ruleTarget,
	iface net.Interface,
	remoteNetwork netip.Prefix,
	localNetwork netip.Prefix,
	protocol string,
	portRange PortRange,
	module string,
	stateFlag string,
	states firewall.ConnectionStates,
	chainPrefix string,
	portFlag string,
	hopLimit uint8,
	sports []int,
	dports []int,
	comment string,
	mark uint32,
	nameComment string,
) string {
	var chain, remoteAddrFlag, localAddrFlag, ifaceFlag string

	switch direction {
	case chainInput, chainPrerouting:
		if direction == chainPrerouting {
			chain = "PREROUTING"
		} else {
			chain = "INPUT"
		}
		remoteAddrFlag = "-s"
		localAddrFlag = "-d"
		ifaceFlag = "-i"
	case chainOutput, chainPostrouting:
		if direction == chainPostrouting {
			chain = "POSTROUTING"
		} else {
			chain = "OUTPUT"
		}
		remoteAddrFlag = "-d"
		localAddrFlag = "-s"
		ifaceFlag = "-o"
	case chainForward:
		chain = "FORWARD"
		remoteAddrFlag = "-d"
		localAddrFlag = "-s"
		ifaceFlag = "-o"
	}

	rule := chainPrefix + chain
	if iface.Name != "" {
		rule += " " + ifaceFlag + " " + iface.Name
	}
	if remoteNetwork != (netip.Prefix{}) {
		rule += " " + remoteAddrFlag + " " + remoteNetwork.String()
	}
	if localNetwork != (netip.Prefix{}) {
		rule += " " + localAddrFlag + " " + localNetwork.String()
	}
	if protocol != "" {
		rule += " -p " + protocol
	}
	if mark != 0 {
		if target == connmark {
			rule += fmt.Sprintf(" -m mark --mark %#x", mark)
		} else {
			rule += fmt.Sprintf(" -m connmark --mark %#x", mark)
		}
	}
	if portRange.Min != 0 && portFlag != "" {
		rule += fmt.Sprintf(" %s %d:%d", portFlag, portRange.Min, portRange.Max)
	} else {
		if len(sports) > 0 {
			rule += fmt.Sprintf(" %s %s", "--sport", strings.Join(internal.IntsToStrings(sports), ","))
		}
		if len(dports) > 0 {
			rule += fmt.Sprintf(" %s %s", "--dport", strings.Join(internal.IntsToStrings(dports), ","))
		}
	}

	if module != "" {
		rule += " -m " + module
	}
	if stateFlag != "" && states.States != nil {
		var statesStr []string
		for _, state := range states.States {
			statesStr = append(statesStr, connectionStateToString(state))
		}
		rule += " " + stateFlag + " " + strings.Join(statesStr, ",")
		if states.SrcAddr.IsValid() && !states.SrcAddr.IsUnspecified() {
			rule += fmt.Sprintf(" --ctorigsrc %s", states.SrcAddr.String())
		}
	}

	if hopLimit > 0 {
		rule += fmt.Sprintf(" -m hl --hl-eq %d", hopLimit)
	}

	jump := " -j "

	if comment == "" {
		comment = defaultComment
	}

	var acceptComment string
	if len(comment) > 0 {
		acceptComment = " -m comment --comment " + comment
	}
	var ruleNameComment string
	if len(nameComment) > 0 {
		ruleNameComment = " -m comment --comment " + nameComment
	}
	return rule + acceptComment + ruleNameComment + jump + string(target)
}

// connectionStateToString converts package connection state to string
func connectionStateToString(state firewall.ConnectionState) string {
	switch state {
	case firewall.Related:
		return "RELATED"
	case firewall.Established:
		return "ESTABLISHED"
	case firewall.New:
		return "NEW"
	}
	return ""
}

func (ipt *IPTables) GetActiveRules() ([]string, error) {
	ipt.Lock()
	defer ipt.Unlock()
	re, err := regexp.Compile(`--comment \S*`)
	if err != nil {
		return nil, fmt.Errorf("compiling regexp: %w", err)
	}
	rulesMap := make(map[string]bool)
	for _, table := range usedIPTables {
		for _, iptableVersion := range ipt.supportedIPTables {
			rules, err := getRuleOutput(iptableVersion, table)
			if err != nil {
				return nil, err
			}
			for _, line := range bytes.Split(rules, []byte{'\n'}) {
				if len(line) == 0 {
					continue
				}
				// check for comment name in rule
				fmt.Println("Looking at line: " + string(line))
				matches := re.FindAll(line, -1)
				var stringMatches []string
				for _, match := range matches {
					fmt.Println("found match" + string(match))
					stringMatches = append(stringMatches, string(match))
				}
				// first comment is the usual nordvpn, second is the name of the rule
				// if there is only one, we skip as we don't have the rule name comment
				if len(stringMatches) > 1 {
					formattedRuleName := strings.Split(string(matches[1]), " ")[1]
					rulesMap[formattedRuleName] = true
				}
			}
		}
	}
	var ruleList []string
	for k := range rulesMap{
		ruleList = append(ruleList, k)
	}
	return ruleList, nil
}

func getRuleOutput(iptableVersion string, table string) ([]byte, error){
	out, err := exec.Command(iptableVersion, "-t", table, "-S", "-w", internal.SecondsToWaitForIptablesLock).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("listing rules: %w", err)
	}
	return out, nil
}