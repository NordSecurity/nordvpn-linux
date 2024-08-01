// Package iptables implements iptables firewall agent.
package iptables

import (
	"fmt"
	"net"
	"net/netip"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	ipv4Table = "iptables"
	ipv6Table = "ip6tables"
)

const (
	accept   ruleTarget = "ACCEPT"
	drop     ruleTarget = "DROP"
	connmark ruleTarget = "CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff"
)

// ruleTarget specifies what can be passed as an argument to `-j`
type ruleTarget string

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
		}
	}
	return nil
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
						for _, input := range toInputSlice(rule.Direction) {
							for _, icmpv6Type := range defaultIcmpv6(rule.Icmpv6Types) {
								for _, target := range toTargetSlice(rule.Allow, input, rule.Marks) {
									for _, mark := range rule.Marks {
										if pRange.Min != 0 {
											for _, portFlag := range portsDirectionToPortsFlag(rule.PortsDirection) {
												newRule := generateIPTablesRule(
													input, target, iface, remoteNetwork, localNetwork, protocol, pRange,
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
												input, target, iface, remoteNetwork, localNetwork, protocol, pRange,
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

// toInputSlice returns a slice of which iptables have to be created.
// E. g. for inbound rule we create 1 rule in INPUT chain, for outbound - 1 rule in OUTPUT chain, For TwoWay - rule per both chains
func toInputSlice(direction firewall.Direction) []bool {
	switch direction {
	case firewall.Inbound:
		return []bool{true}
	case firewall.Outbound:
		return []bool{false}
	case firewall.TwoWay:
		return []bool{true, false}
	}
	return nil
}

func toTargetSlice(allowPackets bool, input bool, marks []uint32) []ruleTarget {
	var targets []ruleTarget
	if allowPackets {
		targets = append(targets, accept)
	} else {
		targets = append(targets, drop)
	}

	if input { // connmark is meant for OUTPUT chain only
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
	input bool,
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
	chain := "OUTPUT"
	remoteAddrFlag := "-d"
	localAddrFlag := "-s"
	ifaceFlag := "-o"
	if input {
		chain = "INPUT"
		remoteAddrFlag = "-s"
		localAddrFlag = "-d"
		ifaceFlag = "-i"
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
		comment = "nordvpn"
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
