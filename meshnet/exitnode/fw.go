package exitnode

import (
	"bytes"
	"fmt"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
)

const (
	meshSrcSubnet     = "100.64.0.0/10"
	msqRuleComment    = "nordvpn"
	filterRuleComment = "nordvpn"
	// Used to ignore errors about missing rules when that is expected
	missingRuleMessage = "Bad rule (does a matching rule exist in that chain?)"
	ovpnInterfaceName  = "nordtun"
)

type runCommandFunc func(command string, args ...string) ([]byte, error)

func RunCommandFuncExec(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).CombinedOutput()
}

type MasqueradeSetter interface {
	EnableMasquerading(intfNames []string) error
	ClearMasquerading(intfNames []string) error
}

func GetMasqueradeSetter(technology config.Technology) MasqueradeSetter {
	var masqueradeSetter MasqueradeSetter
	if technology == config.Technology_OPENVPN {
		masqueradeSetter = NewOpenVPNMasqueradeSetter()
	} else {
		masqueradeSetter = NewNordlynxMasqueradeSetter()
	}

	return masqueradeSetter
}

type NordlynxMasqueradeSetter struct {
	cmdFunc runCommandFunc
}

func NewNordlynxMasqueradeSetter() *NordlynxMasqueradeSetter {
	return &NordlynxMasqueradeSetter{
		cmdFunc: RunCommandFuncExec,
	}
}

func (nm *NordlynxMasqueradeSetter) EnableMasquerading(intfNames []string) error {
	if err := enableMasquerading(intfNames, nm.cmdFunc); err != nil {
		return fmt.Errorf("enabling Nordlynx masquerading: %w", err)
	}

	return nil
}

func (nm *NordlynxMasqueradeSetter) ClearMasquerading(intfNames []string) error {
	if err := clearMasquerading(intfNames, nm.cmdFunc); err != nil {
		return fmt.Errorf("clearing Nordlynx masquerading: %w", err)
	}

	return nil
}

type OpenVPNMasqueradeSetter struct {
	cmdFunc runCommandFunc
}

func NewOpenVPNMasqueradeSetter() *OpenVPNMasqueradeSetter {
	return &OpenVPNMasqueradeSetter{
		cmdFunc: RunCommandFuncExec,
	}
}

func (om *OpenVPNMasqueradeSetter) EnableMasquerading(intfNames []string) error {
	// iptables -t nat -A POSTROUTING -s 100.64.0.0/10 -o nordtun -j MASQUERADE -m comment --comment "nordvpn"
	cmd := "iptables"
	if output, err := om.cmdFunc(cmd, "-t", "nat", "-A", "POSTROUTING", "-s", meshSrcSubnet, "-o",
		ovpnInterfaceName, "-j", "MASQUERADE", "-m", "comment", "--comment",
		msqRuleComment); err != nil {
		return fmt.Errorf("enabling OpenVPN masquerading: %w: %s", err, string(output))
	}

	if err := enableMasquerading(intfNames, om.cmdFunc); err != nil {
		return fmt.Errorf("enabling OpenVPN masquerading: %w", err)
	}

	return nil
}

func (om *OpenVPNMasqueradeSetter) ClearMasquerading(intfNames []string) error {
	// iptables -t nat -A POSTROUTING -s 100.64.0.0/10 -o nordtun -j MASQUERADE -m comment --comment "nordvpn"
	cmd := "iptables"
	if output, err := om.cmdFunc(cmd, "-t", "nat", "-D", "POSTROUTING", "-s", meshSrcSubnet, "-o",
		ovpnInterfaceName, "-j", "MASQUERADE", "-m", "comment", "--comment",
		msqRuleComment); err != nil {
		return fmt.Errorf("clearing OpenVPN masquerading: %w: %s", err, string(output))
	}

	if err := clearMasquerading(intfNames, om.cmdFunc); err != nil {
		return fmt.Errorf("clearing OpenVPN masquerading: %w", err)
	}

	return nil
}

func enableMasquerading(intfNames []string, commandFunc runCommandFunc) error {
	for _, intfName := range intfNames {
		if rc, err := checkMasquerading(intfName, commandFunc); rc || err != nil {
			// already set or error happened
			return err
		}

		// iptables -t nat -A POSTROUTING -s 100.64.0.0/10 -o eth0 -j MASQUERADE -m comment --comment "nordvpn"
		cmd := "iptables"
		args := fmt.Sprintf(
			"-t nat -A POSTROUTING -s %s -o %s -j MASQUERADE -m comment --comment %s",
			meshSrcSubnet,
			intfName,
			msqRuleComment,
		)
		// #nosec G204 -- input is properly sanitized
		out, err := commandFunc(cmd, strings.Split(args, " ")...)
		if err != nil {
			return fmt.Errorf("iptables adding masquerading rule: %w: %s", err, string(out))
		}
	}
	return nil
}

func clearMasquerading(intfNames []string, commandFunc runCommandFunc) error {
	for _, intfName := range intfNames {
		// remove all rules with comment
		for {
			found := false
			cmd := "iptables"
			args := "-t nat -L POSTROUTING -v -n --line-numbers"

			out, err := commandFunc(cmd, strings.Split(args, " ")...)
			if err != nil {
				return fmt.Errorf("iptables listing rules: %w: %s", err, string(out))
			}
			// parse cmd output line-by-line
			for _, line := range bytes.Split(out, []byte{'\n'}) {
				if len(line) == 0 {
					continue
				}
				// check for comment and interface name in rule
				if strings.Contains(string(line), msqRuleComment) &&
					strings.Contains(string(line), intfName) {
					lineParts := strings.Fields(string(line[:]))
					ruleno := lineParts[0]
					cmd := "iptables"
					args := "-t nat -D POSTROUTING %s"
					args = fmt.Sprintf(args, ruleno)
					// #nosec G204 -- input is properly sanitized
					out, err := commandFunc(cmd, strings.Split(args, " ")...)
					if err != nil {
						return fmt.Errorf("iptables deleting rule: %w: %s", err, string(out))
					}
					found = true
					break
				}
			}
			if !found { // repeat until not found
				break
			}
		}
	}
	return nil
}

func checkMasquerading(intfName string, commandFunc runCommandFunc) (bool, error) {
	cmd := "iptables"
	args := "-t nat -L POSTROUTING -v -n"
	// #nosec G204 -- input is properly sanitized
	out, err := commandFunc(cmd, strings.Split(args, " ")...)
	if err != nil {
		return false, fmt.Errorf("iptables listing rules: %w: %s", err, string(out))
	}
	// parse cmd output line-by-line
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		// check for comment, ip and interface name in rule
		if strings.Contains(string(line), msqRuleComment) &&
			strings.Contains(string(line), intfName) {
			return true, nil
		}
	}
	return false, nil
}

func checkFilteringRule(cidrIP string, commandFunc runCommandFunc) (bool, error) {
	lineNum, err := checkFilteringRulesLine([]string{cidrIP}, commandFunc)
	return lineNum != -1, err
}

// returns in which line of iptables output the rule is found or -1 if not found
func checkFilteringRulesLine(cidrIPs []string, commandFunc runCommandFunc) (int, error) {
	cmd := "iptables"
	args := "-t filter -L FORWARD -v -n"

	out, err := commandFunc(cmd, strings.Split(args, " ")...)
	if err != nil {
		return -1, fmt.Errorf("iptables listing rules: %w: %s", err, string(out))
	}
	// parse cmd output line-by-line
	for lineNum, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		// check for ip (single ip e.g. 100.64.0.50 or subnet 100.64.0.0/10) and comment in rule
		if ruleContainsAllIPs(string(line), cidrIPs) &&
			strings.Contains(string(line), filterRuleComment) {
			return lineNum, nil
		}
	}
	return -1, nil
}

func ruleContainsAllIPs(line string, cidrIPs []string) bool {
	for _, ip := range cidrIPs {
		if !strings.Contains(string(line), strings.TrimSuffix(ip, "/32")) {
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
	cmd := "iptables"
	// filter out all (insert rule as 1st), except who's allowed (insert rules above)
	for _, flag := range []string{"-s", "-d"} {
		args := fmt.Sprintf(
			"-t filter -I FORWARD 1 %s %s -j DROP -m comment --comment %s",
			flag,
			meshSrcSubnet,
			filterRuleComment,
		)

		out, err := commandFunc(cmd, strings.Split(args, " ")...)
		if err != nil {
			return fmt.Errorf("iptables inserting rule: %w: %s", err, string(out))
		}
	}

	args := fmt.Sprintf(
		"-t filter -I FORWARD 1 -d %s -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT -m comment --comment %s",
		meshSrcSubnet,
		filterRuleComment,
	)

	out, err := commandFunc(cmd, strings.Split(args, " ")...)
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

func resetPeersTraffic(peers []TrafficPeer, commandFunc runCommandFunc) error {
	for _, peer := range peers {
		err := modifyPeerTraffic(peer.IP, "-D", true, true, commandFunc)
		if err != nil && !strings.Contains(err.Error(), missingRuleMessage) {
			return fmt.Errorf("deleting peer traffic rule: %w", err)
		}
		err = allowOnlyLocalNetworkAccess(peer.IP, "-D", commandFunc)
		if err != nil && !strings.Contains(err.Error(), missingRuleMessage) {
			return fmt.Errorf("deleting local network access rule: %w", err)
		}
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

	return nil
}

func allowOnlyLocalNetworkAccess(subnet netip.Prefix, flag string, commandFunc runCommandFunc) error {
	for _, localSubnet := range []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("169.254.0.0/16"),
	} {
		cmd := "iptables"
		args := fmt.Sprintf(
			"-t filter %s FORWARD -s %s -d %s -j ACCEPT -m comment --comment %s",
			flag,
			subnet.String(),
			localSubnet.String(),
			filterRuleComment,
		)
		// #nosec G204 -- input is properly sanitized
		out, err := commandFunc(cmd, strings.Split(args, " ")...)
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
	cmd := "iptables"
	args := fmt.Sprintf(
		"-t filter %s FORWARD %s %s -j %s -m comment --comment %s",
		flag,
		sourceFlag,
		subnet.String(),
		acceptFlag,
		filterRuleComment,
	)

	out, err := commandFunc(cmd, strings.Split(args, " ")...)
	if err != nil {
		return fmt.Errorf("iptables modifying rule: %w: %s", err, string(out))
	}
	return nil
}

// clearFiltering drops all the rules in the FORWARD chain containing
// a comment
func clearFiltering(commandFunc runCommandFunc) error {
	cmd := "iptables"

	out, err := commandFunc(cmd, "-S")
	if err != nil {
		return fmt.Errorf("listing iptables rules: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "FORWARD") ||
			!strings.Contains(line, fmt.Sprintf("--comment %s", filterRuleComment)) {
			continue
		}

		out, err := commandFunc(cmd, strings.Split(strings.ReplaceAll(line, "-A ", "-D "), " ")...)
		if err != nil {
			return fmt.Errorf(
				"deleting FORWARD rule %s: %w: %s",
				line, err, string(out))
		}
	}
	return nil
}
