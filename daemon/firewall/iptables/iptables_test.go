package iptables

import (
	"fmt"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"golang.org/x/exp/slices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentInterface(t *testing.T) {
	assert.Implements(t, (*firewall.Agent)(nil), New("", "", "", []string{ipv4Table, ipv6Table}))
}
func TestConnectionStateToString(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name            string
		connectionState firewall.ConnectionState
		str             string
	}{
		{name: "new", connectionState: firewall.New, str: "NEW"},
		{name: "established", connectionState: firewall.Established, str: "ESTABLISHED"},
		{name: "related", connectionState: firewall.Related, str: "RELATED"},
		{name: "invalid", connectionState: firewall.ConnectionState(999), str: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.str, connectionStateToString(tt.connectionState))
		})
	}
}

func TestGenerateIPTablesRule(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		input       bool
		target      ruleTarget
		iface       string
		remoteNet   string
		localNet    string
		protocol    string
		port        PortRange
		module      string
		stateFlag   string
		states      firewall.ConnectionStates
		chainPrefix string
		portFlag    string
		rule        string
		icmpv6Type  int
		hopLimit    uint8
		addrFlag    string
		dports      []int
		sports      []int
		comment     string
		mark        uint32
	}{
		{
			input: false, target: drop, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", chainPrefix: "",
			rule: "OUTPUT -m comment --comment nordvpn -j DROP",
		}, {
			input: true, target: accept, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", chainPrefix: "",
			rule: "INPUT -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: false, target: drop, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "OUTPUT -o lo -d 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j DROP",
		}, {
			input: false, target: drop, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "OUTPUT -o lo -d 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j DROP",
		}, {
			input: false, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "OUTPUT -o lo -d 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: false, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "OUTPUT -o lo -d 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "udp", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --sport 555:555 -m udp -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "udp", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --dport 555:555 -m udp -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "conntrack", stateFlag: "--ctstate",
			states: firewall.ConnectionStates{States: []firewall.ConnectionState{firewall.Established, firewall.Related}}, chainPrefix: "", portFlag: "--sport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --sport 555:555 -m conntrack --ctstate ESTABLISHED,RELATED -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "conntrack", stateFlag: "--ctstate",
			states: firewall.ConnectionStates{States: []firewall.ConnectionState{firewall.Established, firewall.Related}}, chainPrefix: "", portFlag: "--dport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --dport 555:555 -m conntrack --ctstate ESTABLISHED,RELATED -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "conntrack", stateFlag: "--ctstate",
			states: firewall.ConnectionStates{SrcAddr: netip.MustParseAddr("2.2.2.2"), States: []firewall.ConnectionState{firewall.Established, firewall.Related}}, chainPrefix: "", portFlag: "--dport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --dport 555:555 -m conntrack --ctstate ESTABLISHED,RELATED --ctorigsrc 2.2.2.2 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: false, target: drop, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", states: firewall.ConnectionStates{}, chainPrefix: "PRIMITIVE_",
			rule: "PRIMITIVE_OUTPUT -m comment --comment nordvpn -j DROP",
		}, {
			input: true, target: accept, iface: "lo", remoteNet: "2606:4700:4700::1111/128",
			protocol:   "ipv6-icmp",
			icmpv6Type: 133,
			hopLimit:   255,
			rule:       "INPUT -i lo -s 2606:4700:4700::1111/128 -p ipv6-icmp --icmpv6-type 133 -m hl --hl-eq 255 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: false, target: accept, iface: "lo", localNet: "::1/128",
			protocol:   "ipv6-icmp",
			icmpv6Type: 134,
			hopLimit:   255,
			rule:       "OUTPUT -o lo -s ::1/128 -p ipv6-icmp --icmpv6-type 134 -m hl --hl-eq 255 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: false, target: accept, iface: "lo", localNet: "::1/128",
			protocol: "udp",
			sports:   []int{570},
			dports:   []int{546, 547},
			rule:     "OUTPUT -o lo -s ::1/128 -p udp --sport 570 --dport 546,547 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, iface: "lo", localNet: "::1/128",
			rule: "INPUT -i lo -s ::1/128 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: false, target: accept, iface: "lo", localNet: "::1/128",
			rule: "OUTPUT -o lo -s ::1/128 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: true, target: accept, mark: 0x123,
			rule: "INPUT -m connmark --mark 0x123 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: false, target: accept, mark: 0x123,
			rule: "OUTPUT -m connmark --mark 0x123 -m comment --comment nordvpn -j ACCEPT",
		}, {
			input: false, target: connmark, mark: 0x123,
			rule: "OUTPUT -m mark --mark 0x123 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var remoteNetwork netip.Prefix
			if tt.remoteNet != "" {
				netw, err := netip.ParsePrefix(tt.remoteNet)
				require.NoError(t, err)
				remoteNetwork = netw
			}

			var localNetwork netip.Prefix
			if tt.localNet != "" {
				netw, err := netip.ParsePrefix(tt.localNet)
				require.NoError(t, err)
				localNetwork = netw
			}
			rule := generateIPTablesRule(tt.input, tt.target, net.Interface{Name: tt.iface},
				remoteNetwork, localNetwork, tt.protocol, tt.port, tt.module, tt.stateFlag, tt.states, tt.chainPrefix,
				tt.portFlag,
				tt.icmpv6Type,
				tt.hopLimit,
				tt.sports,
				tt.dports,
				tt.comment,
				tt.mark,
			)
			assert.Equal(t, tt.rule, rule)
		})
	}
}

func TestGenerateNonEmptyRule(t *testing.T) {
	category.Set(t, category.Unit)

	filledRule := firewall.Rule{
		Interfaces:     []net.Interface{{Name: "iface"}},
		RemoteNetworks: []netip.Prefix{netip.MustParsePrefix("1.1.1.1/32")},
		LocalNetworks:  []netip.Prefix{{}},
		Ports:          []int{8000},
		Marks:          []uint32{123},
		Protocols:      []string{"tcp", "udp"},
		Direction:      firewall.TwoWay,
	}

	tests := []struct {
		name string
		rule firewall.Rule
		res  firewall.Rule
	}{
		{name: "filled rule", rule: filledRule, res: filledRule},
		{name: "empty rule", rule: firewall.Rule{}, res: firewall.Rule{
			Interfaces:     []net.Interface{{}},
			RemoteNetworks: []netip.Prefix{{}},
			LocalNetworks:  []netip.Prefix{{}},
			Ports:          []int{0},
			Marks:          []uint32{0},
			Protocols:      []string{""},
		}},
		{name: "partial rule", rule: firewall.Rule{
			Direction:  firewall.Outbound,
			Interfaces: []net.Interface{{Name: "iface"}},
			Ports:      []int{8000, 8001},
		}, res: firewall.Rule{
			Direction:      firewall.Outbound,
			Interfaces:     []net.Interface{{Name: "iface"}},
			Ports:          []int{8000, 8001},
			Marks:          []uint32{0},
			RemoteNetworks: []netip.Prefix{{}},
			LocalNetworks:  []netip.Prefix{{}},
			Protocols:      []string{""},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := generateNonEmptyRule(tt.rule)
			assert.Equal(t, tt.res, rule)
		})
	}
}

func TestToInputSlice(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		ruleType firewall.Direction
		out      []bool
	}{
		{name: "inbound", ruleType: firewall.Inbound, out: []bool{true}},
		{name: "outbound", ruleType: firewall.Outbound, out: []bool{false}},
		{name: "two way", ruleType: firewall.TwoWay, out: []bool{true, false}},
		{name: "invalid type", ruleType: 500, out: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := toInputSlice(tt.ruleType)
			assert.Equal(t, tt.out, out)
		})
	}
}

func TestToTargetSlice(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		allowPackets bool
		inputChain   bool
		marks        []uint32
		out          []ruleTarget
	}{
		{
			name:         "nil marks",
			allowPackets: false,
			inputChain:   false,
			marks:        nil,
			out:          []ruleTarget{drop},
		},
		{
			name:         "no marks",
			allowPackets: true,
			inputChain:   true,
			marks:        []uint32{},
			out:          []ruleTarget{accept},
		},
		{
			name:         "allow incoming packets with mark",
			allowPackets: true,
			inputChain:   true,
			marks:        []uint32{0x123},
			out:          []ruleTarget{accept},
		},
		{
			name:         "allow outgoing packets with mark",
			allowPackets: true,
			inputChain:   false,
			marks:        []uint32{0x123},
			out:          []ruleTarget{accept, connmark},
		},
		{
			name:         "drop incoming packets with mark",
			allowPackets: false,
			inputChain:   true,
			marks:        []uint32{0x123},
			out:          []ruleTarget{drop},
		},
		{
			name:         "drop outgoing packets with mark",
			allowPackets: false,
			inputChain:   false,
			marks:        []uint32{0x123},
			out:          []ruleTarget{drop, connmark},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := toTargetSlice(test.allowPackets, test.inputChain, test.marks)
			assert.Equal(t, test.out, out)
		})
	}
}

func TestRuleToIPTables(t *testing.T) {
	category.Set(t, category.Unit)

	net1111 := netip.MustParsePrefix("1.1.1.1/32")
	net2220 := netip.MustParsePrefix("2.2.2.0/24")
	netIpv6 := netip.MustParsePrefix("2606:4700:4700::1111/128")

	tests := []struct {
		name            string
		rule            firewall.Rule
		module          string
		stateFlag       string
		chainPrefix     string
		ipv4TablesRules []string
		ipv6TablesRules []string
		ipv4Count       int
		ipv6Count       int
	}{
		{
			name:            "empty outbound rule",
			rule:            firewall.Rule{Direction: firewall.Outbound},
			ipv4TablesRules: []string{"OUTPUT -m comment --comment nordvpn -j DROP"},
			ipv6TablesRules: []string{"OUTPUT -m comment --comment nordvpn -j DROP"},
		},
		{
			name: "outbound rule",
			rule: firewall.Rule{
				Direction:      firewall.Outbound,
				RemoteNetworks: []netip.Prefix{net1111},
			},
			ipv4TablesRules: []string{"OUTPUT -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP"},
		},
		{
			name: "two way rule",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "two way rule ipv6",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{netIpv6},
			},
			ipv6TablesRules: []string{
				"INPUT -s 2606:4700:4700::1111/128 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 2606:4700:4700::1111/128 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "multi interfaces only",
			rule: firewall.Rule{
				Direction:  firewall.TwoWay,
				Interfaces: []net.Interface{{Name: "lo"}, {Name: "eth0"}},
			},
			ipv4TablesRules: []string{
				"INPUT -i lo -m comment --comment nordvpn -j DROP",
				"OUTPUT -o lo -m comment --comment nordvpn -j DROP",
				"INPUT -i eth0 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o eth0 -m comment --comment nordvpn -j DROP",
			},
			ipv6TablesRules: []string{
				"INPUT -i lo -m comment --comment nordvpn -j DROP",
				"OUTPUT -o lo -m comment --comment nordvpn -j DROP",
				"INPUT -i eth0 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o eth0 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "multi interfaces and networks rule",
			rule: firewall.Rule{
				Direction:      firewall.Outbound,
				RemoteNetworks: []netip.Prefix{net1111, net2220, netIpv6},
				Interfaces:     []net.Interface{{Name: "lo"}, {Name: "eth0"}},
			},
			ipv4TablesRules: []string{
				"OUTPUT -o lo -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o lo -d 2.2.2.0/24 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o eth0 -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o eth0 -d 2.2.2.0/24 -m comment --comment nordvpn -j DROP",
			},
			ipv6TablesRules: []string{
				"OUTPUT -o lo -d 2606:4700:4700::1111/128 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o eth0 -d 2606:4700:4700::1111/128 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "multiple interfaces, protocols, ports and networks",
			rule: firewall.Rule{
				RemoteNetworks: []netip.Prefix{net1111, net2220, netIpv6},
				Interfaces:     []net.Interface{{Name: "lo"}, {Name: "eth0"}},
				Protocols:      []string{"tcp", "udp"},
				Ports:          []int{111, 222, 333},
				Direction:      firewall.Inbound,
			},
			ipv4Count: 48,
			ipv6Count: 24,
		},
		{
			name: "two way multiple interfaces, protocols, ports and networks",
			rule: firewall.Rule{
				RemoteNetworks: []netip.Prefix{net1111, net2220, netIpv6},
				Interfaces:     []net.Interface{{Name: "lo"}, {Name: "eth0"}},
				Protocols:      []string{"tcp", "udp"},
				Ports:          []int{111, 222, 333},
				Direction:      firewall.TwoWay,
			},
			ipv4Count: 96,
			ipv6Count: 48,
		},
		// Unit test cases for 3.8.4 hotfix. Split iptables rule into 2 due to a single rule being interpreted as AND, and 2 rules as OR.
		{
			name: "two way rule - nil ports",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Ports:          []int(nil),
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "two way rule - one port (0)",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Ports:          []int{0},
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "two way rule - one port (111)",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Ports:          []int{111},
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 --sport 111:111 -m comment --comment nordvpn -j DROP",
				"INPUT -s 1.1.1.1/32 --dport 111:111 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 1.1.1.1/32 --sport 111:111 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 1.1.1.1/32 --dport 111:111 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "two way rule - two ports (0, 111)",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Ports:          []int{0, 111},
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"INPUT -s 1.1.1.1/32 --sport 111:111 -m comment --comment nordvpn -j DROP",
				"INPUT -s 1.1.1.1/32 --dport 111:111 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 1.1.1.1/32 --sport 111:111 -m comment --comment nordvpn -j DROP",
				"OUTPUT -d 1.1.1.1/32 --dport 111:111 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "unblock destination ports rule",
			rule: firewall.Rule{
				Name:           "unblock source ports",
				Direction:      firewall.Inbound,
				Protocols:      []string{"tcp"},
				Ports:          []int{111},
				PortsDirection: firewall.Destination,
				RemoteNetworks: []netip.Prefix{
					net1111,
				},
				Allow: true,
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 -p tcp --dport 111:111 -m comment --comment nordvpn -j ACCEPT",
			},
		},
		{
			name: "unblock source ports rule",
			rule: firewall.Rule{
				Name:           "unblock source ports",
				Direction:      firewall.Inbound,
				Protocols:      []string{"tcp"},
				Ports:          []int{111},
				PortsDirection: firewall.Source,
				RemoteNetworks: []netip.Prefix{
					net1111,
				},
				Allow: true,
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 -p tcp --sport 111:111 -m comment --comment nordvpn -j ACCEPT",
			},
		},
		{
			name: "unblock source ports range rule",
			rule: firewall.Rule{
				Name:           "unblock source ports",
				Direction:      firewall.Inbound,
				Protocols:      []string{"tcp"},
				Ports:          []int{111, 112},
				PortsDirection: firewall.Source,
				RemoteNetworks: []netip.Prefix{
					net1111,
				},
				Allow: true,
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 -p tcp --sport 111:112 -m comment --comment nordvpn -j ACCEPT",
			},
		},
		{
			name: "unblock multiple source ports rule",
			rule: firewall.Rule{
				Name:           "unblock source ports",
				Direction:      firewall.Inbound,
				Protocols:      []string{"tcp"},
				Ports:          []int{111, 222, 333},
				PortsDirection: firewall.Source,
				RemoteNetworks: []netip.Prefix{
					net1111,
				},
				Allow: true,
			},
			ipv4TablesRules: []string{
				"INPUT -s 1.1.1.1/32 -p tcp --sport 111:111 -m comment --comment nordvpn -j ACCEPT",
				"INPUT -s 1.1.1.1/32 -p tcp --sport 222:222 -m comment --comment nordvpn -j ACCEPT",
				"INPUT -s 1.1.1.1/32 -p tcp --sport 333:333 -m comment --comment nordvpn -j ACCEPT",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allRules := ruleToIPTables(tt.rule, tt.module, tt.stateFlag, tt.chainPrefix)
			var countingOnly bool
			if tt.ipv4Count > 0 {
				countingOnly = true
				assert.Equal(t, tt.ipv4Count, len(allRules[ipv4Table]))
			}
			if tt.ipv6Count > 0 {
				countingOnly = true
				assert.Equal(t, tt.ipv6Count, len(allRules[ipv6Table]))
			}
			if countingOnly == false {
				assert.Equal(t, tt.ipv4TablesRules, allRules[ipv4Table])
				assert.Equal(t, tt.ipv6TablesRules, allRules[ipv6Table])
			}
		})
	}
}

func TestFirewall_AddDeleteRules(t *testing.T) {
	category.Set(t, category.Firewall)
	tests := []struct {
		name  string
		rules map[string]firewall.Rule
		err   bool
	}{
		{name: "connection drop rule", rules: map[string]firewall.Rule{
			"drop_connection": {Direction: firewall.TwoWay, Allow: false}},
		},
		{name: "connection accept rule", rules: map[string]firewall.Rule{
			"allow_connection": {Direction: firewall.TwoWay, Allow: true}},
		},
		{name: "single accept rule", rules: map[string]firewall.Rule{
			"allow_lo_interface": {Direction: firewall.TwoWay, Allow: true, Interfaces: []net.Interface{{Name: "lo"}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New("", "", "", []string{ipv4Table, ipv6Table})
			// save pre-existing rules
			preRules, err := getSystemRules([]string{ipv4Table, ipv6Table})
			assert.NoError(t, err)

			// add all rules
			for _, rule := range tt.rules {
				err = f.Add(rule)
				assert.NoError(t, err)
			}

			// check if current rules do not match the pre-existing rules
			currRules, err := getSystemRules([]string{ipv4Table, ipv6Table})
			assert.NoError(t, err)

			var ruleNames []string
			for name, rule := range tt.rules {
				ruleNames = append(ruleNames, name)
				allRules := ruleToIPTables(rule, f.stateModule, f.stateFlag, f.chainPrefix)
				for key := range allRules {
					assert.True(t, containsSlice(t, currRules[key], allRules[key]))
				}
			}

			// delete added rules and check that current rules match pre-existing rules
			for _, rule := range tt.rules {
				err = f.Delete(rule)
				assert.NoError(t, err)
			}
			postRules, err := getSystemRules([]string{ipv4Table, ipv6Table})
			assert.NoError(t, err)
			assert.Equal(t, preRules, postRules)
		})
	}
}

func TestPortsToRanges(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name   string
		ports  []int
		ranges []PortRange
	}{
		{name: "empoty slice", ports: nil, ranges: nil},
		{name: "single port", ports: []int{1}, ranges: []PortRange{{Min: 1, Max: 1}}},
		{name: "single range", ports: []int{1, 2}, ranges: []PortRange{{Min: 1, Max: 2}}},
		{name: "unsorted range", ports: []int{2, 1, 3}, ranges: []PortRange{{Min: 1, Max: 3}}},
		{name: "multiple ports multiple ranges", ports: []int{1, 3}, ranges: []PortRange{{Min: 1, Max: 1}, {Min: 3, Max: 3}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges := PortsToPortRanges(tt.ports)
			assert.Equal(t, tt.ranges, ranges)
		})
	}
}

func containsSlice(t *testing.T, list, sublist []string) bool {
	t.Helper()
	for _, s := range sublist {
		if !slices.Contains(list, s) {
			return false
		}
	}
	return true
}

func getSystemRules(supportedIPTables []string) (map[string][]string, error) {
	rules := make(map[string][]string)
	for _, cmd := range supportedIPTables {
		out, err := exec.Command(cmd, "-S").CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("executing '%s -S': %w", cmd, err)
		}
		var res []string
		for _, line := range strings.Split(string(out), "\n") {
			res = append(res, trimPrefixes(line, "-A"))
		}
		rules[cmd] = res
	}
	return rules, nil
}
