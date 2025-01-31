package forwarder

import (
	"bytes"
	"fmt"
	"net/netip"
	"strings"
)

const (
	meshSrcSubnet  = "100.64.0.0/10"
	msqRuleComment = "nordvpn"
	// Used for exitnode rules that should remain in the firewall for the entire runtime
	filterRuleComment = "nordvpn-exitnode-permanent"
	// Used for exitnode rules that are removed and re-added in resetPeersTraffic, on changes in
	// meshnet state
	exitnodeTransientFilterRuleComment  = "nordvpn-exitnode-transient"
	allowlistTranisentFilterRuleComment = "nordvpn-allowlist-transient"
	// Used to ignore errors about missing rules when that is expected
	missingRuleMessage = "Bad rule (does a matching rule exist in that chain?)"

	iptablesCmd = "iptables"
)

type runCommandFunc func(command string, arg ...string) ([]byte, error)

// clearRules removes all rules in chain of table that contain the given comment. Rules will be removed in order given
// in the comments argument, i.e all rules containing the first variadic arg will be removed first, all rules containing
// the second variadic arg will be removed second, etc.
func clearRules(commandFunc runCommandFunc, chain string, table string, comments ...string) error {
	out, err := commandFunc(iptablesCmd, "-t", table, "-S", chain)
	if err != nil {
		return fmt.Errorf("listing iptables rules: %w", err)
	}
	for _, comment := range comments {
		for _, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(line, fmt.Sprintf("--comment %s", comment)) {
				continue
			}

			args := strings.Split(strings.ReplaceAll(line, "-A ", "-D "), " ")
			args = append([]string{"-t", table}, args...)
			out, err := commandFunc(iptablesCmd, args...)
			if err != nil {
				return fmt.Errorf(
					"deleting rule %s: %w, output: %s",
					line, err, string(out))
			}
		}
	}
	return nil
}

func clearMasquerading(commandFunc runCommandFunc) error {
	return clearRules(commandFunc, "POSTROUTING", "nat", msqRuleComment)
}

func enableMasquerading(peerAddress string, commandFunc runCommandFunc) error {
	// read: what comes from meshnet and goes outside meshnet should be translated
	// iptables -t nat -A POSTROUTING -s 100.64.0.0/10 ! -d 100.64.0.0/10 -m comment --comment nordvpn -j MASQUERADE
	args := fmt.Sprintf(
		"-t nat -A POSTROUTING -s %s ! -d %s -j MASQUERADE -m comment --comment %s",
		peerAddress,
		meshSrcSubnet,
		msqRuleComment,
	)
	// #nosec G204 -- input is properly sanitized
	out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
	if err != nil {
		return fmt.Errorf("iptables adding masquerading rule: %w: %s", err, string(out))
	}
	return nil
}

func checkFilteringRule(cidrIP string, commandFunc runCommandFunc) (bool, error) {
	lineNum, err := checkFilteringRulesLine([]string{cidrIP}, commandFunc)
	return lineNum != -1, err
}

// returns in which line of iptables output the rule is found or -1 if not found
func checkFilteringRulesLine(cidrIPs []string, commandFunc runCommandFunc) (int, error) {
	args := "-t filter -L FORWARD -v -n"

	out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
	if err != nil {
		return -1, fmt.Errorf("iptables listing rules: %w: %s", err, string(out))
	}
	// parse cmd output line-by-line
	for lineNum, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		// check for ip (single ip e.g. 100.64.0.50 or subnet 100.64.0.0/10) and comment in rule
		isNordvpnExitnodeRule := strings.Contains(string(line), filterRuleComment) ||
			strings.Contains(string(line), exitnodeTransientFilterRuleComment)
		if ruleContainsAllIPs(string(line), cidrIPs) && isNordvpnExitnodeRule {
			return lineNum, nil
		}
	}
	return -1, nil
}

func ruleContainsAllIPs(line string, cidrIPs []string) bool {
	for _, ip := range cidrIPs {
		if !strings.Contains(line, strings.TrimSuffix(ip, "/32")) {
			return false
		}
	}
	return true
}

func refreshPrivateSubnetsBlock(commandFunc runCommandFunc) error {
	for _, subnet := range []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("169.254.0.0/16"),
	} {
		err := modifyPeerTraffic(subnet, "-D", false, false, commandFunc)
		if err != nil && !strings.Contains(err.Error(), missingRuleMessage) {
			return fmt.Errorf("deleting private subnets block: %w", err)
		}
		if err := modifyPeerTraffic(
			subnet,
			"-I",
			false,
			false,
			commandFunc,
		); err != nil {
			return fmt.Errorf(
				"blocking traffic to '%s': %w",
				subnet,
				err,
			)
		}
	}
	return nil
}

func blockPhysicalForwarding(intfNames []string, commandFunc runCommandFunc) error {
	for _, intfName := range intfNames {
		// iptables -t filter -I FORWARD -o eth0 -j DROP -m comment --comment "<linux-app identifier>"
		args := fmt.Sprintf(
			"-t filter -I FORWARD -o %s -j DROP -m comment --comment %s",
			intfName,
			exitnodeTransientFilterRuleComment,
		)
		// #nosec G204 -- input is properly sanitized
		out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
		if err != nil {
			return fmt.Errorf("iptables inserting rule: %w: %s", err, string(out))
		}
	}
	return nil
}

func enableFiltering(commandFunc runCommandFunc) error {
	if ok, err := checkFilteringRule(meshSrcSubnet, commandFunc); ok || err != nil {
		return err
	}
	// iptables -t filter -I FORWARD 1 -s 100.64.0.0/10 -j DROP -m comment --comment "<linux-app identifier>"
	// filter out all (insert rule as 1st), except who's allowed (insert rules above)
	for _, flag := range []string{"-s", "-d"} {
		args := fmt.Sprintf(
			"-t filter -I FORWARD 1 %s %s -j DROP -m comment --comment %s",
			flag,
			meshSrcSubnet,
			filterRuleComment,
		)

		out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
		if err != nil {
			return fmt.Errorf("iptables inserting rule: %w: %s", err, string(out))
		}
	}

	args := fmt.Sprintf(
		"-t filter -I FORWARD 1 -d %s -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT -m comment --comment %s",
		meshSrcSubnet,
		filterRuleComment,
	)

	out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
	if err != nil {
		return fmt.Errorf("iptables inserting rule: %w: %s", err, string(out))
	}

	return nil
}

func removeAllowlistRules(commandFunc runCommandFunc) error {
	return clearRules(commandFunc, "FORWARD", "filter", allowlistTranisentFilterRuleComment)
}

func addAllowlistRules(commandFunc runCommandFunc, interfaceNames []string, allowlistedSubnets []string) error {
	for _, iface := range interfaceNames {
		for _, subnet := range allowlistedSubnets {
			cmd := fmt.Sprintf("-I FORWARD -d %s -o %s -m comment --comment %s -j ACCEPT",
				subnet,
				iface,
				allowlistTranisentFilterRuleComment)
			cmdArgs := strings.Split(cmd, " ")
			_, err := commandFunc(iptablesCmd, cmdArgs...)
			if err != nil {
				return fmt.Errorf("allowing local traffic for subnet %s from interface %s: %W",
					subnet, iface, err)
			}

		}
	}
	return nil
}

func resetAllowlistRules(commandFunc runCommandFunc,
	interfaceNames []string,
	killswitch bool,
	enableAllowlist bool,
	allowlistedSubnets []string) error {
	if err := removeAllowlistRules(commandFunc); err != nil {
		return fmt.Errorf("removing allowlisted subnets: %w", err)
	}

	if enableAllowlist || killswitch {
		if err := addAllowlistRules(commandFunc, interfaceNames, allowlistedSubnets); err != nil {
			return fmt.Errorf("adding allowlisted subnets: %w", err)
		}
	}

	return nil
}

type TrafficPeer struct {
	IP           netip.Prefix
	Routing      bool
	LocalNetwork bool
}

func resetForwardTraffic(
	peers []TrafficPeer,
	interfaceNames []string,
	commandFunc runCommandFunc,
	killswitch bool,
	enableAllowlist bool,
	allowlistedSubnets []string) error {
	if err := clearMasquerading(commandFunc); err != nil {
		return fmt.Errorf("clearing masquerade rules: %w", err)
	}

	if err := clearRules(commandFunc,
		"FORWARD",
		"filter",
		// allowlist rules should be deleted before exitnode block rules in order to avoid temporarily applying them to
		// unprivileged mesh peers, so order of the var args matters in this case
		allowlistTranisentFilterRuleComment,
		exitnodeTransientFilterRuleComment); err != nil {
		return fmt.Errorf("clearing exitnode forward rules: %w", err)
	}

	// Filter FORWARD rules ends here (read bottom to top!)
	// Any packets which do not pass rules below will be dropped here

	// Allow forwarding to non-private IPs for 'routing allow' peers
	// If killswitch is enabled, this would only allow forwarding through virtual interfaces,
	// because packets forwarded through physical interfaces (e.g. when not connected to VPN) were dropped before
	for _, peer := range peers {
		// Add the rule only if the "forwarding to all" single rule was not added before
		if peer.Routing && (!peer.LocalNetwork || killswitch) {
			if err := modifyPeerTraffic(peer.IP, "-I", true, true, commandFunc); err != nil {
				return fmt.Errorf(
					"adding rule while resetting peers traffic for peer %v: %w",
					peer, err,
				)
			}
		}
	}

	if enableAllowlist || killswitch {
		if err := addAllowlistRules(commandFunc, interfaceNames, allowlistedSubnets); err != nil {
			return fmt.Errorf("adding allowlist rules: %w", err)
		}
	}

	// Disallow forwarding to private IP blocks
	// This only affects 'local deny' peers, because private IPs traffic from 'local allow' peers was permitted before
	if err := refreshPrivateSubnetsBlock(commandFunc); err != nil {
		return fmt.Errorf("refreshing private subnets block while resetting peers traffic: %w", err)
	}

	// If killswitch is enabled, disallow forwarding through physical interfaces
	if killswitch {
		if err := blockPhysicalForwarding(interfaceNames, commandFunc); err != nil {
			return fmt.Errorf("adding killswitch physical forwarding block rules while resetting peers traffic: %w", err)
		}
	}

	for _, peer := range peers {
		if peer.LocalNetwork {
			if peer.Routing && !killswitch {
				// If both local and routing are allowed and killswitch is disabled, allow forwarding to all (public and
				// private) IPs using a single rule
				if err := modifyPeerTraffic(peer.IP, "-I", true, true, commandFunc); err != nil {
					return fmt.Errorf(
						"adding rule while resetting peers traffic for peer %v: %w",
						peer, err,
					)
				}
			} else {
				// Allow forwarding to private IPs for 'local allow' peers, also when killswitch is enabled
				if err := allowLocalNetworkAccess(peer.IP, "-I", commandFunc); err != nil {
					return fmt.Errorf(
						"adding rules to access local network while resetting peers traffic %v: %w",
						peer, err,
					)
				}
			}
		}
	}

	// Filter FORWARD rules starts here, read bottom to top ^^

	for _, peer := range peers {
		if peer.Routing {
			if err := enableMasquerading(peer.IP.String(), commandFunc); err != nil {
				return fmt.Errorf("enabling masquerading for peer: %w", err)
			}
		}
	}

	return nil
}

func allowLocalNetworkAccess(subnet netip.Prefix, flag string, commandFunc runCommandFunc) error {
	// TODO: Probably this should add rules only for actually connected local networks (both physical and virtual, for
	// example Docker), instead of all private ranges, because packets destined to private IPs from non-connected local
	// networks will be forwarded through the default gateway and we do not want that, because this should be controlled
	// by peer.Routing.
	for _, localSubnet := range []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("169.254.0.0/16"),
	} {
		args := fmt.Sprintf(
			"-t filter %s FORWARD -s %s -d %s -j ACCEPT -m comment --comment %s",
			flag,
			subnet.String(),
			localSubnet.String(),
			exitnodeTransientFilterRuleComment,
		)
		// #nosec G204 -- input is properly sanitized
		out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
		if err != nil {
			return fmt.Errorf("iptables modifying rule: %w: %s", err, string(out))
		}
	}

	return nil
}

func modifyPeerTraffic(subnet netip.Prefix,
	flag string,
	source bool,
	allow bool,
	commandFunc runCommandFunc,
) error {
	sourceFlag := fmt.Sprintf("-s %s -d", meshSrcSubnet)
	if source {
		sourceFlag = "-s"
	}
	acceptFlag := "DROP"
	if allow {
		acceptFlag = "ACCEPT"
	}

	// iptables -t filter -I FORWARD -s 100.64.0.159 -j ACCEPT -m comment --comment "<linux-app identifier>"
	args := fmt.Sprintf(
		"-t filter %s FORWARD %s %s -j %s -m comment --comment %s",
		flag,
		sourceFlag,
		subnet.String(),
		acceptFlag,
		exitnodeTransientFilterRuleComment,
	)

	out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)
	if err != nil {
		return fmt.Errorf("iptables modifying rule: %w: %s", err, string(out))
	}
	return nil
}

// clearFiltering drops all the rules in the FORWARD chain containing
// a comment
func clearFiltering(commandFunc runCommandFunc) error {
	return clearRules(commandFunc, "FORWARD", "filter", filterRuleComment, exitnodeTransientFilterRuleComment)
}
