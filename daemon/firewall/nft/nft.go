package nft

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
)

const tableName = "nordvpn"
const appFwmark uint32 = 0xe1f1
const allowedInterfacesSetName = "allowed_interfaces"
const allowedSubnetsSetName = "allowed_subnets"

type nft struct {
	conn *nftables.Conn
}

func (n *nft) Configure(tunnelInterface string, allowList config.Allowlist) error {
	if err := n.configure(tunnelInterface, allowList); err != nil {
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

func (n *nft) configure(tunnelInterface string, allowList config.Allowlist) error {
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

	n.conn.AddSet(allowedInterfaces, []nftables.SetElement{
		{Key: ifname("lo")},
		{Key: ifname(tunnelInterface)},
	})

	var allowedSubnets *nftables.Set
	if len(allowList.Subnets) > 0 {
		allowedSubnets = &nftables.Set{
			Table:    table,
			Name:     allowedSubnetsSetName,
			KeyType:  nftables.TypeIPAddr,
			Interval: true,
		}
		var elems []nftables.SetElement
		for _, subnet := range allowList.Subnets {

			startIp, endIp, err := firstLastV4(subnet)
			if err != nil {
				return fmt.Errorf("failed to parse allowlist IP: %s %w", subnet, err)
			}

			elems = append(elems, nftables.SetElement{
				Key: startIp,
			}, nftables.SetElement{
				Key:         endIp,
				IntervalEnd: true,
			})
		}
		if err := n.conn.AddSet(allowedSubnets, elems); err != nil {
			return fmt.Errorf("failed to set elements: %v", err)
		}
	}

	// n.addPostRouting(table, allowedSubnets)
	n.addPreRouting(table, allowedSubnets)
	n.addInput(table, allowedInterfaces)
	n.addOutput(table, allowedInterfaces, allowedSubnets)
	n.addForward(table, tunnelInterface, allowedSubnets)

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
func (n *nft) addPreRouting(table *nftables.Table, allowedSubnets *nftables.Set) {
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

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: preRoutingChain,
		Exprs: buildRules(expr.VerdictAccept,
			addMetaMarkCheckAndSetCtMark(appFwmark),
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
func (n *nft) addOutput(table *nftables.Table, allowedInterfaces *nftables.Set, allowedSubnets *nftables.Set) {
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

	// ct mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept, addCtMarkCheck(appFwmark)),
	})

	// meta mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept, addMetaMarkCheck(appFwmark)),
	})
}

//	chain forward {
//	    type filter hook forward priority filter; policy drop;
//		iifname <tun> ct state established,related accept
//		oifname <tun> accept
//	  }
func (n *nft) addForward(table *nftables.Table, tunnelInterface string, allowedSubnets *nftables.Set) {
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
}

func addMainTable(conn *nftables.Conn) *nftables.Table {
	return conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   tableName,
	})
}
