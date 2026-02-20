package nft

import (
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

const tableName = "nordvpn"
const appFwmark uint32 = 0xe1f1
const allowedInterfacesSetName = "allowed_interfaces"
const allowedSubnetsSetName = "allowed_subnets"
const tcpSetName = "allowed_tcp"
const udpSetName = "allowed_udp"

type nft struct {
	conn *nftables.Conn
}

func (n *nft) Configure(vpnInfo *firewall.VpnInfo, meshMap *mesh.MachineMap) error {
	if err := n.configure(vpnInfo, nil); err != nil {
		return fmt.Errorf("applying VPN lockdown: %w", err)
	}
	return nil
}

func (n *nft) Flush() error {
	table := addMainTable(n.conn)
	n.conn.DelTable(table)
	return n.conn.Flush()
}

func New() *nft {
	return &nft{
		conn: &nftables.Conn{},
	}
}

type tcpPortsSet *nftables.Set
type udpPortsSet *nftables.Set

func (n *nft) configure(vpnInfo *firewall.VpnInfo, meshMap *mesh.MachineMap) error {
	log.Println("configure FW")
	table := addMainTable(n.conn)
	n.conn.DelTable(table)
	table = addMainTable(n.conn)

	// allowed interfaces set
	allowedInterfaces := &nftables.Set{
		Table:        table,
		Name:         allowedInterfacesSetName,
		KeyType:      nftables.TypeIFName,
		KeyByteOrder: binaryutil.NativeEndian,
	}
	setElements := []nftables.SetElement{
			{Key: ifname("lo")},
	}
	if vpnInfo.TunnelInterface != nil {
		setElements = append(setElements, nftables.SetElement{Key: ifname(*vpnInfo.TunnelInterface)})
	}
	n.conn.AddSet(allowedInterfaces, setElements)


	var allowedSubnets *nftables.Set
	if len(vpnInfo.Allowlist.Subnets) > 0 {
		allowedSubnets = &nftables.Set{
			Table:    table,
			Name:     allowedSubnetsSetName,
			KeyType:  nftables.TypeIPAddr,
			Interval: true,
		}
		var elements []nftables.SetElement
		for _, subnet := range vpnInfo.Allowlist.Subnets {

			startIp, endIp, err := firstLastV4(subnet)
			if err != nil {
				return fmt.Errorf("failed to parse allowlist IP: %s %w", subnet, err)
			}

			elements = append(elements, nftables.SetElement{
				Key: startIp,
			}, nftables.SetElement{
				Key:         endIp,
				IntervalEnd: true,
			})
		}
		if err := n.conn.AddSet(allowedSubnets, elements); err != nil {
			return fmt.Errorf("failed to set elements: %v", err)
		}
	}

	var tcpPorts tcpPortsSet
	if len(vpnInfo.Allowlist.Ports.TCP) > 0 {
		tcpPorts = &nftables.Set{
			Table:    table,
			Name:     tcpSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
		}
		elements := convertPortsToSetElements(vpnInfo.Allowlist.GetTCPPorts())
		if err := n.conn.AddSet(tcpPorts, elements); err != nil {
			return fmt.Errorf("failed to set elements: %v", err)
		}
	}

	var udpPorts udpPortsSet
	if len(vpnInfo.Allowlist.Ports.TCP) > 0 {
		udpPorts = &nftables.Set{
			Table:    table,
			Name:     udpSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
		}
		elements := convertPortsToSetElements(vpnInfo.Allowlist.GetUDPPorts())
		if err := n.conn.AddSet(udpPorts, elements); err != nil {
			return fmt.Errorf("failed to set elements: %v", err)
		}
	}

	// n.addPostRouting(table, allowedSubnets)
	n.addPreRouting(table, allowedSubnets, tcpPorts, udpPorts)
	n.addInput(table, allowedInterfaces)
	n.addOutput(table, allowedInterfaces, allowedSubnets, tcpPorts, udpPorts)
	if vpnInfo.TunnelInterface != nil{
		n.addForward(table, *vpnInfo.TunnelInterface, allowedSubnets, tcpPorts, udpPorts)
	}

	return n.conn.Flush()
}

//	chain postrouting {
//	    type filter hook postrouting priority mangle; policy accept;
//	    # Save packet fwmark
//	    meta mark 0xe1f1 ct mark set meta mark
//	}
// func (n *nft) addPostRouting(table *nftables.Table, allowList *nftables.Set) {
// 	acceptPolicy := nftables.ChainPolicyAccept

// 	postRoutingChain := n.conn.AddChain(&nftables.Chain{
// 		Name:     "postrouting",
// 		Table:    table,
// 		Type:     nftables.ChainTypeFilter,
// 		Hooknum:  nftables.ChainHookPostrouting,
// 		Priority: nftables.ChainPriorityMangle,
// 		Policy:   &acceptPolicy,
// 	})

// 	n.conn.AddRule(&nftables.Rule{
// 		Table: table,
// 		Chain: postRoutingChain,
// 		Exprs: buildRules(expr.VerdictAccept, addMetaMarkCheckAndSetItToCt(appFwmark)),
// 	})

// 	if allowList != nil {
// 		n.conn.AddRule(&nftables.Rule{
// 			Table: table,
// 			Chain: postRoutingChain,
// 			Exprs: buildRules(expr.VerdictAccept, addMetaMarkCheckAndSetItToCt(appFwmark)),
// 		})
// 	}
// }

// //	chain prerouting {
// //	    type filter hook prerouting priority mangle; policy accept;
// //	    ct mark 0xe1f1 meta mark set 0xe1f1
// //	}
func (n *nft) addPreRouting(
	table *nftables.Table,
	allowedSubnets *nftables.Set,
	udpPortsSet tcpPortsSet,
	tcpPortsSet udpPortsSet,
) {
	acceptPolicy := nftables.ChainPolicyAccept

	preRoutingChain := n.conn.AddChain(&nftables.Chain{
		Name:     "prerouting",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityMangle,
		Policy:   &acceptPolicy,
	})

	if allowedSubnets != nil {
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: preRoutingChain,
			Exprs: buildRules(expr.VerdictAccept,
				addIpCheckAndSetMetaMark(appFwmark, allowedSubnets, MATCH_SOURCE),
			),
		})
	}

	if tcpPortsSet != nil {
		// tcp sport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: preRoutingChain,
			Exprs: buildRules(expr.VerdictAccept,
				addPortInSetAndSetMark(appFwmark, unix.IPPROTO_TCP, MATCH_SOURCE, tcpPortsSet),
			),
		})
	}

	if udpPortsSet != nil {
		// udp sport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: preRoutingChain,
			Exprs: buildRules(expr.VerdictAccept,
				addPortInSetAndSetMark(appFwmark, unix.IPPROTO_UDP, MATCH_SOURCE, udpPortsSet),
			),
		})
	}

	//    ct mark 0xe1f1 meta mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: preRoutingChain,
		Exprs: buildRules(expr.VerdictAccept,
			addCtMarkCheck(appFwmark),
			setMetaMark(appFwmark),
		),
	})
}

// //	chain input {
// //	    type filter hook input priority filter; policy drop;
// //	    iifname "lo" accept
// //	    iifname "{{.TunnelInterface}}" accept
// //	    ct mark 0xe1f1 accept
// //		}
func (n *nft) addInput(table *nftables.Table, allowedInterfaces *nftables.Set) {
	dropPolicy := nftables.ChainPolicyDrop

	inputChain := n.conn.AddChain(&nftables.Chain{
		Name:     "input",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &dropPolicy,
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, addInterfacesCheck(allowedInterfaces, IF_INPUT)),
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, addMetaMarkCheck(appFwmark)),
	})

	// ct mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, addCtMarkCheck(appFwmark)),
	})
}

// chain output {
//     type filter hook output priority filter; policy drop;

//     oifname "{{.TunnelInterface}}" accept
//     oifname "lo" accept

//	    ct mark 0xe1f1 accept
//	    meta mark 0xe1f1 accept
//		}
func (n *nft) addOutput(
	table *nftables.Table,
	allowedInterfaces *nftables.Set,
	allowedSubnets *nftables.Set,
	udpPortsSet tcpPortsSet,
	tcpPortsSet udpPortsSet,
) {
	dropPolicy := nftables.ChainPolicyDrop

	outputChain := n.conn.AddChain(&nftables.Chain{
		Name:     "output",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &dropPolicy,
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept, addInterfacesCheck(allowedInterfaces, IF_OUTPUT)),
	})

	if allowedSubnets != nil {
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(expr.VerdictAccept,
				addIpCheckAndSetMetaMark(appFwmark, allowedSubnets, MATCH_DESTINATION),
			),
		})
	}

	if tcpPortsSet != nil {
		// tcp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(expr.VerdictAccept,
				addPortInSetAndSetMark(appFwmark, unix.IPPROTO_TCP, MATCH_DESTINATION, tcpPortsSet),
			),
		})
	}

	if udpPortsSet != nil {
		// udp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(expr.VerdictAccept,
				addPortInSetAndSetMark(appFwmark, unix.IPPROTO_UDP, MATCH_DESTINATION, udpPortsSet),
			),
		})
	}

	// ct mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept, addCtMarkCheck(appFwmark)),
	})

	// meta mark 0x0000e1f1 ct mark set 0x0000e1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept,
			addMetaMarkCheckAndSetCtMark(appFwmark),
		),
	})
}

//	chain forward {
//	    type filter hook forward priority filter; policy drop;
//		iifname <tun> ct state established,related accept
//		oifname <tun> accept
//	  }
func (n *nft) addForward(
	table *nftables.Table,
	tunnelInterface string,
	allowedSubnets *nftables.Set,
	udpPortsSet tcpPortsSet,
	tcpPortsSet udpPortsSet,
) {
	dropPolicy := nftables.ChainPolicyDrop

	forwardChain := n.conn.AddChain(&nftables.Chain{
		Name:     "forward",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &dropPolicy,
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: forwardChain,
		Exprs: buildRules(expr.VerdictAccept, addInterfaceNameCheck(tunnelInterface, IF_OUTPUT)),
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: forwardChain,
		Exprs: buildRules(
			expr.VerdictAccept,
			addInterfaceNameCheck(tunnelInterface, IF_INPUT),
			addCheckCtState(expr.CtStateBitESTABLISHED|expr.CtStateBitRELATED),
		),
	})

	if allowedSubnets != nil {
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictAccept,
				addIpCheckAndSetMetaMark(appFwmark, allowedSubnets, MATCH_DESTINATION),
			),
		})
	}

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: forwardChain,
		Exprs: buildRules(expr.VerdictAccept, addMetaMarkCheck(appFwmark)),
	})

	if tcpPortsSet != nil {
		// tcp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictAccept,
				addPortInSetAndSetMark(appFwmark, unix.IPPROTO_TCP, MATCH_DESTINATION, tcpPortsSet),
			),
		})
	}

	if udpPortsSet != nil {
		// udp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictAccept,
				addPortInSetAndSetMark(appFwmark, unix.IPPROTO_UDP, MATCH_DESTINATION, udpPortsSet),
			),
		})
	}
}

func addMainTable(conn *nftables.Conn) *nftables.Table {
	return conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   tableName,
	})
}
