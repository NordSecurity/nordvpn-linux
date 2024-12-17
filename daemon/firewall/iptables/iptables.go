// Package iptables implements iptables firewall agent.
package iptables

import (
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
)

type PortRange struct {
	Min int
	Max int
}

// @TODO upgrade to netfilter library. for now we use both ipv4, ipv6 because we disable ipv6
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
	flag := "-D"
	errStr := "deleting"
	if add {
		flag = "-I"
		errStr = "adding"
	}
	module, stateFlag := ipt.getStateModule(rule)
	allRules := ruleToIPTables(rule, module, stateFlag, ipt.chainPrefix)

	for _, iptableVersion := range ipt.supportedIPTables {
		ipTablesRules, ok := allRules[iptableVersion]
		if !ok {
			continue
		}
		for _, ipTableRule := range ipTablesRules {
			if !rule.Allow {
				prefix := fmt.Sprintf("-j NFLOG --nflog-prefix \"LOG-pre-%s\"", rule.Name)
				log.Println(internal.DebugPrefix, "[iptables-debug], add rule: ", prefix)
				logRule := strings.Replace(ipTableRule, "-j DROP", prefix, -1)
				args := fmt.Sprintf("%s %s -w"+internal.SecondsToWaitForIptablesLock, flag, logRule)
				out, err := exec.Command(iptableVersion, strings.Split(args, " ")...).CombinedOutput()
				if err != nil {
					log.Printf(internal.ErrorPrefix+" [iptables-debug]"+" failed to add rule: %ss: %s", err, string(out))
				}
			}
			// -w does not accept arguments on older iptables versions
			args := fmt.Sprintf("%s %s -w "+internal.SecondsToWaitForIptablesLock, flag, ipTableRule)
			// #nosec G204 -- input is properly sanitized
			out, err := exec.Command(iptableVersion, strings.Split(args, " ")...).CombinedOutput()
			if err != nil {
				if flag == "-D" && strings.Contains(string(out), "does a matching rule exist in that chain") {
					return nil
				}
				return fmt.Errorf("%s %s rule '%s': %w: %s", errStr, iptableVersion, ipTableRule, err, string(out))
			}

			if !rule.Allow {
				prefix := fmt.Sprintf("-j NFLOG --nflog-prefix \"LOG-post-%s\"", rule.Name)
				log.Println(internal.DebugPrefix, "[iptables-debug], add rule: ", prefix)
				logRule := strings.Replace(ipTableRule, "-j DROP", prefix, -1)
				args := fmt.Sprintf("%s %s -w"+internal.SecondsToWaitForIptablesLock, flag, logRule)
				out, err := exec.Command(iptableVersion, strings.Split(args, " ")...).CombinedOutput()
				if err != nil {
					log.Printf(internal.ErrorPrefix+"[iptables-debug]"+" failed to add rule: %s: %s", err, string(out))
				}
			}
		}
	}
	return nil
}

func generateFlushRules(rules string) []string {
	re := regexp.MustCompile(fmt.Sprintf(`--comment\s+%s(?:\s|$)`, regexp.QuoteMeta(defaultComment)))
	flushRules := []string{}
	for _, rule := range strings.Split(rules, "\n") {
		if re.MatchString(rule) {
			newRule := strings.Replace(rule, "-A", "-D", 1)
			flushRules = append(flushRules, newRule)
		}
	}

	return flushRules
}

func (ipt *IPTables) Flush() error {
	var finalErr error = nil
	for _, iptableVersion := range ipt.supportedIPTables {
		out, err := exec.Command(iptableVersion, "-S").CombinedOutput()
		if err != nil {
			return fmt.Errorf("listing rules: %w", err)
		}

		rules := string(out)
		for _, rule := range generateFlushRules(rules) {
			err := exec.Command(iptableVersion, strings.Split(rule, " ")...).Run()
			if err != nil {
				log.Printf("%s failed to delete rule %s: %s", internal.ErrorPrefix, rule, err)
				finalErr = fmt.Errorf("failed to delete all rules")
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
	// n-fold Cartesian Product, where n stands for the level of for loop nesting
	for _, iface := range rule.Interfaces {
		for _, remoteNetwork := range rule.RemoteNetworks {
			for _, localNetwork := range rule.LocalNetworks {
				for _, pRange := range PortsToPortRanges(rule.Ports) {
					for _, protocol := range rule.Protocols {
						for _, chain := range toChainSlice(rule.Direction) {
							for _, icmpv6Type := range defaultIcmpv6(rule.Icmpv6Types) {
								for _, target := range toTargetSlice(rule.Allow, chain, rule.Marks) {
									for _, mark := range rule.Marks {
										if pRange.Min != 0 {
											for _, portFlag := range portsDirectionToPortsFlag(rule.PortsDirection) {
												newRule := generateIPTablesRule(
													chain, target, iface, remoteNetwork, localNetwork, protocol, pRange,
													module, stateFlag, rule.ConnectionStates, chainPrefix, portFlag,
													icmpv6Type, rule.HopLimit, nil, nil,
													rule.Comment, mark,
												)
												if rule.Ipv6Only || remoteNetwork.Addr().Is6() || localNetwork.Addr().Is6() {
													ipv6TableRules = append(ipv6TableRules, newRule)
												} else if remoteNetwork.Addr().Is4() || localNetwork.Addr().Is4() {
													ipv4TableRules = append(ipv4TableRules, newRule)
												} else {
													ipv6TableRules = append(ipv6TableRules, newRule)
													ipv4TableRules = append(ipv4TableRules, newRule)
												}
											}
										} else {
											newRule := generateIPTablesRule(
												chain, target, iface, remoteNetwork, localNetwork, protocol, pRange,
												module, stateFlag, rule.ConnectionStates, chainPrefix, "",
												icmpv6Type, rule.HopLimit,
												rule.SourcePorts, rule.DestinationPorts,
												rule.Comment, mark,
											)
											if rule.Ipv6Only || remoteNetwork.Addr().Is6() || localNetwork.Addr().Is6() {
												ipv6TableRules = append(ipv6TableRules, newRule)
											} else if remoteNetwork.Addr().Is4() || localNetwork.Addr().Is4() {
												ipv4TableRules = append(ipv4TableRules, newRule)
											} else {
												ipv6TableRules = append(ipv6TableRules, newRule)
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
	}
	return map[string][]string{ipv4Table: ipv4TableRules, ipv6Table: ipv6TableRules}
}

func defaultIcmpv6(icmp6Types []int) []int {
	if len(icmp6Types) > 0 {
		return icmp6Types
	}
	return []int{0}
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
func toChainSlice(direction firewall.Direction) []ruleChain {
	switch direction {
	case firewall.Inbound:
		return []ruleChain{chainInput}
	case firewall.Outbound:
		return []ruleChain{chainOutput}
	case firewall.TwoWay:
		return []ruleChain{chainInput, chainOutput}
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

	if chain != chainOutput { // connmark is meant for OUTPUT chain only
		return targets
	}

	if len(marks) > 0 && marks[0] != 0 {
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
	icmpv6Type int,
	hopLimit uint8,
	sports []int,
	dports []int,
	comment string,
	mark uint32,
) string {
	var chain, remoteAddrFlag, localAddrFlag, ifaceFlag string

	switch direction {
	case chainInput:
		chain = "INPUT"
		remoteAddrFlag = "-s"
		localAddrFlag = "-d"
		ifaceFlag = "-i"
	case chainOutput:
		chain = "OUTPUT"
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

	if icmpv6Type > 0 {
		rule += fmt.Sprintf(" --icmpv6-type %d", icmpv6Type)
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

	return rule + acceptComment + jump + string(target)
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
