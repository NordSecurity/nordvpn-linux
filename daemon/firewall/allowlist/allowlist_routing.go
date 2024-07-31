// Package allowlist implements allowlist routing.
package allowlist

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/iptables"
)

const (
	RuleComment = "nordvpn_allowlist"
	iptablesCmd = "iptables"
)

type Routing interface {
	EnablePorts(ports []int, protocol string, mark string) error
	Disable() error
}

type runCommandFunc func(command string, arg ...string) ([]byte, error)

type IPTables struct {
	runCommandFunc runCommandFunc
}

// NewAllowlistRouting is a default constructor for Allowlist Routing
func NewAllowlistRouting(commandFunc runCommandFunc) *IPTables {
	return &IPTables{
		runCommandFunc: commandFunc,
	}
}

// Adds allowlist routing rules for ports
func (ipt *IPTables) EnablePorts(ports []int, protocol string, mark string) error {
	for _, portRange := range iptables.PortsToPortRanges(ports) {
		destination := fmt.Sprintf("%d:%d", portRange.Min, portRange.Max)
		if portRange.Min == portRange.Max {
			destination = fmt.Sprintf("%d", portRange.Min)
		}
		err := routePortsToIPTables(ipt.runCommandFunc, destination, protocol, mark)
		if err != nil {
			return fmt.Errorf("enabling allowlist for subnets: %w", err)
		}
	}
	return nil
}

// Deletes allowlist routing rules
func (ipt *IPTables) Disable() error {
	err := clearRouting(ipt.runCommandFunc)
	if err != nil {
		return fmt.Errorf("clearing allowlisting: %w", err)
	}
	return nil
}

func routePortsToIPTables(commandFunc runCommandFunc, port string, protocol string, mark string) error {
	if rc, err := checkRouting(commandFunc, protocol+" --dport "+port, mark); rc || err != nil {
		// already set or error happened
		return err
	}
	// iptables -t mangle -I PREROUTING -p tcp --dport 22 -j MARK --set-mark 0xe1f1 -m comment --comment "nordvpn"
	args := fmt.Sprintf(
		"-t mangle -I PREROUTING -p %s --dport %s -j MARK --set-mark %s -m comment --comment %s -w",
		protocol,
		port,
		mark,
		RuleComment,
	)
	// #nosec G204 -- input is properly sanitized
	out, err := commandFunc(iptablesCmd, strings.Fields(args)...)
	if err != nil {
		return fmt.Errorf("iptables inserting rule: %w: %s", err, string(out))
	}

	// iptables -t mangle -I OUTPUT -p tcp --sport 22 -j MARK --set-mark 0xe1f1 -m comment --comment "nordvpn"
	args = fmt.Sprintf(
		"-t mangle -I OUTPUT -p %s --sport %s -j MARK --set-mark %s -m comment --comment %s -w",
		protocol,
		port,
		mark,
		RuleComment,
	)
	// #nosec G204 -- input is properly sanitized
	out, err = commandFunc(iptablesCmd, strings.Fields(args)...)
	if err != nil {
		return fmt.Errorf("iptables inserting rule: %w: %s", err, string(out))
	}
	return nil
}

func getCleanupIPTablesRules(commandFunc runCommandFunc, chain string) error {
	args := "-t mangle -L " + chain + " -v -n --line-numbers -w"

	out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
	if err != nil {
		return fmt.Errorf("iptables listing rules: %w: %s", err, string(out))
	}
	// parse cmd output line-by-line
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		// check for comment name in rule
		if strings.Contains(string(line), RuleComment) {
			lineParts := strings.Fields(string(line[:]))
			ruleno := lineParts[0]
			args := "-t mangle -D " + chain + " %s -w"
			args = fmt.Sprintf(args, ruleno)
			// #nosec G204 -- input is properly sanitized
			out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
			if err != nil {
				return fmt.Errorf("iptables deleting rule: %w: %s", err, string(out))
			}
			return getCleanupIPTablesRules(commandFunc, chain)
		}
	}
	return nil
}

func clearRouting(commandFunc runCommandFunc) error {
	chains := []string{"PREROUTING", "OUTPUT"}
	for _, chain := range chains {
		err := getCleanupIPTablesRules(commandFunc, chain)
		if err != nil {
			return err
		}
	}
	return nil
}

// Check if rule exists
func checkRouting(commandFunc runCommandFunc, ruleType string, mark string) (bool, error) {
	args := "-t mangle -L PREROUTING -v -n -w"

	// #nosec G204 -- input is properly sanitized
	out, err := commandFunc(iptablesCmd, strings.Fields(args)...)
	if err != nil {
		return false, fmt.Errorf("iptables listing rules: %w: %s", err, string(out))
	}

	// parse cmd output line-by-line
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		// check for comment, ip and interface name in rule
		if strings.Contains(string(line), RuleComment) &&
			strings.Contains(string(line), mark) &&
			strings.Contains(string(line), ruleType) {
			return true, nil
		}
	}
	return false, nil
}
