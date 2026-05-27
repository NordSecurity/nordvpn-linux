package iptables

import (
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestGenerateFlushRules(t *testing.T) {
	category.Set(t, category.Firewall)

	currentRulesMangleS := []string{
		"-A PREROUTING -i eth0 -o eth0 -m comment --comment nordvpn -j ACCEPT",
		"-A FORWARD -s 192.168.42.56/24 -i eth0 -m comment --comment \"comment\" -j ACCEPT",
		"-A FORWARD -d 10.55.97.34/24 -o eth0 -m conntrack --ctstate RELATED,ESTABLISHED -m comment --comment \"comment b\" -j ACCEPT",
		"-A FORWARD -i eth0 -m comment --comment \"comment A\" -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -o eth1 -m comment --comment \"comment B\" -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -i eth0 -m comment --comment nordvpn-meshnet -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -o eth1 -m comment --comment meshnet-nordvpn -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -i eth0 -m comment --comment nordvpn-meshnet-test -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -o eth1 -m comment --comment \"nordvpn test\" -j REJECT --reject-with icmp-port-unreachable",
		"-A POSTROUTING -o eth0 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		"-A POSTROUTING -o eth0 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
		"-A POSTROUTING -o eth0 -m comment --comment nordvpn -j DROP",
	}

	currentRulesFilterS := []string{
		"-A OUTPUT -d 169.254.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
		"-A OUTPUT -d 169.254.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
		"-A OUTPUT -d 192.168.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
		"-A OUTPUT -d 192.168.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
		"-A OUTPUT -d 172.16.0.0/12 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
		"-A OUTPUT -d 172.16.0.0/12 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
		"-A OUTPUT -d 10.0.0.0/8 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
		"-A OUTPUT -d 10.0.0.0/8 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
		"-A FORWARD -i eth0 -m comment --comment \"comment A\" -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -o eth1 -m comment --comment \"comment B\" -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -i eth0 -m comment --comment nordvpn-meshnet -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -o eth1 -m comment --comment meshnet-nordvpn -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -i eth0 -m comment --comment nordvpn-meshnet-test -j REJECT --reject-with icmp-port-unreachable",
		"-A FORWARD -o eth1 -m comment --comment \"nordvpn test\" -j REJECT --reject-with icmp-port-unreachable",
	}

	expectedRulesMangle := []string{
		"-t mangle -D PREROUTING -i eth0 -o eth0 -m comment --comment nordvpn -j ACCEPT",
		"-t mangle -D POSTROUTING -o eth0 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		"-t mangle -D POSTROUTING -o eth0 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
		"-t mangle -D POSTROUTING -o eth0 -m comment --comment nordvpn -j DROP",
	}

	expectedRulesFilter := []string{
		"-t filter -D OUTPUT -d 169.254.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
		"-t filter -D OUTPUT -d 169.254.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
		"-t filter -D OUTPUT -d 192.168.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
		"-t filter -D OUTPUT -d 192.168.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
		"-t filter -D OUTPUT -d 172.16.0.0/12 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
		"-t filter -D OUTPUT -d 172.16.0.0/12 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
		"-t filter -D OUTPUT -d 10.0.0.0/8 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
		"-t filter -D OUTPUT -d 10.0.0.0/8 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
	}

	currentRulesFilter := strings.Join(currentRulesFilterS, "\n")
	flushRulesFilter := generateFlushRules(currentRulesFilter, "filter")
	assert.Equal(t, expectedRulesFilter, flushRulesFilter)
	currentRulesMangle := strings.Join(currentRulesMangleS, "\n")
	flushRulesMangle := generateFlushRules(currentRulesMangle, "mangle")
	assert.Equal(t, expectedRulesMangle, flushRulesMangle)
}
