package nft

import (
	"fmt"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
)

const tableName = "nordvpn"
const appFwmark uint32 = 0xe1f1
const allowedInterfacesSetName = "allowed_interfaces"

type nft struct {
	conn *nftables.Conn
}

func (n *nft) Configure(tunnelInterface string) error {
	if err := n.configure(tunnelInterface); err != nil {
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

func (n *nft) configure(tunnelInterface string) error {
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

	n.addPostRouting(table)
	n.addPreRouting(table)
	n.addInput(table, allowedInterfaces)
	n.addOutput(table, allowedInterfaces)
	n.addForward(table, tunnelInterface)

	return n.conn.Flush()
}

//	chain postrouting {
//	    type filter hook postrouting priority mangle; policy accept;
//	    # Save packet fwmark
//	    meta mark 0xe1f1 ct mark set meta mark
//	}
func (n *nft) addPostRouting(table *nftables.Table) {
	acceptPolicy := nftables.ChainPolicyAccept

	postRoutingChain := n.conn.AddChain(&nftables.Chain{
		Name:     "postrouting",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityMangle,
		Policy:   &acceptPolicy,
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: postRoutingChain,
		Exprs: buildRules(expr.VerdictAccept, addMarkCheckAndSetToCt(appFwmark)),
	})
}

// //	chain prerouting {
// //	    type filter hook prerouting priority mangle; policy accept;
// //	    ct mark 0xe1f1 meta mark set 0xe1f1
// //	}
func (n *nft) addPreRouting(table *nftables.Table) {
	acceptPolicy := nftables.ChainPolicyAccept

	preRoutingChain := n.conn.AddChain(&nftables.Chain{
		Name:     "prerouting",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityMangle,
		Policy:   &acceptPolicy,
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: preRoutingChain,
		Exprs: buildRules(expr.VerdictAccept, addMarkCheckAndSetToCt(appFwmark)),
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
func (n *nft) addOutput(table *nftables.Table, allowedInterfaces *nftables.Set) {
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
		Exprs: []expr.Any{
			// meta mark 0xe1f1
			&expr.Meta{
				Key:      expr.MetaKeyMARK,
				Register: 1,
			},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     binaryutil.NativeEndian.PutUint32(appFwmark),
			},

			// accept
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})
}

//	chain forward {
//	    type filter hook forward priority filter; policy drop;
//		iifname <tun> ct state established,related accept
//		oifname <tun> accept
//	  }
func (n *nft) addForward(table *nftables.Table, tunnelInterface string) {
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
}

func addMainTable(conn *nftables.Conn) *nftables.Table {
	return conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   tableName,
	})
}
