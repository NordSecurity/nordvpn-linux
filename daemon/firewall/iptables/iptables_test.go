package iptables

import (
	"fmt"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

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
		chain       ruleChain
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
		hopLimit    uint8
		addrFlag    string
		dports      []int
		sports      []int
		comment     string
		mark        uint32
	}{
		{
			chain: chainOutput, target: drop, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", chainPrefix: "",
			rule: "OUTPUT -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainInput, target: accept, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", chainPrefix: "",
			rule: "INPUT -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPostrouting, target: drop, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", chainPrefix: "",
			rule: "POSTROUTING -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainPrerouting, target: accept, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", chainPrefix: "",
			rule: "PREROUTING -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainForward, target: drop, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", chainPrefix: "",
			rule: "FORWARD -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainForward, target: accept, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", chainPrefix: "",
			rule: "FORWARD -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainForward, target: drop, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "FORWARD -o lo -d 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainOutput, target: drop, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "OUTPUT -o lo -d 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainOutput, target: drop, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "OUTPUT -o lo -d 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainOutput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "OUTPUT -o lo -d 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainOutput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "OUTPUT -o lo -d 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPostrouting, target: drop, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "POSTROUTING -o lo -d 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainPostrouting, target: drop, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "POSTROUTING -o lo -d 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainPostrouting, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "POSTROUTING -o lo -d 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPostrouting, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "POSTROUTING -o lo -d 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainInput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainInput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainInput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "udp", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --sport 555:555 -m udp -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainInput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "udp", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --dport 555:555 -m udp -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPrerouting, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "PREROUTING -i lo -s 1.1.1.1/32 -p tcp --sport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPrerouting, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "tcp",
			port: PortRange{555, 555}, module: "", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "PREROUTING -i lo -s 1.1.1.1/32 -p tcp --dport 555:555 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPrerouting, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "udp", stateFlag: "", chainPrefix: "", portFlag: "--sport",
			rule: "PREROUTING -i lo -s 1.1.1.1/32 -p udp --sport 555:555 -m udp -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPrerouting, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "udp", stateFlag: "", chainPrefix: "", portFlag: "--dport",
			rule: "PREROUTING -i lo -s 1.1.1.1/32 -p udp --dport 555:555 -m udp -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainInput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "conntrack", stateFlag: "--ctstate",
			states: firewall.ConnectionStates{States: []firewall.ConnectionState{firewall.Established, firewall.Related}}, chainPrefix: "", portFlag: "--sport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --sport 555:555 -m conntrack --ctstate ESTABLISHED,RELATED -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainForward, target: accept,
			module: "conntrack", stateFlag: "--ctstate",
			states: firewall.ConnectionStates{States: []firewall.ConnectionState{firewall.Established, firewall.Related}},
			rule:   "FORWARD -m conntrack --ctstate ESTABLISHED,RELATED -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainInput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "conntrack", stateFlag: "--ctstate",
			states: firewall.ConnectionStates{States: []firewall.ConnectionState{firewall.Established, firewall.Related}}, chainPrefix: "", portFlag: "--dport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --dport 555:555 -m conntrack --ctstate ESTABLISHED,RELATED -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainInput, target: accept, iface: "lo", remoteNet: "1.1.1.1/32", protocol: "udp",
			port: PortRange{555, 555}, module: "conntrack", stateFlag: "--ctstate",
			states: firewall.ConnectionStates{SrcAddr: netip.MustParseAddr("2.2.2.2"), States: []firewall.ConnectionState{firewall.Established, firewall.Related}}, chainPrefix: "", portFlag: "--dport",
			rule: "INPUT -i lo -s 1.1.1.1/32 -p udp --dport 555:555 -m conntrack --ctstate ESTABLISHED,RELATED --ctorigsrc 2.2.2.2 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainOutput, target: drop, iface: "", remoteNet: "", protocol: "",
			port: PortRange{0, 0}, module: "", stateFlag: "", states: firewall.ConnectionStates{}, chainPrefix: "PRIMITIVE_",
			rule: "PRIMITIVE_OUTPUT -m comment --comment nordvpn -j DROP",
		}, {
			chain: chainInput, target: accept, mark: 0x123,
			rule: "INPUT -m connmark --mark 0x123 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainOutput, target: accept, mark: 0x123,
			rule: "OUTPUT -m connmark --mark 0x123 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainOutput, target: connmark, mark: 0x123,
			rule: "OUTPUT -m mark --mark 0x123 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		}, {
			chain: chainPrerouting, target: accept, mark: 0x123,
			rule: "PREROUTING -m connmark --mark 0x123 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPostrouting, target: accept, mark: 0x123,
			rule: "POSTROUTING -m connmark --mark 0x123 -m comment --comment nordvpn -j ACCEPT",
		}, {
			chain: chainPostrouting, target: connmark, mark: 0x123,
			rule: "POSTROUTING -m mark --mark 0x123 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
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
			rule := generateIPTablesRule(tt.chain, tt.target, net.Interface{Name: tt.iface},
				remoteNetwork, localNetwork, tt.protocol, tt.port, tt.module, tt.stateFlag, tt.states, tt.chainPrefix,
				tt.portFlag,
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
		physical bool
		out      []ruleChain
	}{
		{name: "inbound non phys", ruleType: firewall.Inbound, physical: false, out: []ruleChain{chainInput}},
		{name: "outbound non phys", ruleType: firewall.Outbound, physical: false, out: []ruleChain{chainOutput}},
		{name: "two way non phys", ruleType: firewall.TwoWay, physical: false, out: []ruleChain{chainInput, chainOutput}},
		{name: "inbound phys", ruleType: firewall.Inbound, physical: true, out: []ruleChain{chainPrerouting}},
		{name: "outbound phys", ruleType: firewall.Outbound, physical: true, out: []ruleChain{chainPostrouting}},
		{name: "two way phys", ruleType: firewall.TwoWay, physical: true, out: []ruleChain{chainPrerouting, chainPostrouting}},
		{name: "forward", ruleType: firewall.Forward, out: []ruleChain{chainForward}, physical: false},
		{name: "invalid type", ruleType: 500, out: nil, physical: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := toChainSlice(tt.ruleType, tt.physical)
			assert.Equal(t, tt.out, out)
		})
	}
}

func TestToTargetSlice(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		allowPackets bool
		chain        ruleChain
		marks        []uint32
		out          []ruleTarget
	}{
		{
			name:         "nil marks",
			allowPackets: false,
			chain:        chainOutput,
			marks:        nil,
			out:          []ruleTarget{drop},
		},
		{
			name:         "no marks",
			allowPackets: true,
			chain:        chainInput,
			marks:        []uint32{},
			out:          []ruleTarget{accept},
		},
		{
			name:         "allow incoming packets with mark",
			allowPackets: true,
			chain:        chainInput,
			marks:        []uint32{0x123},
			out:          []ruleTarget{accept},
		},
		{
			name:         "allow outgoing packets with mark",
			allowPackets: true,
			chain:        chainOutput,
			marks:        []uint32{0x123},
			out:          []ruleTarget{accept, connmark},
		},
		{
			name:         "drop incoming packets with mark",
			allowPackets: false,
			chain:        chainInput,
			marks:        []uint32{0x123},
			out:          []ruleTarget{drop},
		},
		{
			name:         "drop outgoing packets with mark",
			allowPackets: false,
			chain:        chainOutput,
			marks:        []uint32{0x123},
			out:          []ruleTarget{drop, connmark},
		},
		{
			name:         "drop prerouting incoming packets with mark",
			allowPackets: false,
			chain:        chainPrerouting,
			marks:        []uint32{0x123},
			out:          []ruleTarget{drop},
		},
		{
			name:         "drop postrouting outgoing packets with mark",
			allowPackets: false,
			chain:        chainPostrouting,
			marks:        []uint32{0x123},
			out:          []ruleTarget{drop, connmark},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := toTargetSlice(test.allowPackets, test.chain, test.marks)
			assert.Equal(t, test.out, out)
		})
	}
}

func TestRuleToIPTables(t *testing.T) {
	category.Set(t, category.Unit)

	net1111 := netip.MustParsePrefix("1.1.1.1/32")
	net2220 := netip.MustParsePrefix("2.2.2.0/24")

	tests := []struct {
		name            string
		rule            firewall.Rule
		module          string
		stateFlag       string
		chainPrefix     string
		ipv4TablesRules []string
		ipv4Count       int
	}{
		{
			name:            "empty outbound rule",
			rule:            firewall.Rule{Direction: firewall.Outbound},
			ipv4TablesRules: []string{"OUTPUT -m comment --comment nordvpn -j DROP"},
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
			name:            "empty outbound physical rule",
			rule:            firewall.Rule{Direction: firewall.Outbound, Physical: true},
			ipv4TablesRules: []string{"POSTROUTING -m comment --comment nordvpn -j DROP"},
		},
		{
			name: "outbound physical rule",
			rule: firewall.Rule{
				Direction:      firewall.Outbound,
				RemoteNetworks: []netip.Prefix{net1111},
				Physical:       true,
			},
			ipv4TablesRules: []string{"POSTROUTING -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP"},
		},
		{
			name: "two way physical rule",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Physical:       true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
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
		},
		{
			name: "multi interfaces only physical",
			rule: firewall.Rule{
				Direction:  firewall.TwoWay,
				Interfaces: []net.Interface{{Name: "lo"}, {Name: "eth0"}},
				Physical:   true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -i lo -m comment --comment nordvpn -j DROP",
				"POSTROUTING -o lo -m comment --comment nordvpn -j DROP",
				"PREROUTING -i eth0 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -o eth0 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "multi interfaces and networks rule",
			rule: firewall.Rule{
				Direction:      firewall.Outbound,
				RemoteNetworks: []netip.Prefix{net1111, net2220},
				Interfaces:     []net.Interface{{Name: "lo"}, {Name: "eth0"}},
			},
			ipv4TablesRules: []string{
				"OUTPUT -o lo -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o lo -d 2.2.2.0/24 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o eth0 -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"OUTPUT -o eth0 -d 2.2.2.0/24 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "multi interfaces and networks physical rule",
			rule: firewall.Rule{
				Direction:      firewall.Outbound,
				RemoteNetworks: []netip.Prefix{net1111, net2220},
				Interfaces:     []net.Interface{{Name: "lo"}, {Name: "eth0"}},
				Physical:       true,
			},
			ipv4TablesRules: []string{
				"POSTROUTING -o lo -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -o lo -d 2.2.2.0/24 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -o eth0 -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -o eth0 -d 2.2.2.0/24 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "multi interfaces and networks forwarding rule",
			rule: firewall.Rule{
				Direction:      firewall.Forward,
				RemoteNetworks: []netip.Prefix{net1111, net2220},
				Interfaces:     []net.Interface{{Name: "lo"}, {Name: "eth0"}},
			},
			ipv4TablesRules: []string{
				"FORWARD -o lo -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"FORWARD -o lo -d 2.2.2.0/24 -m comment --comment nordvpn -j DROP",
				"FORWARD -o eth0 -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"FORWARD -o eth0 -d 2.2.2.0/24 -m comment --comment nordvpn -j DROP",
			},
		},
		{
			name: "multiple interfaces, protocols, ports and networks",
			rule: firewall.Rule{
				RemoteNetworks: []netip.Prefix{net1111, net2220},
				Interfaces:     []net.Interface{{Name: "lo"}, {Name: "eth0"}},
				Protocols:      []string{"tcp", "udp"},
				Ports:          []int{111, 222, 333},
				Direction:      firewall.Inbound,
			},
			ipv4Count: 48,
		},
		{
			name: "two way multiple interfaces, protocols, ports and networks",
			rule: firewall.Rule{
				RemoteNetworks: []netip.Prefix{net1111, net2220},
				Interfaces:     []net.Interface{{Name: "lo"}, {Name: "eth0"}},
				Protocols:      []string{"tcp", "udp"},
				Ports:          []int{111, 222, 333},
				Direction:      firewall.TwoWay,
			},
			ipv4Count: 96,
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
			name: "two way physical rule - nil ports",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Ports:          []int(nil),
				Physical:       true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
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
			name: "two way physical rule - one port (0)",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Ports:          []int{0},
				Physical:       true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
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
			name: "two way physical rule - one port (111)",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Ports:          []int{111},
				Physical:       true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 --sport 111:111 -m comment --comment nordvpn -j DROP",
				"PREROUTING -s 1.1.1.1/32 --dport 111:111 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -d 1.1.1.1/32 --sport 111:111 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -d 1.1.1.1/32 --dport 111:111 -m comment --comment nordvpn -j DROP",
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
			name: "two way physical rule - two ports (0, 111)",
			rule: firewall.Rule{
				Direction:      firewall.TwoWay,
				RemoteNetworks: []netip.Prefix{net1111},
				Ports:          []int{0, 111},
				Physical:       true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -d 1.1.1.1/32 -m comment --comment nordvpn -j DROP",
				"PREROUTING -s 1.1.1.1/32 --sport 111:111 -m comment --comment nordvpn -j DROP",
				"PREROUTING -s 1.1.1.1/32 --dport 111:111 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -d 1.1.1.1/32 --sport 111:111 -m comment --comment nordvpn -j DROP",
				"POSTROUTING -d 1.1.1.1/32 --dport 111:111 -m comment --comment nordvpn -j DROP",
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
			name: "unblock destination ports physical rule",
			rule: firewall.Rule{
				Name:           "unblock source ports",
				Direction:      firewall.Inbound,
				Protocols:      []string{"tcp"},
				Ports:          []int{111},
				PortsDirection: firewall.Destination,
				RemoteNetworks: []netip.Prefix{
					net1111,
				},
				Allow:    true,
				Physical: true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 -p tcp --dport 111:111 -m comment --comment nordvpn -j ACCEPT",
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
			name: "unblock source ports physical rule",
			rule: firewall.Rule{
				Name:           "unblock source ports",
				Direction:      firewall.Inbound,
				Protocols:      []string{"tcp"},
				Ports:          []int{111},
				PortsDirection: firewall.Source,
				RemoteNetworks: []netip.Prefix{
					net1111,
				},
				Allow:    true,
				Physical: true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 -p tcp --sport 111:111 -m comment --comment nordvpn -j ACCEPT",
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
			name: "unblock source ports range physical rule",
			rule: firewall.Rule{
				Name:           "unblock source ports",
				Direction:      firewall.Inbound,
				Protocols:      []string{"tcp"},
				Ports:          []int{111, 112},
				PortsDirection: firewall.Source,
				RemoteNetworks: []netip.Prefix{
					net1111,
				},
				Allow:    true,
				Physical: true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 -p tcp --sport 111:112 -m comment --comment nordvpn -j ACCEPT",
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
				Allow:    true,
				Physical: true,
			},
			ipv4TablesRules: []string{
				"PREROUTING -s 1.1.1.1/32 -p tcp --sport 111:111 -m comment --comment nordvpn -j ACCEPT",
				"PREROUTING -s 1.1.1.1/32 -p tcp --sport 222:222 -m comment --comment nordvpn -j ACCEPT",
				"PREROUTING -s 1.1.1.1/32 -p tcp --sport 333:333 -m comment --comment nordvpn -j ACCEPT",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allRules := ruleToIPTables(tt.rule, tt.module, tt.stateFlag, tt.chainPrefix)
			if tt.ipv4Count > 0 {
				assert.Equal(t, tt.ipv4Count, len(allRules[ipv4Table]))
			} else {
				assert.Equal(t, tt.ipv4TablesRules, allRules[ipv4Table])
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
		{name: "connection drop physical rule", rules: map[string]firewall.Rule{
			"drop_connection": {Direction: firewall.TwoWay, Allow: false, Physical: true}},
		},
		{name: "connection accept physical rule", rules: map[string]firewall.Rule{
			"allow_connection": {Direction: firewall.TwoWay, Allow: true, Physical: true}},
		},
		{name: "single accept physical rule", rules: map[string]firewall.Rule{
			"allow_lo_interface": {Direction: firewall.TwoWay, Allow: true, Interfaces: []net.Interface{{Name: "lo"}}, Physical: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New("", "", "", []string{ipv4Table})
			// save pre-existing rules
			preRules, err := getSystemRules([]string{ipv4Table})
			assert.NoError(t, err)

			// add all rules
			for _, rule := range tt.rules {
				err = f.Add(rule)
				assert.NoError(t, err)
			}

			// check if current rules do not match the pre-existing rules
			currRules, err := getSystemRules([]string{ipv4Table})
			assert.NoError(t, err)

			for _, rule := range tt.rules {
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
			postRules, err := getSystemRules([]string{ipv4Table})
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
	tables := []string{"mangle", "filter"}
	for _, cmd := range supportedIPTables {
		var res []string
		for _, table := range tables {
			out, err := exec.Command(cmd, "-S", "-t", table, "-w", internal.SecondsToWaitForIptablesLock).CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("executing '%s -S': %w: %s", cmd, err, out)
			}
			for _, line := range strings.Split(string(out), "\n") {
				res = append(res, trimPrefixes(line, "-A"))
			}
		}
		rules[cmd] = res
	}
	return rules, nil
}
