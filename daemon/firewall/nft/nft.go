package nft

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
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
	lanPrivateIpsSetName            = "lan_ranges"
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
	defaultDnsPort                  = 53
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

func (n *nft) Configure(config firewall.Config) error {
	return n.configure(config)
}

func (n *nft) Flush() error {
	table := n.addMainTable()
	n.conn.DelTable(table)
	return n.conn.Flush()
}

func (n *nft) configure(config firewall.Config) error {
	// Add and delete the table, then add again with correct rules.
	// In this way if the table exists it will be deleted and new rules will not be merged with the existing rules
	table := n.addMainTable()
	n.conn.DelTable(table)
	table = n.addMainTable()

	// add excluded interfaces set, lo and tunnel interface
	excludedInterfaces := &nftables.Set{
		Table:        table,
		Name:         excludedInterfacesSetName,
		KeyType:      nftables.TypeIFName,
		KeyByteOrder: binaryutil.NativeEndian,
		// Constant:     true, // disable for strings https://github.com/google/nftables/issues/177
	}

	elems := []nftables.SetElement{
		{Key: ifname("lo")},
	}
	tunInterfaceLen := len(config.TunnelInterface)
	if tunInterfaceLen > 0 {
		if tunInterfaceLen > unix.IFNAMSIZ {
			return fmt.Errorf("interface name is too long: %s", config.TunnelInterface)
		}
		elems = append(elems, nftables.SetElement{Key: ifname(config.TunnelInterface)})
	}

	if err := n.conn.AddSet(excludedInterfaces, elems); err != nil {
		return fmt.Errorf("add excluded interfaces set %w", err)
	}

	lanPrivateRanges, err := n.buildLanRangesSet(table)
	if err != nil {
		return err
	}

	allowlistSubnets, err := n.buildAllowlistSubnets(table, config.Allowlist)
	if err != nil {
		return err
	}

	var tcpPorts *nftables.Set
	if len(config.Allowlist.Ports.TCP) > 0 {
		tcpPorts = &nftables.Set{
			Table:    table,
			Name:     tcpAllowlistSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
			Constant: true,
		}
		elements := convertPortsToSetElements(config.Allowlist.GetTCPPorts())
		if err := n.conn.AddSet(tcpPorts, elements); err != nil {
			return fmt.Errorf("add TCP ports set: %w", err)
		}
	}

	var udpPorts *nftables.Set
	if len(config.Allowlist.Ports.UDP) > 0 {
		udpPorts = &nftables.Set{
			Table:    table,
			Name:     udpAllowlistSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
			Constant: true,
		}
		elements := convertPortsToSetElements(config.Allowlist.GetUDPPorts())
		if err := n.conn.AddSet(udpPorts, elements); err != nil {
			return fmt.Errorf("add UDP ports set: %w", err)
		}
	}

	var fileshare *nftables.Set
	var lanAllowedPeers *nftables.Set
	var routingAllowed *nftables.Set
	var allowedIncomingConnections *nftables.Set
	if config.MeshnetInfo != nil {
		fileshare, err = n.buildFileshare(table, config.MeshnetInfo.MeshnetMap)
		if err != nil {
			return fmt.Errorf("add fileshare peers set: %w", err)
		}

		lanAllowedPeers, err = n.buildLanAllowedPeers(table, config.MeshnetInfo.MeshnetMap)
		if err != nil {
			return fmt.Errorf("add LAN allowed peers set: %w", err)
		}

		routingAllowed, err = n.buildAllowedRoutingPeers(table, config.MeshnetInfo.MeshnetMap)
		if err != nil {
			return fmt.Errorf("add peers allowed to route traffic set: %w", err)
		}

		allowedIncomingConnections, err = n.buildAllowedIncomingConnections(table, config.MeshnetInfo.MeshnetMap)
		if err != nil {
			return fmt.Errorf("add peers allowed to connect set: %w", err)
		}
	}

	n.addPreRoutingChain(table, allowlistSubnets, udpPorts, tcpPorts)
	n.addInputChain(config, table, excludedInterfaces, fileshare, allowedIncomingConnections)
	n.addOutputChain(config, table, excludedInterfaces, allowlistSubnets, udpPorts, tcpPorts, lanPrivateRanges)
	n.addForwardChain(config, table, allowlistSubnets, udpPorts, tcpPorts, lanAllowedPeers, routingAllowed, lanPrivateRanges, config.MeshnetInfo)
	if config.MeshnetInfo != nil && routingAllowed != nil {
		n.addMeshnetNat(table, routingAllowed)
	}

	return n.conn.Flush()
}

func (n *nft) addPreRoutingChain(
	table *nftables.Table,
	allowlistSubnets *nftables.Set,
	udpPortsSet *nftables.Set,
	tcpPortsSet *nftables.Set,
) {
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
		// udp sport @ports_udp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: preRoutingChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkPortInSet(udpPortsSet, unix.IPPROTO_UDP, MATCH_SOURCE),
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

func (n *nft) addInputChain(
	config firewall.Config,
	table *nftables.Table,
	excludedInterfaces *nftables.Set,
	filesharePeers *nftables.Set,
	allowIncomingConnections *nftables.Set,
) {
	chainPolicy := nftables.ChainPolicyAccept
	if config.IsVpnOrKillSwitchSet() {
		chainPolicy = nftables.ChainPolicyDrop
	}

	inputChain := n.conn.AddChain(&nftables.Chain{
		Name:     inputChainName,
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &chainPolicy,
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
		// Add chain for the meshnet and the jump rule to it
		meshChain := n.addMeshnetInputChain(table, config, filesharePeers, allowIncomingConnections)

		// iifname "nordlynx" ip saddr 100.64.0.0/10 jump mesh_input
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: inputChain,
			Exprs: buildJumpRules(
				meshChain.Name,
				checkInterfaceName(config.MeshnetInfo.MeshInterface, IF_INPUT),
				checkIpPartOfSubnet(internal.MeshSubnet, MATCH_SOURCE, expr.CmpOpEq),
			),
		})
	}

	// iifname @allowed_interfaces accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: inputChain,
		Exprs: buildRules(expr.VerdictAccept, addInterfacesCheck(excludedInterfaces, IF_INPUT)),
	})
}

func (n *nft) addMeshnetInputChain(
	table *nftables.Table,
	config firewall.Config,
	filesharePeers *nftables.Set,
	allowIncomingConnections *nftables.Set,
) *nftables.Chain {
	// the chain is not hooked to anything, it is called from input chain
	meshChain := n.conn.AddChain(&nftables.Chain{
		Name:  meshInputChainName,
		Table: table,
	})

	// ip saddr 100.64.0.0/29 accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: meshChain,
		Exprs: buildRules(
			expr.VerdictAccept,
			checkIpPartOfSubnet(internal.ReservedMeshnetSubnet, MATCH_SOURCE, expr.CmpOpEq),
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

	return meshChain
}

func (n *nft) addOutputChain(
	config firewall.Config,
	table *nftables.Table,
	excludedInterfaces *nftables.Set,
	allowlistSubnets *nftables.Set,
	udpPortsSet *nftables.Set,
	tcpPortsSet *nftables.Set,
	lanPrivateRanges *nftables.Set,
) {
	chainPolicy := nftables.ChainPolicyAccept
	if config.IsVpnOrKillSwitchSet() {
		chainPolicy = nftables.ChainPolicyDrop
	}

	outputChain := n.conn.AddChain(&nftables.Chain{
		Name:     outputChainName,
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &chainPolicy,
	})

	// oifname @allowed_interfaces accept
	n.conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: outputChain,
		Exprs: buildRules(expr.VerdictAccept, addInterfacesCheck(excludedInterfaces, IF_OUTPUT)),
	})

	// always drop DNS if port 53 not whitelisted
	if config.IsVpnOrKillSwitchSet() {
		if !config.Allowlist.Ports.TCP[defaultDnsPort] {
			// ip daddr @lan_ranges tcp dport 53 drop
			n.conn.AddRule(&nftables.Rule{
				Table: table,
				Chain: outputChain,
				Exprs: buildRules(
					expr.VerdictDrop,
					checkIpInSet(lanPrivateRanges, MATCH_DESTINATION),
					checkPortNumber(defaultDnsPort, unix.IPPROTO_TCP, MATCH_DESTINATION),
				),
			})
		}

		if !config.Allowlist.Ports.UDP[defaultDnsPort] {
			// ip daddr @lan_ranges udp dport 53 drop
			n.conn.AddRule(&nftables.Rule{
				Table: table,
				Chain: outputChain,
				Exprs: buildRules(
					expr.VerdictDrop,
					checkIpInSet(lanPrivateRanges, MATCH_DESTINATION),
					checkPortNumber(defaultDnsPort, unix.IPPROTO_UDP, MATCH_DESTINATION),
				),
			})
		}
	}

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
			Exprs: buildRules(
				expr.VerdictAccept,
				checkPortInSet(tcpPortsSet, unix.IPPROTO_TCP, MATCH_DESTINATION),
				setMetaMark(n.fwmark),
			),
		})
	}

	if udpPortsSet != nil {
		// udp dport @ports_udp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(
				expr.VerdictAccept,
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

	if config.MeshnetInfo != nil {
		// oifname "nordlynx" ip daddr 100.64.0.0/10 accept
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: outputChain,
			Exprs: buildRules(expr.VerdictAccept,
				checkInterfaceName(config.MeshnetInfo.MeshInterface, IF_OUTPUT),
				checkIpPartOfSubnet(internal.MeshSubnet, MATCH_DESTINATION, expr.CmpOpEq),
			),
		})
	}
}

func (n *nft) addForwardChain(
	config firewall.Config,
	table *nftables.Table,
	allowedSubnets *nftables.Set,
	udpPortsSet *nftables.Set,
	tcpPortsSet *nftables.Set,
	lanAllowedPeers *nftables.Set,
	routingAllowed *nftables.Set,
	lanRangesIps *nftables.Set,
	meshMap *firewall.MeshInfo,
) {
	chainPolicy := nftables.ChainPolicyAccept
	if config.IsVpnOrKillSwitchSet() {
		chainPolicy = nftables.ChainPolicyDrop
	}

	forwardChain := n.conn.AddChain(&nftables.Chain{
		Name:     forwardChainName,
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &chainPolicy,
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
		// udp dport @ports_udp meta mark set 0xe1f1 accept
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
				checkIpPartOfSubnet(internal.MeshSubnet, MATCH_SOURCE, expr.CmpOpEq),
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

	if len(config.TunnelInterface) > 0 {
		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(expr.VerdictAccept, checkInterfaceName(config.TunnelInterface, IF_OUTPUT)),
		})

		n.conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: forwardChain,
			Exprs: buildRules(
				expr.VerdictAccept,
				checkInterfaceName(config.TunnelInterface, IF_INPUT),
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
	rules = append(rules, checkIpPartOfSubnet(internal.MeshSubnet, MATCH_DESTINATION, expr.CmpOpNeq)...)
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
		Constant: true,
	}

	elems, err := convertCidrToSetElements(internal.LocalNetworks)
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
		Constant: true,
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
		Constant: true,
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
		Constant: true,
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
		Constant: true,
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

func (n *nft) buildAllowlistSubnets(table *nftables.Table, allowlist config.Allowlist) (*nftables.Set, error) {
	if len(allowlist.Subnets) == 0 {
		return nil, nil
	}

	set := &nftables.Set{
		Table:    table,
		Name:     allowlistSubnetsSetName,
		KeyType:  nftables.TypeIPAddr,
		Interval: true,
		Constant: true,
	}

	var elements []nftables.SetElement
	for _, subnet := range allowlist.Subnets {
		startIp, endIp, err := calculateFirstAndLastV4Prefix(subnet)
		if err != nil {
			return nil, fmt.Errorf("parse allowlist IP: %s %w", subnet, err)
		}

		elements = append(elements,
			nftables.SetElement{Key: startIp}, nftables.SetElement{Key: endIp, IntervalEnd: true},
		)
	}
	if err := n.conn.AddSet(set, elements); err != nil {
		return nil, fmt.Errorf("add allowlist set: %w", err)
	}

	return set, nil
}

func (n *nft) addMainTable() *nftables.Table {
	return n.conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   tableName,
	})
}
