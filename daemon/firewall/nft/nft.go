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
	"github.com/google/nftables/userdata"
	"golang.org/x/sys/unix"
)

const (
	tableName                       = "nordvpn"
	excludedInterfacesSetName       = "excluded_interfaces"
	allowlistSubnetsSetName         = "allowlist_subnets"
	tcpAllowlistSetName             = "tcp_allowlist"
	udpAllowlistSetName             = "udp_allowlist"
	lanPrivateIpsSetName            = "lan_ranges"
	inputChainName                  = "input"
	outputChainName                 = "output"
	forwardChainName                = "forward"
	meshInputChainName              = "mesh_input"
	allowlistInputChainName         = "allowlist_input"
	meshPeerToInternet              = "mesh_peer_to_internet"
	internetToMeshPeer              = "internet_to_mesh_peer"
	meshNatChainName                = "mesh_nat"
	allowlistNatChainName           = "allowlist_nat"
	fileshareAllowedPeersSet        = "fileshare_allowed_peers"
	allowIncomingConnectionPeersSet = "allow_incoming_connections"
	allowTrafficRoutingPeersSet     = "allow_peer_traffic_routing"
	lanAccessPeersSet               = "peer_local_network_access"
	defaultDNSPort                  = 53
)

type nftContext struct {
	table                          *nftables.Table
	excludedInterfaces             *nftables.Set
	lanRanges                      *nftables.Set
	allowlistSubnets               *nftables.Set
	tcpPorts                       *nftables.Set
	udpPorts                       *nftables.Set
	fileshareAllowedPeers          *nftables.Set
	meshLanAllowedPeers            *nftables.Set
	meshRoutingAllowed             *nftables.Set
	meshAllowedIncomingConnections *nftables.Set
}

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
	nftCtx := &nftContext{}

	// Add and delete the table, then add again with correct rules.
	// In this way if the table exists it will be deleted and new rules will not be merged with the existing rules
	nftCtx.table = n.addMainTable()
	n.conn.DelTable(nftCtx.table)
	nftCtx.table = n.addMainTable()

	if err := n.addExcludedInterfacesSet(config, nftCtx); err != nil {
		return err
	}

	if err := n.addLanRangesSet(nftCtx); err != nil {
		return err
	}

	if err := n.addAllowlistSubnets(config.Allowlist, nftCtx); err != nil {
		return err
	}

	if err := n.addAllowlistPorts(config, nftCtx); err != nil {
		return err
	}

	if config.MeshnetInfo != nil {
		if !config.MeshnetInfo.BlockFileshare {
			if err := n.addFilesharePeers(config.MeshnetInfo.MeshnetMap, nftCtx); err != nil {
				return err
			}
		}

		if err := n.addLanAllowedPeers(config.MeshnetInfo.MeshnetMap, nftCtx); err != nil {
			return err
		}

		if err := n.addAllowedRoutingPeers(config.MeshnetInfo.MeshnetMap, nftCtx); err != nil {
			return err
		}

		if err := n.addAllowedIncomingConnections(config.MeshnetInfo.MeshnetMap, nftCtx); err != nil {
			return err
		}
	}

	n.addInputChain(config, nftCtx)
	n.addOutputChain(config, nftCtx)
	n.addForwardChain(config, nftCtx)
	if config.MeshnetInfo != nil && nftCtx.meshRoutingAllowed != nil {
		n.addMeshnetNat(nftCtx)
	}

	if len(config.TunnelInterface) > 0 && (nftCtx.udpPorts != nil || nftCtx.tcpPorts != nil) {
		n.addAllowlistNat(config, nftCtx)
	}

	return n.conn.Flush()
}

func (n *nft) addInputChain(config firewall.Config, nftCtx *nftContext) {
	chainPolicy := nftables.ChainPolicyAccept
	if config.IsVpnOrKillSwitchSet() {
		chainPolicy = nftables.ChainPolicyDrop
	}

	inputChain := n.conn.AddChain(&nftables.Chain{
		Name:     inputChainName,
		Table:    nftCtx.table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &chainPolicy,
	})

	// iifname lo accept
	n.conn.AddRule(&nftables.Rule{
		Table: nftCtx.table,
		Chain: inputChain,
		Exprs: buildRules(
			&expr.Verdict{Kind: expr.VerdictAccept},
			checkInterfaceName("lo", ifNameInput, expr.CmpOpEq),
		),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "local to local"),
	})

	// ct mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table:    nftCtx.table,
		Chain:    inputChain,
		Exprs:    buildRules(&expr.Verdict{Kind: expr.VerdictAccept}, checkCtMark(n.fwmark)),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "response for sockets with SO_MARK"),
	})

	// meshnet
	if nftCtx.fileshareAllowedPeers != nil || nftCtx.meshAllowedIncomingConnections != nil {
		// Add chain for the meshnet and the jump rule to it
		meshChain := n.addMeshnetInputChain(nftCtx)

		// iifname "nordlynx" ip saddr 100.64.0.0/10 jump mesh_input
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: inputChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictJump, Chain: meshChain.Name},
				checkInterfaceName(config.MeshnetInfo.MeshInterface, ifNameInput, expr.CmpOpEq),
				checkIPIsPartOfSubnet(internal.MeshSubnet, matchSource, expr.CmpOpEq),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "meshnet to local"),
		})
	}

	if len(config.TunnelInterface) > 0 {
		// iifname "nordlynx" ct state established,related accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: inputChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkInterfaceName(config.TunnelInterface, ifNameInput, expr.CmpOpEq),
				checkCtState(expr.CtStateBitESTABLISHED|expr.CtStateBitRELATED),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "response to connections inside tunnel"),
		})
	}

	if nftCtx.allowlistSubnets != nil || nftCtx.tcpPorts != nil || nftCtx.udpPorts != nil {
		chain := n.addAllowlistInputChain(nftCtx)

		// iifname != "nordtun" jump allowlist_input
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: inputChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictJump, Chain: chain.Name},
				checkInterfaceName(config.TunnelInterface, ifNameInput, expr.CmpOpNeq),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "allowlist to local"),
		})
	}
}

func (n *nft) addOutputChain(config firewall.Config, nftCtx *nftContext) {
	chainPolicy := nftables.ChainPolicyAccept
	if config.IsVpnOrKillSwitchSet() {
		chainPolicy = nftables.ChainPolicyDrop
	}

	outputChain := n.conn.AddChain(&nftables.Chain{
		Name:     outputChainName,
		Table:    nftCtx.table,
		Type:     nftables.ChainTypeRoute,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityMangle,
		Policy:   &chainPolicy,
	})

	// drop DNS if port 53 not whitelisted
	n.addLanDNSDrop(config, nftCtx, outputChain)

	if nftCtx.allowlistSubnets != nil {
		// ip daddr @allowed_subnets accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: outputChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkIPIsInSet(nftCtx.allowlistSubnets, matchDest),
				setMetaMark(n.fwmark),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "local to allowlist IPs"),
		})
	}

	if nftCtx.tcpPorts != nil {
		// tcp sport @ports_tcp accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: outputChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortIsInSet(nftCtx.tcpPorts, unix.IPPROTO_TCP, matchSource),
				setMetaMark(n.fwmark),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "from allowlist TCP ports"),
		})
	}

	if nftCtx.udpPorts != nil {
		// udp sport @ports_udp accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: outputChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortIsInSet(nftCtx.udpPorts, unix.IPPROTO_UDP, matchSource),
				setMetaMark(n.fwmark),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "from allowlist UDP ports"),
		})
	}

	// ct mark 0xe1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: nftCtx.table,
		Chain: outputChain,
		Exprs: buildRules(
			&expr.Verdict{Kind: expr.VerdictAccept},
			checkCtMark(n.fwmark)),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "VPN transport continuation"),
	})

	// meta mark 0x0000e1f1 ct mark set 0x0000e1f1 accept
	n.conn.AddRule(&nftables.Rule{
		Table: nftCtx.table,
		Chain: outputChain,
		Exprs: buildRules(
			&expr.Verdict{Kind: expr.VerdictAccept},
			checkMetaMarkAndSetCtMark(n.fwmark),
		),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "mark connection for socket with SO_MARK"),
	})

	// oifname @excluded_interfaces accept
	n.conn.AddRule(&nftables.Rule{
		Table: nftCtx.table,
		Chain: outputChain,
		Exprs: buildRules(
			&expr.Verdict{Kind: expr.VerdictAccept},
			interfaceNameInSet(nftCtx.excludedInterfaces, ifNameOutput),
		),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "local to local and local to VPN"),
	})

	if config.MeshnetInfo != nil {
		// oifname "nordlynx" ip daddr 100.64.0.0/10 accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: outputChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkInterfaceName(config.MeshnetInfo.MeshInterface, ifNameOutput, expr.CmpOpEq),
				checkIPIsPartOfSubnet(internal.MeshSubnet, matchDest, expr.CmpOpEq),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "local to meshnet"),
		})
	}
}

func (n *nft) addForwardChain(config firewall.Config, nftCtx *nftContext) {
	chainPolicy := nftables.ChainPolicyAccept
	if config.IsVpnOrKillSwitchSet() {
		chainPolicy = nftables.ChainPolicyDrop
	}

	forwardChain := n.conn.AddChain(&nftables.Chain{
		Name:     forwardChainName,
		Table:    nftCtx.table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &chainPolicy,
	})

	if nftCtx.meshRoutingAllowed != nil {
		meshToInternetChain := n.addMeshPeerToInternet(config, nftCtx)
		// ip saddr 100.64.0.0/10 jump mesh_peer_internet
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: forwardChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictJump, Chain: meshToInternetChain.Name},
				checkIPIsPartOfSubnet(internal.MeshSubnet, matchSource, expr.CmpOpEq),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "traffic from mesh peer"),
		})

		internetToMeshChain := n.addInternetToMeshPeer(config, nftCtx)
		// ip daddr 100.64.0.0/10 jump internet_to_mesh_peer
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: forwardChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictJump, Chain: internetToMeshChain.Name},
				checkIPIsPartOfSubnet(internal.MeshSubnet, matchDest, expr.CmpOpEq),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "traffic to mesh peer"),
		})

	}

	n.addLanDNSDrop(config, nftCtx, forwardChain)

	if nftCtx.allowlistSubnets != nil {
		// ip daddr @allowed_subnets accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: forwardChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkIPIsInSet(nftCtx.allowlistSubnets, matchDest),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "internet to allowlist IPs"),
		})

		// ip saddr @allowlist_subnets ct state established,related accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: forwardChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkIPIsInSet(nftCtx.allowlistSubnets, matchSource),
				checkCtState(expr.CtStateBitESTABLISHED|expr.CtStateBitRELATED),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "allow responses to allowlist IPs"),
		})
	}

	if len(config.TunnelInterface) > 0 {
		// oif "nordtun" accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: forwardChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkInterfaceName(config.TunnelInterface, ifNameOutput, expr.CmpOpEq),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "internet to allowlist IPs"),
		})
	}
}

func (n *nft) addAllowlistPorts(config firewall.Config, nftCtx *nftContext) error {
	if len(config.Allowlist.Ports.TCP) > 0 {
		nftCtx.tcpPorts = &nftables.Set{
			Table:    nftCtx.table,
			Name:     tcpAllowlistSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
			Constant: true,
		}
		elements := convertPortsToSetElements(config.Allowlist.GetTCPPorts())
		if err := n.conn.AddSet(nftCtx.tcpPorts, elements); err != nil {
			return fmt.Errorf("add TCP ports set: %w", err)
		}
	}

	if len(config.Allowlist.Ports.UDP) > 0 {
		nftCtx.udpPorts = &nftables.Set{
			Table:    nftCtx.table,
			Name:     udpAllowlistSetName,
			KeyType:  nftables.TypeInetService,
			Interval: true,
			Constant: true,
		}
		elements := convertPortsToSetElements(config.Allowlist.GetUDPPorts())
		if err := n.conn.AddSet(nftCtx.udpPorts, elements); err != nil {
			return fmt.Errorf("add UDP ports set: %w", err)
		}
	}
	return nil
}

func (n *nft) addExcludedInterfacesSet(config firewall.Config, nftCtx *nftContext) error {
	elems := []nftables.SetElement{
		{Key: ifname("lo")},
	}
	// add excluded interfaces set, lo and tunnel interface
	nftCtx.excludedInterfaces = &nftables.Set{
		Table:        nftCtx.table,
		Name:         excludedInterfacesSetName,
		KeyType:      nftables.TypeIFName,
		KeyByteOrder: binaryutil.NativeEndian,
		// Constant:     true, // disable for strings https://github.com/google/nftables/issues/177
	}

	tunInterfaceLen := len(config.TunnelInterface)
	if tunInterfaceLen > 0 {
		if tunInterfaceLen > unix.IFNAMSIZ {
			return fmt.Errorf("interface name is too long: %s", config.TunnelInterface)
		}
		elems = append(elems, nftables.SetElement{Key: ifname(config.TunnelInterface)})
	}

	if err := n.conn.AddSet(nftCtx.excludedInterfaces, elems); err != nil {
		return fmt.Errorf("add excluded interfaces set %w", err)
	}
	return nil
}

func (n *nft) addAllowlistInputChain(nftCtx *nftContext) *nftables.Chain {

	chain := n.conn.AddChain(&nftables.Chain{
		Name:  allowlistInputChainName,
		Table: nftCtx.table,
	})

	if nftCtx.allowlistSubnets != nil {
		// ip saddr @allowed_subnets meta mark set 0x0000e1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkIPIsInSet(nftCtx.allowlistSubnets, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "allowlist IPs to local"),
		})
	}

	if nftCtx.tcpPorts != nil {
		// tcp dport @ports_tcp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortIsInSet(nftCtx.tcpPorts, unix.IPPROTO_TCP, matchDest),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "to local TCP ports"),
		})
	}

	if nftCtx.udpPorts != nil {
		// udp dport @ports_udp meta mark set 0xe1f1 accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortIsInSet(nftCtx.udpPorts, unix.IPPROTO_UDP, matchDest),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "to local UDP ports"),
		})
	}

	return chain
}

func (n *nft) addMeshnetInputChain(nftCtx *nftContext) *nftables.Chain {
	// the chain is not hooked to anything, it is called from input chain
	meshChain := n.conn.AddChain(&nftables.Chain{
		Name:  meshInputChainName,
		Table: nftCtx.table,
	})

	// ip saddr 100.64.0.0/29 accept
	n.conn.AddRule(&nftables.Rule{
		Table: nftCtx.table,
		Chain: meshChain,
		Exprs: buildRules(
			&expr.Verdict{Kind: expr.VerdictAccept},
			checkIPIsPartOfSubnet(internal.ReservedMeshnetSubnet, matchSource, expr.CmpOpEq),
		),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "meshnet private IP"),
	})

	if nftCtx.fileshareAllowedPeers != nil {
		// tcp dport 49111 ip saddr @fileshare_allowed_peers accept
		n.conn.AddRule((&nftables.Rule{
			Table: nftCtx.table,
			Chain: meshChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortNumber(internal.FilesharePort, unix.IPPROTO_TCP, matchDest),
				checkIPIsInSet(nftCtx.fileshareAllowedPeers, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "meshnet to fileshare"),
		}))
	}

	// tcp dport 49111 drop
	n.conn.AddRule((&nftables.Rule{
		Table: nftCtx.table,
		Chain: meshChain,
		Exprs: buildRules(
			&expr.Verdict{Kind: expr.VerdictDrop},
			checkPortNumber(internal.FilesharePort, unix.IPPROTO_TCP, matchDest),
		),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "meshnet to fileshare"),
	}))

	if nftCtx.meshAllowedIncomingConnections != nil {
		// ip saddr @allow_incoming_connections accept
		n.conn.AddRule((&nftables.Rule{
			Table: nftCtx.table,
			Chain: meshChain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkIPIsInSet(nftCtx.meshAllowedIncomingConnections, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "meshnet to local"),
		}))
	}

	// drop
	n.conn.AddRule((&nftables.Rule{
		Table: nftCtx.table,
		Chain: meshChain,
		Exprs: buildRules(&expr.Verdict{Kind: expr.VerdictDrop}),
	}))

	return meshChain
}

func (n *nft) addLanDNSDrop(config firewall.Config, nftCtx *nftContext, chain *nftables.Chain) {
	if config.IsVpnOrKillSwitchSet() {
		if !config.Allowlist.Ports.TCP[defaultDNSPort] {
			// ip daddr @lan_ranges tcp dport 53 drop
			n.conn.AddRule(&nftables.Rule{
				Table: nftCtx.table,
				Chain: chain,
				Exprs: buildRules(
					&expr.Verdict{Kind: expr.VerdictDrop},
					checkIPIsInSet(nftCtx.lanRanges, matchDest),
					checkPortNumber(defaultDNSPort, unix.IPPROTO_TCP, matchDest),
				),
				UserData: userdata.AppendString(nil, userdata.TypeComment, "block to LAN DNS for TCP"),
			})
		}

		if !config.Allowlist.Ports.UDP[defaultDNSPort] {
			// ip daddr @lan_ranges udp dport 53 drop
			n.conn.AddRule(&nftables.Rule{
				Table: nftCtx.table,
				Chain: chain,
				Exprs: buildRules(
					&expr.Verdict{Kind: expr.VerdictDrop},
					checkIPIsInSet(nftCtx.lanRanges, matchDest),
					checkPortNumber(defaultDNSPort, unix.IPPROTO_UDP, matchDest),
				),
				UserData: userdata.AppendString(nil, userdata.TypeComment, "block to LAN DNS for UDP"),
			})
		}
	}
}

func (n *nft) addMeshPeerToInternet(config firewall.Config, nftCtx *nftContext) *nftables.Chain {
	chain := n.conn.AddChain(&nftables.Chain{
		Name:  meshPeerToInternet,
		Table: nftCtx.table,
	})

	// ip saddr != @allow_peer_traffic_routing drop
	n.conn.AddRule(&nftables.Rule{
		Table: nftCtx.table,
		Chain: chain,
		Exprs: buildRules(
			&expr.Verdict{Kind: expr.VerdictDrop},
			checkIPIsNotInSet(nftCtx.meshRoutingAllowed, matchSource),
		),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "traffic from not allowed peers"),
	})

	if nftCtx.allowlistSubnets != nil {
		// ip daddr @allowed_subnets accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkIPIsInSet(nftCtx.allowlistSubnets, matchDest),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "mesh peer to allowlist IPs"),
		})
	}

	if nftCtx.tcpPorts != nil {
		// tcp dport @ports_tcp accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortIsInSet(nftCtx.tcpPorts, unix.IPPROTO_TCP, matchDest),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "mesh peer to allowlist TCP"),
		})
	}

	if nftCtx.udpPorts != nil {
		// udp dport @ports_udp accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortIsInSet(nftCtx.udpPorts, unix.IPPROTO_UDP, matchDest),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "mesh peer to allowlist UDP"),
		})
	}

	if nftCtx.meshLanAllowedPeers != nil {
		// ip daddr @lan_ranges ip saddr != @peer_local_network_access drop
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictDrop},
				checkIPIsInSet(nftCtx.meshLanAllowedPeers, matchDest),
				checkIPIsNotInSet(nftCtx.meshLanAllowedPeers, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "mesh peer to LAN"),
		})
	}
	// allow traffic thru VPN when connected
	// or when no VPN connected and KS=0, everywhere
	if len(config.TunnelInterface) > 0 || !config.KillSwitch {
		// ip daddr != 100.64.0.0/10 oif "nordtun" accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkIPIsPartOfSubnet(internal.MeshSubnet, matchDest, expr.CmpOpNeq),
				checkInterfaceName(config.TunnelInterface, ifNameOutput, expr.CmpOpEq),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "mesh peer to VPN"),
		})
	}

	// drop all
	n.conn.AddRule((&nftables.Rule{
		Table: nftCtx.table,
		Chain: chain,
		Exprs: buildRules(&expr.Verdict{Kind: expr.VerdictDrop}),
	}))

	return chain
}

func (n *nft) addInternetToMeshPeer(config firewall.Config, nftCtx *nftContext) *nftables.Chain {
	chain := n.conn.AddChain(&nftables.Chain{
		Name:  internetToMeshPeer,
		Table: nftCtx.table,
	})

	// ip daddr != @allow_peer_traffic_routing drop
	n.conn.AddRule(&nftables.Rule{
		Table: nftCtx.table,
		Chain: chain,
		Exprs: buildRules(
			&expr.Verdict{Kind: expr.VerdictDrop},
			checkIPIsNotInSet(nftCtx.meshRoutingAllowed, matchDest),
		),
		UserData: userdata.AppendString(nil, userdata.TypeComment, "traffic to allowed peers"),
	})

	if nftCtx.allowlistSubnets != nil {
		// ip saddr @allowed_subnets accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkIPIsInSet(nftCtx.allowlistSubnets, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "allowlist IPs to mesh peer"),
		})
	}

	if nftCtx.tcpPorts != nil {
		// tcp sport @ports_tcp accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortIsInSet(nftCtx.tcpPorts, unix.IPPROTO_TCP, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "allowlist TCP to mesh peer"),
		})
	}

	if nftCtx.udpPorts != nil {
		// udp sport @ports_udp accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkPortIsInSet(nftCtx.udpPorts, unix.IPPROTO_UDP, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "allowlist UDP to mesh peer"),
		})
	}

	if nftCtx.meshLanAllowedPeers != nil {
		// ip saddr @lan_ranges ip daddr != @peer_local_network_access drop
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictDrop},
				checkIPIsInSet(nftCtx.meshLanAllowedPeers, matchSource),
				checkIPIsNotInSet(nftCtx.meshLanAllowedPeers, matchDest),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "LAN to mesh peer"),
		})
	}
	if len(config.MeshnetInfo.MeshInterface) > 0 {
		// oifname "nordlynx" ct state established,related accept
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: chain,
			Exprs: buildRules(
				&expr.Verdict{Kind: expr.VerdictAccept},
				checkInterfaceName(config.MeshnetInfo.MeshInterface, ifNameOutput, expr.CmpOpEq),
				checkCtState(expr.CtStateBitESTABLISHED|expr.CtStateBitRELATED),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "response to mesh peer"),
		})
	}

	// drop all
	n.conn.AddRule((&nftables.Rule{
		Table: nftCtx.table,
		Chain: chain,
		Exprs: buildRules(&expr.Verdict{Kind: expr.VerdictDrop}),
	}))

	return chain
}

func (n *nft) addMeshnetNat(nftCtx *nftContext) {
	natChain := n.conn.AddChain(&nftables.Chain{
		Name:     meshNatChainName,
		Table:    nftCtx.table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	})

	// ip saddr @allow_peer_traffic_routing ip daddr != 100.64.0.0/10 masquerade
	n.conn.AddRule(&nftables.Rule{
		Table: nftCtx.table,
		Chain: natChain,
		Exprs: buildRules(
			&expr.Masq{},
			checkIPIsInSet(nftCtx.meshRoutingAllowed, matchSource),
			checkIPIsPartOfSubnet(internal.MeshSubnet, matchDest, expr.CmpOpNeq),
		),
	})
}

func (n *nft) addAllowlistNat(config firewall.Config, nftCtx *nftContext) {
	natChain := n.conn.AddChain(&nftables.Chain{
		Name:     allowlistNatChainName,
		Table:    nftCtx.table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	})

	// oifname != "nordlynx" udp sport @udp_allowlist masquerade
	if nftCtx.udpPorts != nil {
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: natChain,
			Exprs: buildRules(
				&expr.Masq{},
				checkInterfaceName(config.TunnelInterface, ifNameOutput, expr.CmpOpNeq),
				checkPortIsInSet(nftCtx.udpPorts, unix.IPPROTO_UDP, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "fix source IP for UDP allowlist ports"),
		})
	}

	// oifname != "nordlynx" tcp sport @tcp_allowlist masquerade
	if nftCtx.tcpPorts != nil {
		n.conn.AddRule(&nftables.Rule{
			Table: nftCtx.table,
			Chain: natChain,
			Exprs: buildRules(
				&expr.Masq{},
				checkInterfaceName(config.TunnelInterface, ifNameOutput, expr.CmpOpNeq),
				checkPortIsInSet(nftCtx.tcpPorts, unix.IPPROTO_TCP, matchSource),
			),
			UserData: userdata.AppendString(nil, userdata.TypeComment, "fix source IP for TCP allowlist ports"),
		})
	}
}

func (n *nft) addLanRangesSet(nftCtx *nftContext) error {
	nftCtx.lanRanges = &nftables.Set{
		Table:    nftCtx.table,
		Name:     lanPrivateIpsSetName,
		KeyType:  nftables.TypeIPAddr,
		Interval: true,
		Constant: true,
	}

	elems, err := convertCidrToSetElements(internal.LocalNetworks)
	if err != nil {
		return err
	}

	if err := n.conn.AddSet(nftCtx.lanRanges, elems); err != nil {
		return err
	}

	return nil
}

func (n *nft) addFilesharePeers(meshMap mesh.MachineMap, nftCtx *nftContext) error {
	nftCtx.fileshareAllowedPeers = &nftables.Set{
		Table:    nftCtx.table,
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

	if err := n.conn.AddSet(nftCtx.fileshareAllowedPeers, elems); err != nil {
		return fmt.Errorf("add fileshare peers set: %w", err)
	}

	return nil
}

func (n *nft) addLanAllowedPeers(meshMap mesh.MachineMap, nftCtx *nftContext) error {
	nftCtx.meshLanAllowedPeers = &nftables.Set{
		Table:    nftCtx.table,
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

	if err := n.conn.AddSet(nftCtx.meshLanAllowedPeers, elems); err != nil {
		return fmt.Errorf("add LAN allowed peers set: %w", err)
	}

	return nil
}

func (n *nft) addAllowedIncomingConnections(meshMap mesh.MachineMap, nftCtx *nftContext) error {
	nftCtx.meshAllowedIncomingConnections = &nftables.Set{
		Table:    nftCtx.table,
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

	if err := n.conn.AddSet(nftCtx.meshAllowedIncomingConnections, elems); err != nil {
		return fmt.Errorf("add peers allowed to connect set: %w", err)
	}

	return nil
}

func (n *nft) addAllowedRoutingPeers(meshMap mesh.MachineMap, nftCtx *nftContext) error {
	nftCtx.meshRoutingAllowed = &nftables.Set{
		Table:    nftCtx.table,
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

	if err := n.conn.AddSet(nftCtx.meshRoutingAllowed, elems); err != nil {
		return fmt.Errorf("add peers allowed to route traffic set: %w", err)
	}

	return nil
}

func (n *nft) addAllowlistSubnets(allowlist config.Allowlist, nftCtx *nftContext) error {
	if len(allowlist.Subnets) == 0 {
		return nil
	}

	nftCtx.allowlistSubnets = &nftables.Set{
		Table:    nftCtx.table,
		Name:     allowlistSubnetsSetName,
		KeyType:  nftables.TypeIPAddr,
		Interval: true,
		Constant: true,
	}

	var elements []nftables.SetElement
	for _, subnet := range allowlist.Subnets {
		startIP, endIP, err := calculateFirstAndLastV4Prefix(subnet)
		if err != nil {
			return fmt.Errorf("parse allowlist IP: %s %w", subnet, err)
		}

		elements = append(elements,
			nftables.SetElement{Key: startIP}, nftables.SetElement{Key: endIP, IntervalEnd: true},
		)
	}
	if err := n.conn.AddSet(nftCtx.allowlistSubnets, elements); err != nil {
		return fmt.Errorf("add allowlist set: %w", err)
	}

	return nil
}

func (n *nft) addMainTable() *nftables.Table {
	return n.conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   tableName,
	})
}
