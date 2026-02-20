package nft

import (
	"fmt"
	"log"
	"net"

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
const lanIpRangesSetName = "lan_ips"

const defaultMeshSubnet = "100.64.0.0/10"
const reservedMeshnetSubnet = "100.64.0.0/29"
const fileshareAllowedPeersSet = "fileshare_allowed_peers"
const allowIncomingConnectionPeersSet = "allow_incoming_connections"
const allowTrafficRoutingPeersSet = "allow_peer_traffic_routing"
const blockedLanAccessPeersSet = "blocked_peer_local_network_access"

type nft struct {
	conn *nftables.Conn
}

func (n *nft) Configure(vpnInfo *firewall.VpnInfo, meshnetInfo *firewall.MeshInfo) error {
	if err := n.configure(vpnInfo, meshnetInfo); err != nil {
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

func (n *nft) configure(vpnInfo *firewall.VpnInfo, meshnetInfo *firewall.MeshInfo) error {
	isVpnOrKsSet := vpnInfo != nil && (len(vpnInfo.TunnelInterface) > 0 || vpnInfo.KillSwitch)

	log.Println("configure FW", isVpnOrKsSet)

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

	interfacesElements := []nftables.SetElement{
		{Key: ifname("lo")},
	}

	if vpnInfo.TunnelInterface != "" {
		interfacesElements = append(interfacesElements, nftables.SetElement{Key: ifname(vpnInfo.TunnelInterface)})
	}

	n.conn.AddSet(allowedInterfaces, interfacesElements)

	allowedSubnets, err := n.buildAllowlistSubnets(table, vpnInfo)
	if err != nil {
		return err
	}

	var tcpPorts tcpPortsSet
	if len(vpnInfo.AllowList.Ports.TCP) > 0 {
		tcpPorts = &nftables.Set{
			Table:    table,
			Name:     tcpSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
		}
		elements := convertPortsToSetElements(vpnInfo.AllowList.GetTCPPorts())
		if err := n.conn.AddSet(tcpPorts, elements); err != nil {
			return fmt.Errorf("failed to set elements: %v", err)
		}
	}

	var udpPorts udpPortsSet
	if len(vpnInfo.AllowList.Ports.TCP) > 0 {
		udpPorts = &nftables.Set{
			Table:    table,
			Name:     udpSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
		}
		elements := convertPortsToSetElements(vpnInfo.AllowList.GetUDPPorts())
		if err := n.conn.AddSet(udpPorts, elements); err != nil {
			return fmt.Errorf("failed to set elements: %v", err)
		}
	}

	lanRangesIps, err := n.buildLanRangesSet(table)
	if err != nil {
		return err
	}

	var fileshare *nftables.Set
	var lanBlockedForPeers *nftables.Set
	var routingAllowed *nftables.Set
	var allowedIncomingConnections *nftables.Set
	if meshnetInfo != nil {
		fileshare, err = n.buildFileshare(table, meshnetInfo.MeshnetMap)
		if err != nil {
			return err
		}

		lanBlockedForPeers, err = n.buildBlockedLanPeers(table, meshnetInfo.MeshnetMap)
		if err != nil {
			return err
		}

		routingAllowed, err = n.buildAllowedRoutingPeers(table, meshnetInfo.MeshnetMap)
		if err != nil {
			return err
		}

		allowedIncomingConnections, err = n.buildAllowedIncomingConnections(table, meshnetInfo.MeshnetMap)
		if err != nil {
			return err
		}

	}
	// n.addPostRouting(table, allowedSubnets)
	n.addPreRouting(table, allowedSubnets, tcpPorts, udpPorts)
	n.addInput(isVpnOrKsSet, table, allowedInterfaces, fileshare, allowedIncomingConnections)
	n.addOutput(isVpnOrKsSet, table, allowedInterfaces, allowedSubnets, tcpPorts, udpPorts, meshnetInfo)
	n.addForward(table, vpnInfo.TunnelInterface, allowedSubnets, tcpPorts, udpPorts, lanBlockedForPeers, routingAllowed, lanRangesIps)

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
func (n *nft) addInput(
	isVpnOrKsSet bool,
	table *nftables.Table,
	allowedInterfaces *nftables.Set,
	filesharePeers *nftables.Set,
	allowIncomingConnections *nftables.Set,
) {
	dropPolicy := nftables.ChainPolicyAccept
	if isVpnOrKsSet {
		dropPolicy = nftables.ChainPolicyDrop
	}

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
		Exprs: buildRules(expr.VerdictAccept, addMetaMarkCheck(appFwmark)),
	})

	// ct mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, addCtMarkCheck(appFwmark)),
	})

	// meshnet
	if filesharePeers != nil || allowIncomingConnections != nil {
		meshChain := n.conn.AddChain(&nftables.Chain{
			Name:  "mesh_input",
			Table: table,
		})

		_, ipNet, err := net.ParseCIDR(reservedMeshnetSubnet)
		if err != nil {
			log.Fatal(err)
		}
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: meshChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				addCheckIpInSubnet(ipNet, MATCH_SOURCE),
			),
		})

		if filesharePeers != nil {
			n.conn.AddRule((&nftables.Rule{
				Table: table,
				Chain: meshChain,
				Exprs: buildRules(
					expr.VerdictAccept,
					addPortCheck(49111, unix.IPPROTO_TCP, MATCH_DESTINATION),
					addIpCheckInSet(filesharePeers, MATCH_SOURCE, false),
				),
			}))

			n.conn.AddRule((&nftables.Rule{
				Table: table,
				Chain: meshChain,
				Exprs: buildRules(
					expr.VerdictDrop,
					addPortCheck(49111, unix.IPPROTO_TCP, MATCH_DESTINATION),
				),
			}))
		}

		if allowIncomingConnections != nil {
			n.conn.AddRule((&nftables.Rule{
				Table: table,
				Chain: meshChain,
				Exprs: buildRules(
					expr.VerdictAccept,
					addIpCheckInSet(allowIncomingConnections, MATCH_SOURCE, false),
				),
			}))
		}

		n.conn.AddRule((&nftables.Rule{
			Table: table,
			Chain: meshChain,
			Exprs: buildRules(
				expr.VerdictDrop,
			),
		}))

		// jump to meshnet input chain
		_, ipNet, err = net.ParseCIDR(defaultMeshSubnet)
		if err != nil {
			log.Fatal(err)
		}

		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: inputChain,
			Exprs: buildJumpRules(meshChain.Name,
				addInterfaceNameCheck("nordlynx", IF_INPUT),
				addCheckIpInSubnet(ipNet, MATCH_SOURCE),
			),
		})
	}
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, addInterfacesCheck(allowedInterfaces, IF_INPUT)),
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
	isVpnOrKsSet bool,
	table *nftables.Table,
	allowedInterfaces *nftables.Set,
	allowedSubnets *nftables.Set,
	udpPortsSet tcpPortsSet,
	tcpPortsSet udpPortsSet,
	meshMap *firewall.MeshInfo,
) {
	dropPolicy := nftables.ChainPolicyAccept
	if isVpnOrKsSet {
		dropPolicy = nftables.ChainPolicyDrop
	}

	outputChain := n.conn.AddChain(&nftables.Chain{
		Name:     "output",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &dropPolicy,
	})

	// TODO: check for OpenVPN
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

	if meshMap != nil {
		_, ipNet, err := net.ParseCIDR(defaultMeshSubnet)
		if err != nil {
			log.Fatal(err)
		}

		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(expr.VerdictAccept,
				addInterfaceNameCheck("nordlynx", IF_OUTPUT),
				addCheckIpInSubnet(ipNet, MATCH_DESTINATION),
			),
		})
	}
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
	lanBlockedForPeers *nftables.Set,
	routingAllowed *nftables.Set,
	lanRangesIps *nftables.Set,
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

	if lanBlockedForPeers != nil && lanRangesIps != nil {
		// block peers that must not access LAN
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictDrop,
				addIpCheckInSet(lanBlockedForPeers, MATCH_SOURCE, false),
				addIpCheckInSet(lanRangesIps, MATCH_DESTINATION, false),
			),
		})
	}

	if routingAllowed != nil {
		// allow routing
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictAccept,
				addIpCheckInSet(routingAllowed, MATCH_SOURCE, false),
			),
		})
	}
}

func (n *nft) buildLanRangesSet(table *nftables.Table) (*nftables.Set, error) {
	lanNets := &nftables.Set{
		Table:    table,
		Name:     lanIpRangesSetName,
		KeyType:  nftables.TypeIPAddr,
		Interval: true,
	}

	elems, err := converCidrToSetElements([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
	})
	if err != nil {
		return nil, err
	}

	if err := n.conn.AddSet(lanNets, elems); err != nil {
		return nil, fmt.Errorf("failed to set elements: %w", err)
	}

	return lanNets, nil
}

func (n *nft) buildFileshare(table *nftables.Table, meshMap mesh.MachineMap) (*nftables.Set, error) {
	set := &nftables.Set{
		Table:    table,
		Name:     fileshareAllowedPeersSet,
		KeyType:  nftables.TypeIPAddr,
		Interval: false,
	}

	var elems []nftables.SetElement
	for _, peer := range meshMap.Peers {
		if !peer.Address.IsValid() {
			continue
		}

		if peer.DoIAllowFileshare {
			elems = append(elems,
				nftables.SetElement{Key: peer.Address.AsSlice()},
			)
		}
	}

	if err := n.conn.AddSet(set, elems); err != nil {
		return nil, fmt.Errorf("failed to set elements: %w", err)
	}

	return set, nil
}

func (n *nft) buildBlockedLanPeers(table *nftables.Table, meshMap mesh.MachineMap) (*nftables.Set, error) {
	set := &nftables.Set{
		Table:    table,
		Name:     blockedLanAccessPeersSet,
		KeyType:  nftables.TypeIPAddr,
		Interval: false,
	}

	var elems []nftables.SetElement
	for _, peer := range meshMap.Peers {
		if !peer.Address.IsValid() {
			continue
		}

		lanAllowed := peer.DoIAllowRouting && peer.DoIAllowLocalNetwork
		if !lanAllowed {
			elems = append(elems,
				nftables.SetElement{Key: peer.Address.AsSlice()},
			)
		}
	}

	if err := n.conn.AddSet(set, elems); err != nil {
		return nil, fmt.Errorf("failed to set elements: %w", err)
	}

	return set, nil
}

func (n *nft) buildAllowedIncomingConnections(table *nftables.Table, meshMap mesh.MachineMap) (*nftables.Set, error) {
	set := &nftables.Set{
		Table:    table,
		Name:     allowIncomingConnectionPeersSet,
		KeyType:  nftables.TypeIPAddr,
		Interval: false,
	}

	var elems []nftables.SetElement
	for _, peer := range meshMap.Peers {
		if !peer.Address.IsValid() {
			continue
		}

		if peer.DoIAllowInbound {
			elems = append(elems,
				nftables.SetElement{Key: peer.Address.AsSlice()},
			)
		}
	}

	if err := n.conn.AddSet(set, elems); err != nil {
		return nil, fmt.Errorf("failed to set elements: %w", err)
	}

	return set, nil
}

func (n *nft) buildAllowedRoutingPeers(table *nftables.Table, meshMap mesh.MachineMap) (*nftables.Set, error) {
	set := &nftables.Set{
		Table:    table,
		Name:     allowTrafficRoutingPeersSet,
		KeyType:  nftables.TypeIPAddr,
		Interval: false,
	}

	var elems []nftables.SetElement
	for _, peer := range meshMap.Peers {
		if !peer.Address.IsValid() {
			continue
		}

		if peer.DoIAllowRouting {
			elems = append(elems,
				nftables.SetElement{Key: peer.Address.AsSlice()},
			)
		}
	}

	if err := n.conn.AddSet(set, elems); err != nil {
		return nil, fmt.Errorf("failed to set elements: %w", err)
	}

	return set, nil
}

func (n *nft) buildAllowlistSubnets(table *nftables.Table, vpnInfo *firewall.VpnInfo) (*nftables.Set, error) {
	if vpnInfo == nil {
		return nil, nil
	}

	var allowedSubnets *nftables.Set
	if len(vpnInfo.AllowList.Subnets) > 0 {
		allowedSubnets = &nftables.Set{
			Table:    table,
			Name:     allowedSubnetsSetName,
			KeyType:  nftables.TypeIPAddr,
			Interval: true,
		}
		var elements []nftables.SetElement
		for _, subnet := range vpnInfo.AllowList.Subnets {

			startIp, endIp, err := calculateFirstAndLastV4(subnet)
			if err != nil {
				return nil, fmt.Errorf("failed to parse allowlist IP: %s %w", subnet, err)
			}

			elements = append(elements, nftables.SetElement{
				Key: startIp,
			}, nftables.SetElement{
				Key:         endIp,
				IntervalEnd: true,
			})
		}
		if err := n.conn.AddSet(allowedSubnets, elements); err != nil {
			return nil, fmt.Errorf("failed to set elements: %v", err)
		}
	}

	return allowedSubnets, nil
}

func addMainTable(conn *nftables.Conn) *nftables.Table {
	return conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   tableName,
	})
}
