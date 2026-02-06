package nft

import (
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

const tableName = "nordvpn"
const appFwmark uint32 = 0xe1f1
const allowedInterfacesSetName = "allowed_interfaces"

const rules = `add table inet nordvpn
delete table inet nordvpn

table inet nordvpn {
    chain postrouting {
        type filter hook postrouting priority mangle; policy accept;
        # Save packet fwmark
        meta mark 0xe1f1 ct mark set meta mark
    }

    chain prerouting {
        type filter hook prerouting priority mangle; policy accept;
        ct mark 0xe1f1 meta mark set ct mark
    }

  chain input {
    type filter hook input priority filter; policy drop;

    iifname "lo" accept
    iifname "{{.TunnelInterface}}" accept

    ct state established,related ct mark 0xe1f1 accept
    ct mark 0xe1f1 accept

	udp sport 53 ct state established accept
    tcp sport 53 ct state established accept
  }

    chain output {
        type filter hook output priority filter; policy drop;

        oifname "{{.TunnelInterface}}" accept
        oifname "lo" accept

        ct state new,established,related ct mark 0xe1f1 accept
        meta mark 0xe1f1 accept

		udp dport 53 accept
        tcp dport 53 accept
    }
  chain forward {
    type filter hook forward priority filter; policy drop;
  }
}`

type nft struct {
	conn *nftables.Conn
}

func (n *nft) Add(rule firewall.Rule) error {
	if rule.Name == "enable" {
		if rule.SimplifiedName == "" {
			return errors.New("Empty tun name")
		}
		if err := n.applyRules(rule.SimplifiedName); err != nil {
			return fmt.Errorf("applying VPN lockdown: %w", err)
		}
		return nil
	}
	return nil
}

func (n *nft) Delete(rule firewall.Rule) error {
	if rule.Name == "enable" {
		fmt.Println("Delete block block block")
		if err := n.removeRules(); err != nil {
			return fmt.Errorf("applying VPN lockdown: %w", err)
		}
		return nil
	}
	return nil
}

func (n *nft) Flush() error {
	return n.removeRules()
}

func (*nft) GetActiveRules() ([]string, error) {
	return nil, nil
}

func New(stateModule string, stateFlag string, chainPrefix string, supportedIPTables []string) *nft {
	return &nft{
		conn: &nftables.Conn{},
	}
}

func (n *nft) applyRules(tunnelInterface string) error {
	table := addMainTable(n.conn)
	n.conn.DelTable(table)
	table = addMainTable(n.conn)

	if err := n.conn.Flush(); err != nil {
		return err
	}

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

	if err := n.conn.Flush(); err != nil {
		return err
	}

	n.addPostRouting(table)

	if err := n.conn.Flush(); err != nil {
		return err
	}
	n.addPreRouting(table)
	if err := n.conn.Flush(); err != nil {
		return err
	}
	n.addInput(table, allowedInterfaces)
	if err := n.conn.Flush(); err != nil {
		return err
	}
	n.addOutput(table, allowedInterfaces)

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
// //	    ct state established,related ct mark 0xe1f1 accept
// //	    ct mark 0xe1f1 accept
// //			ct state established udp sport 53 accept
// //		    ct state established tcp sport 53 accept
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

	// ct state established udp sport 53 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept,
			addCheckCtState(expr.CtStateBitESTABLISHED|expr.CtStateBitRELATED),
			addPortCheck(53, unix.IPPROTO_UDP, MATCH_SOURCE),
		),
	})

	// ct state established tcp sport 53 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept,
			addCheckCtState(expr.CtStateBitESTABLISHED|expr.CtStateBitRELATED),
			addPortCheck(53, unix.IPPROTO_TCP, MATCH_SOURCE),
		),
	})
}

// chain output {
//     type filter hook output priority filter; policy drop;

//     oifname "{{.TunnelInterface}}" accept
//     oifname "lo" accept

//     ct mark 0xe1f1 accept
//     meta mark 0xe1f1 accept

//		udp dport 53 accept
//	    tcp dport 53 accept
//	}
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

	// tcp dport 53 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept,
			addPortCheck(53, unix.IPPROTO_TCP, MATCH_DESTINATION),
		),
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept,
			addPortCheck(53, unix.IPPROTO_UDP, MATCH_DESTINATION),
		),
	})
}

func (n *nft) removeRules() error {
	table := addMainTable(n.conn)
	n.conn.DelTable(table)
	return n.conn.Flush()
}

func addMainTable(conn *nftables.Conn) *nftables.Table {
	return conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   tableName,
	})
}
