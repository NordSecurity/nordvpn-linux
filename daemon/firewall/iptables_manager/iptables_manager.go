package iptablesmanager

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	iptablesCommand  = "iptables"
	ip6tablesCommand = "ip6tables"
)

// IpVersion determines which version of iptables command should be used, i.e iptables, ip6tables or both
type IpVersion int

const (
	IPv4 IpVersion = iota
	IPv6
	Both
)

// RulePriority determines a line in iptables where rule should be inserted. In iptables, rules are numbered in
// descending order, with rule with lower line number taking precedence over following rules. Higher priority will
// result in lower line number.
type RulePriority int

func (r RulePriority) toCommentArgs() string {
	return fmt.Sprintf("-m comment --comment nordvpn-%d", r)
}

func (r RulePriority) toComment() string {
	return fmt.Sprintf("nordvpn-%d", r)
}

type iptablesChain int

const (
	Input  iptablesChain = iota
	Output               = iota
)

func (c iptablesChain) String() string {
	switch c {
	case Input:
		return "INPUT"
	case Output:
		return "OUTPUT"
	}
	return ""
}

// CommandRunner is an abstraction over linux command execution.
type CommandRunner interface {
	RunCommand(string, string) (string, error)
}

// ExecCommandRunner is implementation of CommandRunner that facilitates commands execution with Exec calls.
// nolint:unused // Will be used once FirewallManager is integrated
type ExecCommandRunner struct {
}

// nolint:unused // Will be used once FirewallManager is integrated
func (ExecCommandRunner) RunCommand(command string, args string) (string, error) {
	// #nosec G204 -- input is properly sanitized
	output, err := exec.Command(args, strings.Split(args, " ")...).CombinedOutput()
	return string(output), err
}

// IPTablesManager manages priority and execution of firewall rules with iptables.
type IPTablesManager struct {
	ip6tablesSupported bool
	enabled            bool
	cmdRunner          CommandRunner
}

// nolint:unused // Will be used once FirewallManager is integrated
func AreIP6TablesSupported() bool {
	// #nosec G204 -- input is properly sanitized
	_, err := exec.Command(ip6tablesCommand, "-S", "-w", internal.SecondsToWaitForIptablesLock).CombinedOutput()
	return err != nil
}

func NewIPTablesManager(cmdRunner CommandRunner, enabled bool, ip6tablesSupported bool) IPTablesManager {
	return IPTablesManager{
		cmdRunner:          cmdRunner,
		enabled:            enabled,
		ip6tablesSupported: ip6tablesSupported,
	}
}

func (i IPTablesManager) executeCommand(insert bool, rule FwRule) error {
	if !i.enabled {
		return nil
	}

	command := rule.ToDeleteCommand()

	if rule.version == IPv4 || rule.version == Both {
		if insert {
			index, err := i.getRuleLine(iptablesCommand, rule.chain, rule.priority)
			if err != nil {
				return fmt.Errorf("calculating rule index: %w", err)
			}
			command = rule.ToInsertAppendCommand(index)
		}

		if _, err := i.cmdRunner.RunCommand(iptablesCommand, command); err != nil {
			return err
		}
	}

	if rule.version == IPv4 || !i.ip6tablesSupported {
		return nil
	}

	if insert {
		index, err := i.getRuleLine(ip6tablesCommand, rule.chain, rule.priority)
		if err != nil {
			return fmt.Errorf("calculating rule index: %w", err)
		}
		command = rule.ToInsertAppendCommand(index)
	}

	if _, err := i.cmdRunner.RunCommand(ip6tablesCommand, command); err != nil {
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
func (i IPTablesManager) getRuleLine(command string, chain iptablesChain, priority RulePriority) (int, error) {
	// Run command with --numeric to avoid reverse DNS lookup. This takes a long time and is unnecessary for the purpose
	// of line number calculation(we ignore everything but the 'nordvpn-<priority>' comment or the lack of thereof).
	args := "-L " + chain.String() + " --numeric" + " -w " + internal.SecondsToWaitForIptablesLock

	output, err := i.cmdRunner.RunCommand(command, args)
	if err != nil {
		return 0, fmt.Errorf("listing iptables rules: %w", err)
	}

	// Skip first two lines of output they are the chain name and table values name.
	outputLines := strings.Split(output, "\n")
	if len(outputLines) < 2 {
		return 0, fmt.Errorf("invalid output from %s %s command, expected at least two lines", command, args)
	}
	outputLines = outputLines[2:]

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

func (i IPTablesManager) InsertRule(rule FwRule) error {
	return i.executeCommand(true, rule)
}

func (i IPTablesManager) DeleteRule(rule FwRule) error {
	return i.executeCommand(false, rule)
}

type FwRule struct {
	chain    iptablesChain
	version  IpVersion
	params   string
	priority RulePriority
}

// NewFwRule returns a new representation of iptables rule.
//
// Args:
//
//	chain - chain in which rule should be inserted
//	version - version of iptables command which should be used to execute the rule, can be ipv4, ipv6 or both
//	params - rest of the params, need to be valid iptables command arguments separated by spaces
//	priority - priority at which rule should be inserted
func NewFwRule(chain iptablesChain, version IpVersion, params string, priority RulePriority) FwRule {
	return FwRule{
		chain:    chain,
		version:  version,
		params:   params,
		priority: priority,
	}
}

// ToInsertAppendCommand returns the FwRule converted to insert command(-I <CHAIN> <ARGS>) or append command if index is
// -1.
func (f FwRule) ToInsertAppendCommand(index int) string {
	return fmt.Sprintf("-I %s %d %s %s --wait "+internal.SecondsToWaitForIptablesLock, f.chain, index, f.params, f.priority.toCommentArgs())
}

func (f FwRule) ToDeleteCommand() string {
	return fmt.Sprintf("-D %s %s %s --wait "+internal.SecondsToWaitForIptablesLock, f.chain, f.params, f.priority.toCommentArgs())
}
