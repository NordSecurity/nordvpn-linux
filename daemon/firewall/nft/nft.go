package nft

import (
	"fmt"
	"log"
	"net/netip"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

const tableName = "nordvpn"
const tableNameMesh = "nordvpn-meshnet"
const appFwmark uint32 = 0xe1f1
const allowedInterfacesSetName = "allowed_interfaces"
const defaultMeshSubnet = "100.64.0.0/10"

type nft struct {
	conn *nftables.Conn
}
// need set of needed ip addresses (self and all peers)
// create seperate table(optional?)
// add forward and input rules from rules.nft
func (n *nft) Configure(tunnelInterface string) error {
	if err := n.configure(tunnelInterface); err != nil {
		return fmt.Errorf("applying VPN lockdown: %w", err)
	}
	return nil
}

func (n *nft) ConfigureMesh(cfg mesh.MachineMap) error {
	if err := n.configureMesh(cfg); err != nil{
		return fmt.Errorf("applying meshnet rules: %w", err)
	}
	log.Printf("correctly set mesh rules for map: %v", cfg)
	return nil
}

func (n *nft) Flush() error {
	table := addTable(n.conn, tableName)
	n.conn.DelTable(table)
	return n.conn.Flush()
}

func (n *nft) FlushMesh() error {
	table := addTable(n.conn, tableNameMesh)
	n.conn.DelTable(table)
	return n.conn.Flush()
}

func New() *nft {
	return &nft{
		conn: &nftables.Conn{},
	}
}

func (n *nft) configure(tunnelInterface string) error {
	table := addTable(n.conn, tableName)
	n.conn.DelTable(table)
	table = addTable(n.conn, tableName)

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

func (n *nft) configureMesh(meshMap mesh.MachineMap) error {
	table := addTable(n.conn, tableNameMesh)
	n.conn.DelTable(table)
	table = addTable(n.conn, tableNameMesh)
	if err := n.conn.Flush(); err != nil {
		return err
	}
	dropPolicy := nftables.ChainPolicyDrop
	acceptPolicy := nftables.ChainPolicyAccept
	inputChain := n.conn.AddChain(&nftables.Chain{
		Name:     "input",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &acceptPolicy,
	})
	log.Printf("PRINTING ADDR AS NORMAL: %v", meshMap.Address)
	log.Println("PRINTING ADDR AS SLICE", meshMap.Address.AsSlice())
	// self accept
	// 	ip saddr 100.94.110.153   counter packets 0 bytes 0 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, addSourceIPCheck(meshMap.Address)),
	})
	// var peerRules []nftables.Rule
	// accept peers
	lanSubnets := []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "169.254.0.0/16"}
	for _, peer := range meshMap.Peers {
		if !peer.Address.IsValid() {
			continue
		}

		lanAllowed := peer.DoIAllowRouting && peer.DoIAllowLocalNetwork
		if peer.DoIAllowFileshare {
			n.conn.InsertRule(&nftables.Rule{
				Table: table,
				Chain: inputChain,
				Exprs: buildRules(expr.VerdictAccept, addSourceIPCheck(peer.Address), addPortCheck(49111, unix.IPPROTO_TCP, MATCH_DESTINATION)),
			})
		}
		if peer.DoIAllowInbound{
			if !lanAllowed {
				// 	ip saddr 100.108.155.28 ip daddr 169.254.0.0/16   counter packets 0 bytes 0 drop `block lan`
				// 	ip saddr 100.108.155.28 ip daddr 192.168.0.0/16   counter packets 0 bytes 0 drop
				// 	ip saddr 100.108.155.28 ip daddr 172.16.0.0/12   counter packets 0 bytes 0 drop
				// 	ip saddr 100.108.155.28 ip daddr 10.0.0.0/8   counter packets 0 bytes 0 drop
				for _, subnet := range lanSubnets {
					n.conn.AddRule(&nftables.Rule{
						Table: table,
						Chain: inputChain,
						Exprs: buildRules(expr.VerdictDrop, addSourceIPCheck(peer.Address), addCIDRCheck(netip.MustParsePrefix(subnet), MATCH_DESTINATION)),
					})
				}

			}
			// 	ip saddr 100.108.155.28   counter packets 1 bytes 84 accept
			n.conn.AddRule(&nftables.Rule{
				Table: table,
				Chain: inputChain,
				Exprs: buildRules(expr.VerdictAccept, addSourceIPCheck(peer.Address)),
			})
			// 	ip saddr 100.64.0.0/10 ct state related,established ct original saddr 100.94.110.153   counter packets 1 bytes 60 accept
			n.conn.AddRule(&nftables.Rule{
				Table: table,
				Chain: inputChain,
				Exprs: buildRules(expr.VerdictAccept, addCIDRCheck(netip.MustParsePrefix(defaultMeshSubnet), MATCH_SOURCE), addCheckCtState(expr.CtStateBitRELATED | expr.CtStateBitESTABLISHED), addCtOrigSrc(meshMap.Address)),
			})

		}
	}
	// 	ip saddr 100.64.0.0/10   counter packets 0 bytes 0 drop
	// default drop
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictDrop, addCIDRCheck(netip.MustParsePrefix(defaultMeshSubnet), MATCH_SOURCE)),
	})
	// n.conn.Flush()
	// chain INPUT {
		// 	type filter hook input priority filter; policy accept;
		// 	ip saddr 100.108.155.28 tcp dport 49111   counter packets 0 bytes 0 accept `fileshare`
		// 	ip saddr 100.108.155.28 ip daddr 169.254.0.0/16   counter packets 0 bytes 0 drop `block lan`
		// 	ip saddr 100.108.155.28 ip daddr 192.168.0.0/16   counter packets 0 bytes 0 drop
		// 	ip saddr 100.108.155.28 ip daddr 172.16.0.0/12   counter packets 0 bytes 0 drop
		// 	ip saddr 100.108.155.28 ip daddr 10.0.0.0/8   counter packets 0 bytes 0 drop
		// 	ip saddr 100.108.155.28   counter packets 1 bytes 84 accept
		// 	ip saddr 100.94.110.153   counter packets 0 bytes 0 accept
		// 	ip saddr 100.64.0.0/10 ct state related,established ct original saddr 100.94.110.153   counter packets 1 bytes 60 accept
		// 	ip saddr 100.64.0.0/10   counter packets 0 bytes 0 drop
		// }
		// 100.108.155.28 = peer
		// 100.94.110.153 = self
		

		// do the same for forwarder
		// chain FORWARD {
		// type filter hook forward priority filter; policy drop;
		// ip saddr 100.64.0.0/10 ip daddr 169.254.0.0/16  counter packets 0 bytes 0 drop
		// ip saddr 100.64.0.0/10 ip daddr 192.168.0.0/16  counter packets 0 bytes 0 drop
		// ip saddr 100.64.0.0/10 ip daddr 172.16.0.0/12  counter packets 0 bytes 0 drop
		// ip saddr 100.64.0.0/10 ip daddr 10.0.0.0/8  counter packets 0 bytes 0 drop
		// ip daddr 100.64.0.0/10 ct state related,established  counter packets 0 bytes 0 accept
		// ip daddr 100.64.0.0/10  counter packets 0 bytes 0 drop
		// ip saddr 100.64.0.0/10  counter packets 0 bytes 0 drop
	
	forwardChain := n.conn.AddChain(&nftables.Chain{
		Name:     "forward",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &dropPolicy,
	})
	
	for _, subnet := range lanSubnets{	
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictDrop, addCIDRCheck(netip.MustParsePrefix(defaultMeshSubnet), MATCH_SOURCE), addCIDRCheck(netip.MustParsePrefix(subnet), MATCH_DESTINATION)),
		})
	}
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: forwardChain,
		Exprs: buildRules(expr.VerdictAccept, addCIDRCheck(netip.MustParsePrefix(defaultMeshSubnet), MATCH_DESTINATION), addCheckCtState(expr.CtStateBitRELATED | expr.CtStateBitESTABLISHED)),
	})
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: forwardChain,
		Exprs: buildRules(expr.VerdictDrop, addCIDRCheck(netip.MustParsePrefix(defaultMeshSubnet), MATCH_DESTINATION)),
	})
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: forwardChain,
		Exprs: buildRules(expr.VerdictDrop, addCIDRCheck(netip.MustParsePrefix(defaultMeshSubnet), MATCH_SOURCE)),
	})
	n.conn.Flush()
		
	return nil
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

func addTable(conn *nftables.Conn, name string) *nftables.Table{
	return conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   name,
	})
}