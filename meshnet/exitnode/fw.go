package exitnode

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
	transientFilterRuleComment = "nordvpn-exitnode-transient"
	// Used to ignore errors about missing rules when that is expected
	missingRuleMessage = "Bad rule (does a matching rule exist in that chain?)"

	iptablesCmd = "iptables"
)

type runCommandFunc func(command string, arg ...string) ([]byte, error)

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

func clearMasquerading(commandFunc runCommandFunc) error {
	args := "-t nat -S POSTROUTING"
	out, err := commandFunc(iptablesCmd, strings.Split(args, " ")...)

	if err != nil {
		return fmt.Errorf("iptables listing rules: %w: %s", err, string(out))
	}

	for _, line := range bytes.Split(out, []byte{'\n'}) {
		lineString := string(line)
		if !strings.Contains(lineString, msqRuleComment) {
			continue
		}

		args := strings.Split(strings.ReplaceAll(lineString, "-A", "-D"), " ")
		args = append([]string{"-t", "nat"}, args...)
		_, err := commandFunc(iptablesCmd, args...)

		if err != nil {
			return fmt.Errorf("iptables deleting rule: %w: %s", err, lineString)
		}
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
			strings.Contains(string(line), transientFilterRuleComment)
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

type TrafficPeer struct {
	IP           netip.Prefix
	Routing      bool
	LocalNetwork bool
}

func clearExitnodeForwardRules(commandFunc runCommandFunc) error {
	output, err := commandFunc("iptables", "-S", "FORWARD")
	if err != nil {
		return fmt.Errorf("listing iptables: %w", err)
	}

	rules := strings.Split(string(output), "\n")
	for _, rule := range rules {
		if !strings.Contains(rule, transientFilterRuleComment) {
			continue
		}

		deleteCommand := strings.Replace(rule, "-A", "-D", -1)

		if _, err := commandFunc("iptables", strings.Split(deleteCommand, " ")...); err != nil {
			return fmt.Errorf("deleting iptables rule: %w", err)
		}
	}

	return nil
}

func resetPeersTraffic(peers []TrafficPeer, interfaceNames []string, commandFunc runCommandFunc) error {
	if err := clearMasquerading(commandFunc); err != nil {
		return fmt.Errorf("clearing masquerade rules: %w", err)
	}

	if err := clearExitnodeForwardRules(commandFunc); err != nil {
		return fmt.Errorf("clearing exitnode forward rules: %w", err)
	}

	for _, peer := range peers {
		if peer.Routing && !peer.LocalNetwork {
			if err := modifyPeerTraffic(peer.IP, "-I", true, true, commandFunc); err != nil {
				return fmt.Errorf(
					"adding rule while resetting peers traffic for peer %v: %w",
					peer, err,
				)
			}
		}
	}

	if err := refreshPrivateSubnetsBlock(commandFunc); err != nil {
		return fmt.Errorf("refreshing private subnets block while resetting peers traffic: %w", err)
	}

	for _, peer := range peers {
		if peer.LocalNetwork {
			if peer.Routing {
				if err := modifyPeerTraffic(peer.IP, "-I", true, true, commandFunc); err != nil {
					return fmt.Errorf(
						"adding rule while resetting peers traffic for peer %v: %w",
						peer, err,
					)
				}
			} else {
				if err := allowOnlyLocalNetworkAccess(peer.IP, "-I", commandFunc); err != nil {
					return fmt.Errorf(
						"adding rules to access local network while resetting peers traffic %v: %w",
						peer, err,
					)
				}
			}
		}
	}

	for _, peer := range peers {
		if peer.Routing {
			if err := enableMasquerading(peer.IP.String(), commandFunc); err != nil {
				return fmt.Errorf("enabling masquerading for peer: %w", err)
			}
		}
	}

	return nil
}

func allowOnlyLocalNetworkAccess(subnet netip.Prefix, flag string, commandFunc runCommandFunc) error {
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
			transientFilterRuleComment,
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
		transientFilterRuleComment,
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
	out, err := commandFunc(iptablesCmd, "-S")
	if err != nil {
		return fmt.Errorf("listing iptables rules: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		isFilterRule := strings.Contains(line, fmt.Sprintf("--comment %s", filterRuleComment))
		isTransientFilterRule := strings.Contains(line, fmt.Sprintf("--comment %s", transientFilterRuleComment))
		if !strings.Contains(line, "FORWARD") ||
			!(isFilterRule || isTransientFilterRule) {
			continue
		}

		out, err := commandFunc(iptablesCmd, strings.Split(strings.ReplaceAll(line, "-A ", "-D "), " ")...)
		if err != nil {
			return fmt.Errorf(
				"deleting FORWARD rule %s: %w: %s",
				line, err, string(out))
		}
	}
	return nil
}
