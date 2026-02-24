package nft

import (
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

const (
	tableName                       = "nordvpn"
	excludedInterfacesSetName       = "excluded_interfaces"
	allowlistSubnetsSetName         = "allowlist_subnets"
	tcpAllowlistSetName             = "tcp_allowlist"
	udpAllowlistSetName             = "udp_allowlist"
	lanPrivateIpsSetName            = "lap_ranges"
	preroutingChainName             = "prerouting"
	inputChainName                  = "input"
	outputChainName                 = "output"
	forwardChainName                = "forward"
	meshInputChainName              = "mesh_input"
	meshForwardChainName            = "mesh_forward"
	meshNatChainName                = "mesh_nat"
	fileshareAllowedPeersSet        = "fileshare_allowed_peers"
	allowIncomingConnectionPeersSet = "allow_incoming_connections"
	allowTrafficRoutingPeersSet     = "allow_peer_traffic_routing"
	lanAccessPeersSet               = "peer_local_network_access"
)

// nft class is responsible to configure the firewall using the nftables.
// The communication with the kernel is made over netlink.
type nft struct {
	conn   *nftables.Conn
	fwmark uint32
}

func New(fwmark uint32) *nft {
	return &nft{
		conn:   &nftables.Conn{},
		fwmark: fwmark,
	}
}

func (n *nft) Configure(vpnInfo firewall.VpnInfo, meshnetInfo *firewall.MeshInfo) error {
	if err := n.configure(vpnInfo, meshnetInfo); err != nil {
		return fmt.Errorf("nft configure: %w", err)
	}
	return nil
}

func (n *nft) Flush() error {
	table := addMainTable(n.conn)
	n.conn.DelTable(table)
	return n.conn.Flush()
}

type tcpPortsSet *nftables.Set
type udpPortsSet *nftables.Set

func (n *nft) configure(vpnInfo firewall.VpnInfo, meshnetInfo *firewall.MeshInfo) error {
	isVpnOrKsSet := isVpnOrKillswitchSet(vpnInfo)

	log.Println("configure FW", isVpnOrKsSet)

	// Add and delete the table, then add again with correct rules.
	// In this way if the table exists it will be deleted and new rules will not be merged with the existing rules
	table := addMainTable(n.conn)
	n.conn.DelTable(table)
	table = addMainTable(n.conn)

	// add excluded interfaces set, lo and tunnel interface
	excludedInterfaces := &nftables.Set{
		Table:        table,
		Name:         excludedInterfacesSetName,
		KeyType:      nftables.TypeIFName,
		KeyByteOrder: binaryutil.NativeEndian,
	}

	elems := []nftables.SetElement{
		{Key: ifname("lo")},
	}

	if vpnInfo.TunnelInterface != "" {
		elems = append(elems, nftables.SetElement{Key: ifname(vpnInfo.TunnelInterface)})
	}

	if err := n.conn.AddSet(excludedInterfaces, elems); err != nil {
		return fmt.Errorf("add excluded interfaces set %w", err)
	}

	allowlistSubnets, err := n.buildAllowlistSubnets(table, vpnInfo)
	if err != nil {
		return err
	}

	var tcpPorts *nftables.Set
	if len(vpnInfo.AllowList.Ports.TCP) > 0 {
		tcpPorts = &nftables.Set{
			Table:    table,
			Name:     tcpAllowlistSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
		}
		elements := convertPortsToSetElements(vpnInfo.AllowList.GetTCPPorts())
		if err := n.conn.AddSet(tcpPorts, elements); err != nil {
			return fmt.Errorf("add TCP ports set: %w", err)
		}
	}

	var udpPorts *nftables.Set
	if len(vpnInfo.AllowList.Ports.TCP) > 0 {
		udpPorts = &nftables.Set{
			Table:    table,
			Name:     udpAllowlistSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
		}
		elements := convertPortsToSetElements(vpnInfo.AllowList.GetUDPPorts())
		if err := n.conn.AddSet(udpPorts, elements); err != nil {
			return fmt.Errorf("add UDP ports set: %w", err)
		}
	}

	var lanPrivateRanges *nftables.Set
	var fileshare *nftables.Set
	var lanAllowedPeers *nftables.Set
	var routingAllowed *nftables.Set
	var allowedIncomingConnections *nftables.Set
	if meshnetInfo != nil {
		lanPrivateRanges, err = n.buildLanRangesSet(table)
		if err != nil {
			return err
		}
		fileshare, err = n.buildFileshare(table, meshnetInfo.MeshnetMap)
		if err != nil {
			return fmt.Errorf("add fileshare peers set: %w", err)
		}

		lanAllowedPeers, err = n.buildLanAllowedPeers(table, meshnetInfo.MeshnetMap)
		if err != nil {
			return fmt.Errorf("add LAN allowed peers set: %w", err)
		}

		routingAllowed, err = n.buildAllowedRoutingPeers(table, meshnetInfo.MeshnetMap)
		if err != nil {
			return fmt.Errorf("add peers allowed to route traffic set: %w", err)
		}

		allowedIncomingConnections, err = n.buildAllowedIncomingConnections(table, meshnetInfo.MeshnetMap)
		if err != nil {
			return fmt.Errorf("add peers allowed to connect set: %w", err)
		}
	}

	n.addPreRouting(table, allowlistSubnets, tcpPorts, udpPorts)
	n.addInput(isVpnOrKsSet, table, excludedInterfaces, fileshare, allowedIncomingConnections, meshnetInfo)
	n.addOutput(isVpnOrKsSet, table, excludedInterfaces, allowlistSubnets, tcpPorts, udpPorts, meshnetInfo)
	n.addForward(vpnInfo, table, allowlistSubnets, tcpPorts, udpPorts, lanAllowedPeers, routingAllowed, lanPrivateRanges, meshnetInfo)

	if meshnetInfo != nil && routingAllowed != nil {
		n.addMeshnetNat(table, routingAllowed)
	}

	return n.conn.Flush()
}

func (n *nft) addPreRouting(table *nftables.Table, allowlistSubnets *nftables.Set, udpPortsSet tcpPortsSet, tcpPortsSet udpPortsSet) {
	acceptPolicy := nftables.ChainPolicyAccept

	preRoutingChain := n.conn.AddChain(&nftables.Chain{
		Name:     preroutingChainName,
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityMangle,
		Policy:   &acceptPolicy,
	})

	if allowlistSubnets != nil {
		// ip saddr @allowed_subnets meta mark set 0x0000e1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: preRoutingChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkIpInSet(allowlistSubnets, MATCH_SOURCE),
				setMetaMark(n.fwmark),
			),
		})
	}

	if tcpPortsSet != nil {
		// tcp sport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: preRoutingChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkPortInSet(tcpPortsSet, unix.IPPROTO_TCP, MATCH_SOURCE),
				setMetaMark(n.fwmark),
			),
		})
	}

	if udpPortsSet != nil {
		// udp sport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: preRoutingChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkPortInSet(tcpPortsSet, unix.IPPROTO_UDP, MATCH_SOURCE),
				setMetaMark(n.fwmark),
			),
		})
	}

	// ct mark 0xe1f1 meta mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: preRoutingChain,
		Exprs: buildRules(
			expr.VerdictAccept,
			checkConntrack(n.fwmark),
			setMetaMark(n.fwmark),
		),
	})
}

func (n *nft) addInput(
	isVpnOrKsSet bool,
	table *nftables.Table,
	excludedInterfaces *nftables.Set,
	filesharePeers *nftables.Set,
	allowIncomingConnections *nftables.Set,
	meshnetInfo *firewall.MeshInfo,
) {
	dropPolicy := nftables.ChainPolicyAccept
	if isVpnOrKsSet {
		dropPolicy = nftables.ChainPolicyDrop
	}

	inputChain := n.conn.AddChain(&nftables.Chain{
		Name:     inputChainName,
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &dropPolicy,
	})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, checkMetaMark(n.fwmark)),
	})

	// ct mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, checkConntrack(n.fwmark)),
	})

	// meshnet
	if filesharePeers != nil || allowIncomingConnections != nil {
		meshChain := n.conn.AddChain(&nftables.Chain{
			Name:  meshInputChainName,
			Table: table,
		})

		// iifname "nordlynx" ip saddr 100.64.0.0/10 jump mesh_input
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: inputChain,
			Exprs: buildJumpRules(
				meshChain.Name,
				checkInterfaceName(meshnetInfo.MeshInterface, IF_INPUT),
				checkIpPartOfSubnet(internal.MeshSubnet, MATCH_SOURCE),
			),
		})

		// ip saddr 100.64.0.0/29 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: meshChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkIpPartOfSubnet(internal.ReservedMeshnetSubnet, MATCH_SOURCE),
			),
		})

		if filesharePeers != nil {
			// tcp dport 49111 ip saddr @fileshare_allowed_peers accept
			n.conn.AddRule((&nftables.Rule{
				Table: table,
				Chain: meshChain,
				Exprs: buildRules(
					expr.VerdictAccept,
					checkPortNumber(internal.FilesharePort, unix.IPPROTO_TCP, MATCH_DESTINATION),
					checkIpInSet(filesharePeers, MATCH_SOURCE),
				),
			}))
		}

		// tcp dport 49111 drop
		n.conn.AddRule((&nftables.Rule{
			Table: table,
			Chain: meshChain,
			Exprs: buildRules(
				expr.VerdictDrop,
				checkPortNumber(internal.FilesharePort, unix.IPPROTO_TCP, MATCH_DESTINATION),
			),
		}))

		if allowIncomingConnections != nil {
			// ip saddr @allow_incoming_connections accept
			n.conn.AddRule((&nftables.Rule{
				Table: table,
				Chain: meshChain,
				Exprs: buildRules(
					expr.VerdictAccept,
					checkIpInSet(allowIncomingConnections, MATCH_SOURCE),
				),
			}))
		}

		// drop
		n.conn.AddRule((&nftables.Rule{
			Table: table,
			Chain: meshChain,
			Exprs: buildRules(
				expr.VerdictDrop,
			),
		}))
	}

	// iifname @allowed_interfaces accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, addInterfacesCheck(excludedInterfaces, IF_INPUT)),
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
	excludedInterfaces *nftables.Set,
	allowlistSubnets *nftables.Set,
	udpPortsSet tcpPortsSet,
	tcpPortsSet udpPortsSet,
	meshMap *firewall.MeshInfo,
) {
	dropPolicy := nftables.ChainPolicyAccept
	if isVpnOrKsSet {
		dropPolicy = nftables.ChainPolicyDrop
	}

	outputChain := n.conn.AddChain(&nftables.Chain{
		Name:     outputChainName,
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &dropPolicy,
	})

	// oifname @allowed_interfaces accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept, addInterfacesCheck(excludedInterfaces, IF_OUTPUT)),
	})

	if allowlistSubnets != nil {
		// ip daddr @allowed_subnets meta mark set 0x0000e1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkIpInSet(allowlistSubnets, MATCH_DESTINATION),
				setMetaMark(n.fwmark),
			),
		})
	}

	if tcpPortsSet != nil {
		// tcp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(expr.VerdictAccept,
				checkPortInSet(tcpPortsSet, unix.IPPROTO_TCP, MATCH_DESTINATION),
				setMetaMark(n.fwmark),
			),
		})
	}

	if udpPortsSet != nil {
		// udp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(expr.VerdictAccept,
				checkPortInSet(udpPortsSet, unix.IPPROTO_UDP, MATCH_DESTINATION),
				setMetaMark(n.fwmark),
			),
		})
	}

	// ct mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept, checkConntrack(n.fwmark)),
	})

	// meta mark 0x0000e1f1 ct mark set 0x0000e1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept,
			addMetaMarkCheckAndSetCtMark(n.fwmark),
		),
	})

	if meshMap != nil {
		// oifname "nordlynx" ip daddr 100.64.0.0/10 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(expr.VerdictAccept,
				checkInterfaceName(meshMap.MeshInterface, IF_OUTPUT),
				checkIpPartOfSubnet(internal.MeshSubnet, MATCH_DESTINATION),
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
	vpnInfo firewall.VpnInfo,
	table *nftables.Table,
	allowedSubnets *nftables.Set,
	udpPortsSet tcpPortsSet,
	tcpPortsSet udpPortsSet,
	lanAllowedPeers *nftables.Set,
	routingAllowed *nftables.Set,
	lanRangesIps *nftables.Set,
	meshMap *firewall.MeshInfo,
) {
	dropPolicy := nftables.ChainPolicyAccept
	if isVpnOrKillswitchSet(vpnInfo) {
		dropPolicy = nftables.ChainPolicyDrop
	}

	forwardChain := n.conn.AddChain(&nftables.Chain{
		Name:     forwardChainName,
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &dropPolicy,
	})

	if allowedSubnets != nil {
		// ip daddr @allowed_subnets meta mark set 0x0000e1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkIpInSet(allowedSubnets, MATCH_DESTINATION),
				setMetaMark(n.fwmark),
			),
		})
	}

	if tcpPortsSet != nil {
		// tcp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictAccept,
				checkPortInSet(tcpPortsSet, unix.IPPROTO_TCP, MATCH_DESTINATION),
				setMetaMark(n.fwmark),
			),
		})
	}

	if udpPortsSet != nil {
		// udp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictAccept,
				checkPortInSet(udpPortsSet, unix.IPPROTO_UDP, MATCH_DESTINATION),
				setMetaMark(n.fwmark),
			),
		})
	}

	// meta mark 0x0000e1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: forwardChain,
		Exprs: buildRules(expr.VerdictAccept, checkMetaMark(n.fwmark)),
	})

	if routingAllowed != nil {
		meshChain := n.conn.AddChain(&nftables.Chain{
			Name:  meshForwardChainName,
			Table: table,
		})

		// iifname "nordlynx" ip saddr 100.64.0.0/10 jump mesh_forward
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildJumpRules(
				meshChain.Name,
				checkInterfaceName(meshMap.MeshInterface, IF_INPUT),
				checkIpPartOfSubnet(internal.MeshSubnet, MATCH_SOURCE),
			),
		})

		if lanAllowedPeers != nil {
			// ip saddr @peer_local_network_access ip daddr @lan_ips accept
			n.conn.AddRule(&nftables.Rule{
				Table: table,
				Chain: meshChain,
				Exprs: buildRules(
					expr.VerdictAccept,
					checkIpInSet(lanAllowedPeers, MATCH_SOURCE),
					checkIpInSet(lanRangesIps, MATCH_DESTINATION),
				),
			})
		}

		// ip daddr @lan_ips drop
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: meshChain,
			Exprs: buildRules(
				expr.VerdictDrop,
				checkIpInSet(lanRangesIps, MATCH_DESTINATION),
			),
		})

		// ip saddr @allow_peer_traffic_routing accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: meshChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkIpInSet(routingAllowed, MATCH_SOURCE),
			),
		})

		// drop all
		n.conn.AddRule((&nftables.Rule{
			Table: table,
			Chain: meshChain,
			Exprs: buildRules(
				expr.VerdictDrop,
			),
		}))
	}

	if len(vpnInfo.TunnelInterface) > 0 {
		// TODO
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictAccept, checkInterfaceName(vpnInfo.TunnelInterface, IF_OUTPUT)),
		})

		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkInterfaceName(vpnInfo.TunnelInterface, IF_INPUT),
				addCheckCtState(expr.CtStateBitESTABLISHED|expr.CtStateBitRELATED),
			),
		})
	}
}

func (n *nft) addMeshnetNat(table *nftables.Table, routingAllowed *nftables.Set) {
	natChain := n.conn.AddChain(&nftables.Chain{
		Name:     meshNatChainName,
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	})

	// ip saddr @allow_peer_traffic_routing ip daddr != 100.64.0.0/10 masquerade
	rules := checkIpInSet(routingAllowed, MATCH_SOURCE)
	rules = append(rules, checkIpPartOfSubnet(internal.MeshSubnet, MATCH_DESTINATION)...)
	rules = append(rules, &expr.Masq{})

	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: natChain,
		Exprs: rules,
	})
}

func (n *nft) buildLanRangesSet(table *nftables.Table) (*nftables.Set, error) {
	lanNets := &nftables.Set{
		Table:    table,
		Name:     lanPrivateIpsSetName,
		KeyType:  nftables.TypeIPAddr,
		Interval: true,
	}

	elems, err := convertCidrToSetElements([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
	})
	if err != nil {
		return nil, err
	}

	if err := n.conn.AddSet(lanNets, elems); err != nil {
		return nil, err
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
			elems = append(elems, nftables.SetElement{Key: peer.Address.AsSlice()})
		}
	}

	if err := n.conn.AddSet(set, elems); err != nil {
		return nil, err
	}

	return set, nil
}

func (n *nft) buildLanAllowedPeers(table *nftables.Table, meshMap mesh.MachineMap) (*nftables.Set, error) {
	set := &nftables.Set{
		Table:    table,
		Name:     lanAccessPeersSet,
		KeyType:  nftables.TypeIPAddr,
		Interval: false,
	}

	var elems []nftables.SetElement
	for _, peer := range meshMap.Peers {
		if !peer.Address.IsValid() {
			continue
		}

		lanAllowed := peer.DoIAllowRouting && peer.DoIAllowLocalNetwork
		if lanAllowed {
			elems = append(elems, nftables.SetElement{Key: peer.Address.AsSlice()})
		}
	}

	if err := n.conn.AddSet(set, elems); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("set elements: %w", err)
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
		return nil, fmt.Errorf("set elements: %w", err)
	}

	return set, nil
}

func (n *nft) buildAllowlistSubnets(table *nftables.Table, vpnInfo firewall.VpnInfo) (*nftables.Set, error) {
	if len(vpnInfo.AllowList.Subnets) == 0 {
		return nil, nil
	}

	set := &nftables.Set{
		Table:    table,
		Name:     allowlistSubnetsSetName,
		KeyType:  nftables.TypeIPAddr,
		Interval: true,
	}

	var elements []nftables.SetElement
	for _, subnet := range vpnInfo.AllowList.Subnets {
		startIp, endIp, err := calculateFirstAndLastV4Prefix(subnet)
		if err != nil {
			return nil, fmt.Errorf("parse allowlist IP: %s %w", subnet, err)
		}

		elements = append(elements,
			nftables.SetElement{Key: startIp}, nftables.SetElement{Key: endIp, IntervalEnd: true},
		)
	}
	if err := n.conn.AddSet(set, elements); err != nil {
		return nil, fmt.Errorf("add allowlist set: %v", err)
	}

	return set, nil
}

func addMainTable(conn *nftables.Conn) *nftables.Table {
	return conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   tableName,
	})
}

func isVpnOrKillswitchSet(vpnInfo firewall.VpnInfo) bool {
	return len(vpnInfo.TunnelInterface) > 0 || vpnInfo.KillSwitch
}
