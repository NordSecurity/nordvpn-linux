/*
Package networker abstracts network configuration from the rest of the system.
*/
package networker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/allowlist"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall/forwarder"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/ipv6"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/slices"
)

var (
	// errNilVPN is returned when there is a bug in program logic.
	errNilVPN = errors.New("vpn is nil")
	// ErrMeshNotActive to report to outside
	ErrMeshNotActive = errors.New("mesh is not active")
	// ErrMeshPeerIsNotRoutable to report to outside
	ErrMeshPeerIsNotRoutable = errors.New("mesh peer is not routable")
	// ErrMeshPeerNotFound to report to outside
	ErrMeshPeerNotFound = errors.New("mesh peer not found")
	// ErrNothingToCancel is returned when `Cancel()` is called but there is no in progress
	// connection to be canceled
	ErrNothingToCancel = errors.New("nothing to cancel")
	defaultMeshSubnet  = netip.MustParsePrefix("100.64.0.0/10")
)

// ErrNoSuchRule is returned when networker tried to remove
// a rule, but such rule does not exist
type ErrNoSuchRule struct {
	ruleName string
}

func (e ErrNoSuchRule) Error() string {
	return fmt.Sprintf("allow rule does not exist for %s", e.ruleName)
}

const (
	// a string to be prepended with peers public key and appended with peers ip address to form the internal rule name
	// for allowing the incomig connections
	allowIncomingRule = "-allow-rule-"
	// a string to be prepended with peers public key and appended with peers ip address to form the internal rule name
	// for blocking incoming connections into local networks
	blockLanRule               = "-block-lan-rule-"
	meshnetFirewallRuleComment = "nordvpn-meshnet"
	denyPrivateDNSRule         = "deny-private-dns"
)

// Networker configures networking for connections.
//
// At the moment interface is designed to support only VPN connections.
type Networker interface {
	Start(
		context.Context,
		vpn.Credentials,
		vpn.ServerData,
		config.Allowlist,
		config.DNS,
		bool, // in case mesh peer connect - route to remote peer's LAN or not
	) error
	// Cancel is created instead of using context.Context because `Start` is shared between VPN
	// and meshnet networkers
	Stop() error      // stop vpn
	UnSetMesh() error // stop meshnet
	SetDNS(nameservers []string) error
	UnsetDNS() error
	IsVPNActive() bool
	IsMeshnetActive() bool
	ConnectionStatus() state.ConnectionStatus
	EnableFirewall() error
	DisableFirewall() error
	EnableRouting()
	DisableRouting()
	SetAllowlist(allowlist config.Allowlist) error
	IsNetworkSet() bool
	SetKillSwitch(config.Allowlist) error
	UnsetKillSwitch() error
	PermitIPv6() error
	DenyIPv6() error
	SetVPN(vpn.VPN)
	LastServerName() string
	SetLanDiscovery(bool)
	UnsetFirewall() error
	GetConnectionParameters() (vpn.ServerData, bool)
}

// Combined configures networking for VPN connections.
//
// It is implemented in such a way, that all public methods
// use sync.Mutex and all private ones don't.
type Combined struct {
	vpnet              vpn.VPN
	mesh               meshnet.Mesh
	gateway            routes.GatewayRetriever
	publisher          events.Publisher[string]
	allowlistRouter    routes.Service
	dnsSetter          dns.Setter
	ipv6               ipv6.Blocker
	fw                 firewall.Service
	allowlistRouting   allowlist.Routing
	devices            device.ListFunc
	policyRouter       routes.PolicyService
	dnsHostSetter      dns.HostnameSetter
	router             routes.Service
	peerRouter         routes.Service
	exitNode           forwarder.ForwardChainManager
	isNetworkSet       bool // used during cleanup
	isKillSwitchSet    bool // used during cleanup
	isV6TrafficAllowed bool // used during cleanup
	isVpnSet           bool // used during cleanup
	isMeshnetSet       bool
	rules              []string // firewall rule names
	nextVPN            vpn.VPN
	cfg                mesh.MachineMap
	allowlist          config.Allowlist
	lastServer         vpn.ServerData
	lastCreds          vpn.Credentials
	startTime          *time.Time
	lastNameservers    []string
	lastPrivateKey     string
	ipv6Enabled        bool
	fwmark             uint32
	mu                 sync.Mutex
	lanDiscovery       bool
	// need to memorize route to remote LAN state set on mesh peer connect
	// according how remote peer has set its permission, for later when
	// doing mesh refresh which may happen in background e.g. when network
	// change event happens
	enableLocalTraffic bool
	// list with the existing OS interfaces when VPN was connected.
	// This is used at network changes to know when a new interface was inserted
	interfaces           mapset.Set[string]
	isFilesharePermitted bool
	connectionInfo       *state.ConnectionInfo
	dnsDenied            bool
}

// NewCombined returns a ready made version of
// Combined.
func NewCombined(
	vpnet vpn.VPN,
	mesh meshnet.Mesh,
	gateway routes.GatewayRetriever,
	publisher events.Publisher[string],
	allowlistRouter routes.Service,
	dnsSetter dns.Setter,
	ipv6 ipv6.Blocker,
	fw firewall.Service,
	allowlist allowlist.Routing,
	devices device.ListFunc,
	policyRouter routes.PolicyService,
	dnsHostSetter dns.HostnameSetter,
	router routes.Service,
	peerRouter routes.Service,
	exitNode forwarder.ForwardChainManager,
	fwmark uint32,
	lanDiscovery bool,
	connectionInfo *state.ConnectionInfo,
) *Combined {
	return &Combined{
		vpnet:              vpnet,
		mesh:               mesh,
		gateway:            gateway,
		publisher:          publisher,
		allowlistRouter:    allowlistRouter,
		dnsSetter:          dnsSetter,
		ipv6:               ipv6,
		fw:                 fw,
		allowlistRouting:   allowlist,
		devices:            devices,
		policyRouter:       policyRouter,
		dnsHostSetter:      dnsHostSetter,
		router:             router,
		peerRouter:         peerRouter,
		exitNode:           exitNode,
		rules:              []string{},
		fwmark:             fwmark,
		lanDiscovery:       lanDiscovery,
		enableLocalTraffic: true,
		interfaces:         mapset.NewSet[string](),
		connectionInfo:     connectionInfo,
	}
}

// Start VPN connection after preparing the network.
func (netw *Combined) Start(
	ctx context.Context,
	creds vpn.Credentials,
	serverData vpn.ServerData,
	allowlist config.Allowlist,
	nameservers config.DNS,
	enableLocalTraffic bool,
) (err error) {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	defer netw.updateConnectionStatusAfterStart()
	netw.enableLocalTraffic = enableLocalTraffic
	if netw.isConnectedToVPN() {
		return netw.restart(ctx, creds, serverData, nameservers)
	}
	return netw.start(ctx, creds, serverData, allowlist, nameservers)
}

// updateConnectionStatus builds the [state.ConnectionStatus] and updates it in [Combined].
// In case of an error, empty [state.ConnectionStatus] is set.
//
// Not thread safe.
func (netw *Combined) updateConnectionStatusAfterStart() {
	if !netw.isConnectedToVPN() {
		return
	}

	tech := config.Technology_OPENVPN
	tunnelName := netw.vpnet.Tun().Interface().Name
	if netw.vpnet.Tun().Interface().Name == nordlynx.InterfaceName {
		tech = config.Technology_NORDLYNX
	} else if tunnelName == internal.NordWhisperInterfaceName {
		tech = config.Technology_NORDWHISPER
	}

	actualConnParams, isActive := netw.vpnet.GetConnectionParameters()

	connectionStatus := state.ConnectionStatus{
		State:           pb.ConnectionState_CONNECTED,
		Technology:      tech,
		Protocol:        netw.lastServer.Protocol,
		IP:              netw.lastServer.IP,
		Name:            netw.lastServer.Name,
		Hostname:        netw.lastServer.Hostname,
		Country:         netw.lastServer.Country,
		CountryCode:     netw.lastServer.CountryCode,
		City:            netw.lastServer.City,
		VirtualLocation: netw.lastServer.VirtualLocation,
		PostQuantum:     isActive && actualConnParams.PostQuantum,
		Obfuscated:      isActive && actualConnParams.Obfuscated,
		TunnelName:      tunnelName,
		StartTime:       netw.startTime,
	}

	netw.connectionInfo.SetStatus(connectionStatus)
}

// failureRecover what's possible if vpn start fails
func failureRecover(netw *Combined) {
	if !netw.isMeshnetSet {
		if err := netw.policyRouter.CleanupRouting(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
	}

	if err := netw.router.Flush(); err != nil {
		log.Println(internal.DeferPrefix, err)
	}

	if err := netw.vpnet.Stop(); err != nil {
		log.Println(internal.DeferPrefix, err)
	}

	if netw.isNetworkSet && !netw.isKillSwitchSet {
		if err := netw.unsetNetwork(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
	}

	if netw.isV6TrafficAllowed {
		if err := netw.stopAllowedIPv6Traffic(); err != nil {
			log.Println(internal.DebugPrefix, err)
		}
	}
	netw.isVpnSet = false
}

func (netw *Combined) start(
	ctx context.Context,
	creds vpn.Credentials,
	serverData vpn.ServerData,
	allowlist config.Allowlist,
	nameservers config.DNS,
) (err error) {
	if netw.isVpnSet {
		return errors.New("already started")
	}
	if netw.vpnet == nil {
		return errNilVPN
	}

	defer func() {
		if err != nil {
			failureRecover(netw)
		}
	}()

	netw.publisher.Publish("starting vpn")

	if serverData.IP == (netip.Addr{}) {
		serverData = netw.lastServer
	}
	if err = netw.vpnet.Start(ctx, creds, serverData); err != nil {
		if err := netw.vpnet.Stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return err
	}

	netw.publisher.Publish("Setting the routing rules up")

	// if routing rules were set - they will be adjusted as needed
	if err = netw.policyRouter.SetupRoutingRules(
		serverData.IP.Is6(),
		netw.enableLocalTraffic,
		netw.lanDiscovery,
		allowlist.Subnets,
	); err != nil {
		return err
	}

	if err = netw.configureNetwork(allowlist, serverData, nameservers); err != nil {
		return err
	}

	netw.isVpnSet = true
	netw.lastServer = serverData
	netw.lastCreds = creds
	netw.lastNameservers = nameservers
	start := time.Now()
	netw.startTime = &start
	netw.interfaces = device.InterfacesWithDefaultRoute(mapset.NewSet(netw.vpnet.Tun().Interface().Name))
	return nil
}

func (netw *Combined) configureNetwork(
	allowlist config.Allowlist,
	serverData vpn.ServerData,
	nameservers config.DNS,
) error {
	netw.publisher.Publish("starting network configuration")

	if err := netw.configureFirewall(allowlist); err != nil {
		return err
	}

	if err := netw.addDefaultRoute(); err != nil {
		return err
	}

	if err := netw.configureDNS(serverData, nameservers); err != nil {
		return err
	}

	if netw.isMeshnetSet {
		if err := netw.refresh(netw.cfg); err != nil {
			return fmt.Errorf("refreshing meshnet: %w", err)
		}
	}

	return netw.disableIPv6IfNeeded()
}

func (netw *Combined) disableIPv6IfNeeded() error {
	if !netw.ipv6Enabled {
		if err := netw.denyIPv6(); err != nil {
			return err
		}
	}

	return nil
}

func (netw *Combined) configureDNS(serverData vpn.ServerData, nameservers config.DNS) error {
	dnsGetter := &dns.NameServers{}

	if netw.isMeshnetSet && defaultMeshSubnet.Contains(serverData.IP) {
		return netw.setDNS(dnsGetter.Get(false, false))
	} else {
		return netw.setDNS(nameservers)
	}
}

func (netw *Combined) addDefaultRoute() error {
	err := netw.router.Add(routes.Route{
		Subnet:  netip.MustParsePrefix("0.0.0.0/0"),
		Device:  netw.vpnet.Tun().Interface(),
		TableID: netw.policyRouter.TableID(),
	})
	if err != nil {
		return fmt.Errorf("adding the default route: %w", err)
	}
	return err
}

func (netw *Combined) configureFirewall(allowlist config.Allowlist) error {
	if err := netw.setNetwork(allowlist); err != nil && !errors.Is(err, firewall.ErrRuleAlreadyExists) {
		if !netw.isNetworkSet {
			return err
		} else {
			netw.publisher.Publish("re-setting firewall failed: " + err.Error())
		}
	}

	if err := netw.resetAllowlist(); err != nil {
		return err
	}
	return nil
}

func (netw *Combined) restart(
	ctx context.Context,
	creds vpn.Credentials,
	serverData vpn.ServerData,
	nameservers config.DNS,
) (err error) {
	if netw.vpnet == nil {
		return errNilVPN
	}

	defer func() {
		if err != nil {
			failureRecover(netw)
		}
	}()

	// remove default route
	if err := netw.router.Flush(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	err = netw.vpnet.Stop()
	if err != nil {
		return err
	}

	netw.publisher.Publish("restarting vpn")

	netw.switchToNextVpn()

	if serverData.IP == (netip.Addr{}) {
		serverData = netw.lastServer
	}
	if err = netw.vpnet.Start(ctx, creds, serverData); err != nil {
		if err := netw.vpnet.Stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return err
	}

	// after restarting need to restore routing - because tun interface was recreated
	// assuming all other routing rules are left as it was before restart
	if err = netw.addDefaultRoute(); err != nil {
		return err
	}

	if err := netw.configureDNS(serverData, nameservers); err != nil {
		return err
	}

	if err := netw.disableIPv6IfNeeded(); err != nil {
		log.Println(internal.ErrorPrefix, "failed to disable ipv6", err)
	}

	netw.lastServer = serverData
	netw.lastCreds = creds
	start := time.Now()
	netw.startTime = &start
	return nil
}

// Stop VPN connection and clean up network after it stopped.
func (netw *Combined) Stop() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	defer netw.updateConnectionStatusAfterStop()
	if netw.isVpnSet {
		err := netw.stop()
		if err != nil && !errors.Is(err, errNilVPN) {
			return err
		}

		netw.interfaces = mapset.NewSet[string]()
	}
	return nil
}

func (netw *Combined) updateConnectionStatusAfterStop() {
	if netw.isConnectedToVPN() {
		return
	}

	netw.connectionInfo.SetStatus(state.ConnectionStatus{
		State:     pb.ConnectionState_DISCONNECTED,
		StartTime: nil,
	})
}

func (netw *Combined) stop() error {
	if netw.vpnet == nil {
		return errNilVPN
	}
	netw.publisher.Publish("stopping network configuration")
	if err := netw.ipv6.Unblock(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}
	err := netw.unsetDNS()
	if err != nil {
		return err
	}
	netw.publisher.Publish("removing route to tunnel")
	if !netw.isMeshnetSet {
		if err := netw.policyRouter.CleanupRouting(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
	} else {
		// if routing rules were set - they will be adjusted as needed
		if err = netw.policyRouter.SetupRoutingRules(
			false,
			true, // by default, enableLocalTraffic=true
			netw.lanDiscovery,
			netw.allowlist.Subnets,
		); err != nil {
			return fmt.Errorf("netw stop, adjusting routing rules: %w", err)
		}
	}

	netw.publisher.Publish("removing route to the vpn server")
	if err := netw.router.Flush(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	netw.publisher.Publish("stopping vpn")
	err = netw.vpnet.Stop()
	if err != nil {
		return err
	}
	if !netw.isKillSwitchSet {
		if err = netw.unsetNetwork(); err != nil {
			return fmt.Errorf("unsetting network: %w", err)
		}
	}

	netw.switchToNextVpn()
	netw.isVpnSet = false
	return nil
}

// switchToNextVpn check if VPN technology was changed when connect was in progress
func (netw *Combined) switchToNextVpn() {
	if netw.nextVPN != nil {
		netw.vpnet = netw.nextVPN
		netw.nextVPN = nil
	}
}

// ConnectionStatus get connection information
func (netw *Combined) ConnectionStatus() state.ConnectionStatus {
	return netw.connectionInfo.Status()
}

// LastServerName returns last used server hostname
func (netw *Combined) LastServerName() string {
	return netw.lastServer.Hostname
}

// SetDNS to the given nameservers.
func (netw *Combined) SetDNS(nameservers []string) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if !netw.isConnectedToVPN() {
		return nil
	}

	netw.lastNameservers = nameservers
	return netw.setDNS(nameservers)
}

func (netw *Combined) setDNS(nameservers []string) error {
	err := netw.dnsSetter.Set(netw.vpnet.Tun().Interface().Name, nameservers)
	if err != nil {
		return fmt.Errorf("networker setting dns: %w", err)
	}
	return nil
}

// UnsetDNS to original settings.
func (netw *Combined) UnsetDNS() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if !netw.isConnectedToVPN() {
		return nil
	}
	return netw.unsetDNS()
}

func (netw *Combined) unsetDNS() error {
	err := netw.dnsSetter.Unset(netw.vpnet.Tun().Interface().Name)
	if err != nil {
		return fmt.Errorf("networker unsetting dns: %w", err)
	}
	return nil
}

func (netw *Combined) PermitIPv6() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	netw.ipv6Enabled = true
	return netw.ipv6.Unblock()
}

func (netw *Combined) DenyIPv6() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.denyIPv6()
}

func (netw *Combined) denyIPv6() error {
	netw.ipv6Enabled = false
	if !netw.isNetworkSet {
		return nil
	}
	return netw.ipv6.Block()
}

func (netw *Combined) blockTraffic() error {
	ifaces, err := netw.devices()
	if err != nil {
		return err
	}

	// block PREROUTING & POSTROUTING
	return netw.fw.Add([]firewall.Rule{
		{
			Name:       "drop",
			Direction:  firewall.TwoWay,
			Interfaces: ifaces,
			Allow:      false,
			Physical:   true,
		},
	})
}

func (netw *Combined) unblockTraffic() error {
	return netw.fw.Delete([]string{"drop"})
}

/*
https://tools.ietf.org/html/rfc4890

Error messages that are essential to the establishment and
maintenance of communications:
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 1   -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 2   -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 3   -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 4   -j ACCEPT

Connectivity checking messages:
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 128   -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 129   -j ACCEPT

Address Configuration and Router Selection messages:
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 133 -m hl --hl-eq 255 -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 134 -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 135 -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 136 -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 141 -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 142 -j ACCEPT

Link-Local Multicast Receiver Notification messages:
-6 -A INPUT -s fe80::/10 -p ipv6-icmp --icmpv6-type 130 -j ACCEPT
-6 -A INPUT -s fe80::/10 -p ipv6-icmp --icmpv6-type 131 -j ACCEPT
-6 -A INPUT -s fe80::/10 -p ipv6-icmp --icmpv6-type 132 -j ACCEPT
-6 -A INPUT -s fe80::/10 -p ipv6-icmp --icmpv6-type 143 -j ACCEPT

SEND Certificate Path Notification messages:
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 148 -j ACCEPT
-6 -A INPUT              -p ipv6-icmp --icmpv6-type 149 -j ACCEPT

Multicast Router Discovery messages:
-6 -A INPUT -s fe80::/10 -p ipv6-icmp --icmpv6-type 151 -j ACCEPT
-6 -A INPUT -s fe80::/10 -p ipv6-icmp --icmpv6-type 152 -j ACCEPT
-6 -A INPUT -s fe80::/10 -p ipv6-icmp --icmpv6-type 153 -j ACCEPT

DHCP6
-6 -A INPUT -d fe80::/64 -p udp -m udp --dport 546 -m comment --comment dhcp6 -j ACCEPT
-6 -A OUTPUT -s fe80::/64 -p udp -m udp --dport 547 -m comment --comment dhcp6 -j ACCEPT
*/
func (netw *Combined) allowIPv6Traffic() error {
	ifaces, err := netw.devices()
	if err != nil {
		return err
	}

	err = netw.fw.Add([]firewall.Rule{
		{
			Name:        "vpn_allowlist_icmp6_errors",
			Interfaces:  ifaces,
			Protocols:   []string{"ipv6-icmp"},
			Direction:   firewall.TwoWay,
			Allow:       true,
			Ipv6Only:    true,
			Icmpv6Types: []int{1, 2, 3, 4, 128, 129},
			Physical:    true,
		},
		{
			Name:        "vpn_allowlist_icmp6_address",
			Interfaces:  ifaces,
			Protocols:   []string{"ipv6-icmp"},
			Direction:   firewall.TwoWay,
			Allow:       true,
			Ipv6Only:    true,
			Icmpv6Types: []int{133, 134, 135, 136, 141, 142, 148, 149},
			HopLimit:    255,
			Physical:    true,
		},
		{
			Name:       "vpn_allowlist_icmp6_multicast",
			Interfaces: ifaces,
			LocalNetworks: []netip.Prefix{
				netip.PrefixFrom(netip.AddrFrom16(
					[16]byte{0xfe, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				), 10),
			},
			Protocols:   []string{"ipv6-icmp"},
			Direction:   firewall.TwoWay,
			Allow:       true,
			Ipv6Only:    true,
			Icmpv6Types: []int{130, 131, 132, 143, 151, 152, 153},
			Physical:    true,
		},
		{
			Name:       "vpn_allowlist_dhcp6_in",
			Interfaces: ifaces,
			LocalNetworks: []netip.Prefix{
				netip.PrefixFrom(netip.AddrFrom16(
					[16]byte{0xfe, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				), 10),
			},
			Protocols:        []string{"udp"},
			DestinationPorts: []int{546},
			Direction:        firewall.Inbound,
			Allow:            true,
			Ipv6Only:         true,
			Physical:         true,
		},
		{
			Name:       "vpn_allowlist_dhcp6_out",
			Interfaces: ifaces,
			LocalNetworks: []netip.Prefix{
				netip.PrefixFrom(netip.AddrFrom16(
					[16]byte{0xfe, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				), 10),
			},
			Protocols:        []string{"udp"},
			DestinationPorts: []int{547},
			Direction:        firewall.Outbound,
			Allow:            true,
			Ipv6Only:         true,
			Physical:         true,
		},
	})
	if err != nil {
		return err
	}
	netw.isV6TrafficAllowed = true
	return nil
}

func (netw *Combined) stopAllowedIPv6Traffic() error {
	err := netw.fw.Delete([]string{
		"vpn_allowlist_icmp6_errors",
		"vpn_allowlist_icmp6_address",
		"vpn_allowlist_icmp6_multicast",
		"vpn_allowlist_dhcp6_in",
		"vpn_allowlist_dhcp6_out",
	})
	if err != nil {
		return err
	}
	netw.isV6TrafficAllowed = false
	return nil
}

func (netw *Combined) resetAllowlist() error {
	// this is done in order to maintain the order of the firewall rules
	log.Println(internal.InfoPrefix, "reset allow list")
	if err := netw.unsetAllowlist(); err != nil {
		return fmt.Errorf("unsetting allowlist: %w", err)
	}

	if err := netw.setAllowlist(netw.allowlist); err != nil {
		return fmt.Errorf("re-setting allowlist: %w", err)
	}
	return nil
}

// EnableFirewall activates the firewall and applies the rules
// according to the user's settings. (killswitch, allowlist)
func (netw *Combined) EnableFirewall() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if err := netw.fw.Enable(); err != nil {
		return fmt.Errorf("enabling firewall: %w", err)
	}

	return nil
}

// DisableFirewall turns all firewall operations to noop.
func (netw *Combined) DisableFirewall() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if err := netw.fw.Disable(); err != nil {
		return fmt.Errorf("disabling firewall: %w", err)
	}

	return nil
}

func (netw *Combined) EnableRouting() {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if err := netw.policyRouter.Enable(); err != nil {
		log.Println(internal.WarningPrefix)
	}

	tableID := netw.policyRouter.TableID()
	if err := netw.allowlistRouter.Enable(tableID); err != nil {
		log.Println(internal.WarningPrefix)
	}

	if err := netw.router.Enable(tableID); err != nil {
		log.Println(internal.WarningPrefix)
	}

	if err := netw.peerRouter.Enable(tableID); err != nil {
		log.Println(internal.WarningPrefix)
	}
}

func (netw *Combined) DisableRouting() {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if err := netw.allowlistRouter.Disable(); err != nil {
		log.Println(internal.WarningPrefix)
	}

	if err := netw.router.Disable(); err != nil {
		log.Println(internal.WarningPrefix)
	}

	if err := netw.peerRouter.Disable(); err != nil {
		log.Println(internal.WarningPrefix)
	}

	if err := netw.policyRouter.Disable(); err != nil {
		log.Println(internal.WarningPrefix)
	}
}

func (netw *Combined) SetAllowlist(allowlist config.Allowlist) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	if netw.isNetworkSet {
		if err := netw.unsetAllowlist(); err != nil {
			return err
		}

		if err := netw.setAllowlist(allowlist); err != nil {
			return err
		}
	}

	return nil
}

func (netw *Combined) setAllowlist(allowlist config.Allowlist) error {
	ifaces, err := netw.devices()
	if err != nil {
		return err
	}
	// allow traffic to LAN - only when user enabled lan-discovery
	if netw.lanDiscovery {
		allowlist = addLANPermissions(allowlist)
	}

	// start adding set of rules
	rules := []firewall.Rule{}
	var subnets []netip.Prefix
	for _, cidr := range allowlist.Subnets {
		subnet, err := netip.ParsePrefix(cidr)
		if err != nil {
			return errors.Join(fmt.Errorf("parsing subnet CIDR"), err)
		}

		// For local unicast addresses only firewall rules are added
		if subnet.Addr().IsLinkLocalUnicast() {
			subnets = append(subnets, subnet)
			continue
		}

		subnets = append(subnets, subnet)
	}
	if subnets != nil {
		rules = append(rules, firewall.Rule{
			Name:           "allowlist_subnets",
			Interfaces:     ifaces,
			RemoteNetworks: subnets,
			Direction:      firewall.TwoWay,
			Allow:          true,
			Physical:       true,
		})
	}

	for _, pair := range []struct {
		name  string
		ports map[int64]bool
	}{
		{name: "tcp", ports: allowlist.Ports.TCP},
		{name: "udp", ports: allowlist.Ports.UDP},
	} {
		var ports []int
		for port := range pair.ports {
			if port > math.MaxUint16 {
				continue
			}
			ports = append(ports, int(port))
		}
		if ports != nil {
			rules = append(rules, firewall.Rule{
				Name:       "allowlist_ports_" + pair.name,
				Interfaces: ifaces,
				Protocols:  []string{pair.name},
				Direction:  firewall.TwoWay,
				Ports:      ports,
				Allow:      true,
				Physical:   true,
			})
			if err := netw.allowlistRouting.EnablePorts(ports, pair.name, fmt.Sprintf("%#x", netw.fwmark)); err != nil {
				return errors.Join(fmt.Errorf("enabling allowlist routing"), err)
			}
		}
	}
	if err := netw.fw.Add(rules); err != nil {
		return err
	}

	lanAvailable := netw.lanDiscovery || !netw.isNetworkSet
	if err := netw.exitNode.ResetFirewall(lanAvailable,
		netw.isKillSwitchSet,
		allowlist); err != nil {
		return fmt.Errorf("resseting forward firewall: %w", err)
	}

	// if port 53 is whitelisted - do not add drop-dns rules
	if !allowlist.Ports.TCP[53] && !allowlist.Ports.UDP[53] {
		// disable DNS traffic to private LAN ranges - to prevent DNS leaks
		// when /etc/resolv.conf has nameserver default gateway
		if err := netw.denyDNS(); err != nil {
			return err
		}
	}

	netw.allowlist = allowlist

	// adjust allow subnet routing rules
	if err = netw.policyRouter.SetupRoutingRules(
		false,
		netw.enableLocalTraffic,
		netw.lanDiscovery,
		netw.allowlist.Subnets,
	); err != nil {
		return fmt.Errorf(
			"setting routing rules: %w",
			err,
		)
	}

	return nil
}

func (netw *Combined) unsetAllowlist() error {
	log.Println(internal.InfoPrefix, "unset allow list")
	if err := netw.allowlistRouter.Flush(); err != nil {
		log.Println(internal.WarningPrefix, "flushing allowlist router:", err)
	}

	for _, rule := range []string{
		"allowlist_subnets",
		"allowlist_subnets_forward",
		"allowlist_ports_tcp",
		"allowlist_ports_udp",
	} {
		err := netw.fw.Delete([]string{rule})
		if err != nil && !errors.Is(err, firewall.ErrRuleNotFound) {
			return fmt.Errorf("disabling allowlist firewall rules: %w", err)
		}
	}

	if err := netw.allowlistRouting.Disable(); err != nil {
		return fmt.Errorf("disabling allowlist routing: %w", err)
	}

	if !netw.allowlist.Ports.TCP[53] && !netw.allowlist.Ports.UDP[53] {
		if err := netw.undenyDNS(); err != nil {
			return fmt.Errorf("unsetting deny dns: %w", err)
		}
	}
	return nil
}

func (netw *Combined) IsNetworkSet() bool {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.isNetworkSet
}

func (netw *Combined) setNetwork(allowlist config.Allowlist) error {
	err := netw.blockTraffic()
	if err != nil && !errors.Is(err, firewall.ErrRuleAlreadyExists) {
		return err
	}

	ifaces, err := netw.devices()
	if err != nil {
		return err
	}

	if err := netw.fw.Add([]firewall.Rule{
		{
			Name:       "api_allowlist",
			Interfaces: ifaces,
			Direction:  firewall.TwoWay,
			Marks:      []uint32{netw.fwmark},
			Allow:      true,
			Physical:   true,
		},
	}); err != nil {
		return err
	}

	if err := netw.setAllowlist(allowlist); err != nil {
		return err
	}

	if err := netw.exitNode.ResetFirewall(netw.lanDiscovery,
		true,
		netw.allowlist); err != nil {
		log.Println(internal.ErrorPrefix,
			"failed to reset peers firewall rules after enabling killswitch: ",
			err)
	}

	netw.isNetworkSet = true
	return nil
}

func (netw *Combined) UnsetFirewall() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	if netw.isKillSwitchSet {
		return nil
	}
	return netw.fw.Flush()
}

func (netw *Combined) unsetNetwork() error {
	if err := netw.fw.Delete([]string{"api_allowlist"}); err != nil {
		return err
	}

	err := netw.unblockTraffic()
	if err != nil && !errors.Is(err, firewall.ErrRuleNotFound) {
		return err
	}

	if err := netw.unsetAllowlist(); err != nil {
		return err
	}

	// Passing true because LAN is always available when network is unset
	if err := netw.exitNode.ResetFirewall(true, false, netw.allowlist); err != nil {
		log.Println(internal.ErrorPrefix,
			"failed to reset peers firewall rules after disabling killswitch: ",
			err)
	}

	netw.isNetworkSet = false
	return nil
}

func (netw *Combined) SetKillSwitch(allowlist config.Allowlist) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.setKillSwitch(allowlist, false)
}

func (netw *Combined) setKillSwitch(allowlist config.Allowlist, force bool) error {
	if !netw.isNetworkSet || force {
		if err := netw.setNetwork(allowlist); err != nil {
			return err
		}
	}
	netw.isKillSwitchSet = true
	return nil
}

func (netw *Combined) UnsetKillSwitch() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.unsetKillSwitch()
}

func (netw *Combined) unsetKillSwitch() error {
	if !netw.isVpnSet {
		if err := netw.unsetNetwork(); err != nil {
			return err
		}
	}

	netw.isKillSwitchSet = false
	return nil
}

func (netw *Combined) SetVPN(v vpn.VPN) {
	if !netw.vpnet.IsActive() {
		netw.vpnet = v
	} else {
		netw.nextVPN = v
	}
}

// Refresh peer list.
func (netw *Combined) Refresh(c mesh.MachineMap) error {
	log.Println(internal.DebugPrefix+"mesh-refresh", "lock netw mutex")
	netw.mu.Lock()
	defer func() {
		log.Println(internal.DebugPrefix+"mesh-refresh", "unlock netw mutex")
		netw.mu.Unlock()
	}()
	return netw.refresh(c)
}

func (netw *Combined) SetMesh(
	cfg mesh.MachineMap,
	self netip.Addr,
	privateKey string,
) (err error) {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.setMesh(cfg, self, privateKey)
}

func (netw *Combined) setMesh(
	cfg mesh.MachineMap,
	self netip.Addr,
	privateKey string,
) (err error) {
	if netw.isMeshnetSet {
		return errors.New("meshnet already set")
	}
	routingRulesSet := false
	defer func() {
		if err != nil {
			if routingRulesSet {
				if err := netw.policyRouter.CleanupRouting(); err != nil {
					log.Println(internal.DeferPrefix, err)
				}
			}

			if err := netw.defaultMeshUnBlock(); err != nil {
				log.Println(internal.DeferPrefix, err)
			}

			if err := netw.dnsHostSetter.UnsetHosts(); err != nil {
				log.Println(internal.DeferPrefix, err)
			}

			if err := netw.exitNode.Disable(); err != nil {
				log.Println(internal.DeferPrefix, err)
			}

			if err := netw.peerRouter.Flush(); err != nil {
				log.Println(internal.DeferPrefix, err)
			}

			if err := netw.mesh.Disable(); err != nil {
				log.Println(internal.DeferPrefix, err)
			}
		}
	}()

	// If network is started, default might (in libtelio case will)
	// be destroyed, therefore it's safe just to flush it here
	if netw.isVpnSet {
		if err := netw.router.Flush(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
	}

	if err = netw.mesh.Enable(self, privateKey); err != nil {
		if netw.isVpnSet && !netw.mesh.IsActive() {
			netw.isVpnSet = false // prevents already connected error
			return meshnet.ErrTunnelClosed
		}
		return fmt.Errorf("enabling meshnet: %w", err)
	}

	if netw.isVpnSet {
		if err = netw.addDefaultRoute(); err != nil {
			return err
		}
	}

	if err = netw.policyRouter.SetupRoutingRules(
		false,
		netw.enableLocalTraffic,
		netw.lanDiscovery,
		netw.allowlist.Subnets,
	); err != nil {
		return fmt.Errorf(
			"setting routing rules: %w",
			err,
		)
	}
	routingRulesSet = true

	// add routes for new peers and remove for the old ones
	netw.publisher.Publish("adding mesh route")
	if err := netw.peerRouter.Add(routes.Route{
		Subnet:  defaultMeshSubnet,
		Device:  netw.mesh.Tun().Interface(),
		TableID: netw.policyRouter.TableID(),
	}); err != nil {
		return fmt.Errorf(
			"creating default mesh route: %w",
			err,
		)
	}

	err = netw.refresh(cfg)
	if err != nil {
		return err
	}

	netw.isMeshnetSet = true
	netw.lastPrivateKey = privateKey

	return nil
}

func (netw *Combined) refresh(cfg mesh.MachineMap) error {
	log.Println(internal.DebugPrefix+"mesh-refresh", "start refreshing meshnet")
	if err := netw.defaultMeshUnBlock(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	if err := netw.dnsHostSetter.UnsetHosts(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	if err := netw.exitNode.Disable(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	if err := netw.exitNode.Enable(); err != nil {
		return fmt.Errorf("enabling exit node: %w", err)
	}

	if err := netw.mesh.Refresh(cfg); err != nil {
		return fmt.Errorf("refreshing mesh: %w", err)
	}
	netw.cfg = cfg

	var err error
	if err = netw.defaultMeshBlock(cfg.Machine.Address); err != nil {
		return fmt.Errorf("adding default block rule: %w", err)
	}

	if err = netw.allowIncoming(cfg.Machine.PublicKey, cfg.Machine.Address, true); err != nil {
		return fmt.Errorf("allowing to reach self via meshnet: %w", err)
	}

	for _, peer := range cfg.Peers {
		if !peer.Address.IsValid() {
			continue
		}

		lanAllowed := peer.DoIAllowRouting && peer.DoIAllowLocalNetwork

		if peer.DoIAllowInbound {
			err = netw.allowIncoming(peer.PublicKey, peer.Address, lanAllowed)
			if err != nil {
				return fmt.Errorf("allowing inbound traffic for peer: %w", err)
			}
		}

		if peer.DoIAllowFileshare {
			err = netw.allowFileshare(peer.PublicKey, peer.Address)
			if err != nil {
				return fmt.Errorf("allowing fileshare for peer: %w", err)
			}
		}

		// TODO (LVPN-4031): detect which peer we are connected (if connected)
		// to and check if maybe allowLocalAccess permission has changed and
		// if so, change routing to route to local LAN
	}

	lanAvailable := netw.lanDiscovery || !netw.isNetworkSet
	err = netw.exitNode.ResetPeers(cfg.Peers,
		lanAvailable,
		netw.isKillSwitchSet,
		netw.allowlist)
	if err != nil {
		return err
	}

	var hostName string
	var domainNames []string

	if cfg.Machine.Nickname != "" {
		hostName = cfg.Machine.Nickname
		domainNames = []string{
			cfg.Machine.Nickname + ".nord",
			cfg.Machine.Hostname,
			strings.TrimSuffix(cfg.Machine.Hostname, ".nord"),
		}
	} else {
		hostName = cfg.Machine.Hostname
		domainNames = []string{strings.TrimSuffix(cfg.Machine.Hostname, ".nord")}
	}

	hosts := dns.Hosts{dns.Host{
		IP:          cfg.Machine.Address,
		FQDN:        hostName,
		DomainNames: domainNames,
	}}
	hosts = append(hosts, getHostsFromConfig(cfg.Peers)...)
	netw.publisher.Publish("updating mesh dns")
	if err := netw.dnsHostSetter.SetHosts(hosts); err != nil {
		return err
	}

	log.Println(internal.DebugPrefix+"mesh-refresh", "done refreshing meshnet")
	netw.publisher.Publish("refreshing mesh")
	return nil
}

func (netw *Combined) defaultMeshUnBlock() error {
	err := netw.fw.Delete(netw.rules)
	if err != nil {
		return err
	}
	netw.rules = nil
	return nil
}

func (netw *Combined) UnSetMesh() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	// clear this flag only when user turns mesh off, cannot do that in internal func
	// because it is used during refresh, and during refresh we need to know what
	// was set before i.e. during mesh peer connect
	netw.enableLocalTraffic = true
	return netw.unSetMesh()
}

func (netw *Combined) unSetMesh() error {
	if !netw.isMeshnetSet {
		return ErrMeshNotActive
	}
	if err := netw.dnsHostSetter.UnsetHosts(); err != nil {
		return fmt.Errorf("unsetting hosts: %w", err)
	}

	if err := netw.defaultMeshUnBlock(); err != nil {
		return fmt.Errorf(
			"unblocking the peer subnet: %w",
			err,
		)
	}

	if err := netw.exitNode.Disable(); err != nil {
		return fmt.Errorf(
			"disabling exit node: %w",
			err,
		)
	}

	if !netw.isVpnSet {
		if err := netw.policyRouter.CleanupRouting(); err != nil {
			return fmt.Errorf(
				"cleaning up routing: %w",
				err,
			)
		}
	}

	if err := netw.peerRouter.Flush(); err != nil {
		log.Println(internal.WarningPrefix, "clearing peer routes:", err)
	}

	// If network is started, default might (in libtelio case will)
	// be destroyed, therefore it's safe just to flush it here
	if netw.isVpnSet {
		if err := netw.router.Flush(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
	}

	if err := netw.mesh.Disable(); err != nil {
		return fmt.Errorf("disabling the meshnet: %w", err)
	}

	if netw.isVpnSet {
		if err := netw.addDefaultRoute(); err != nil {
			return err
		}
	}

	netw.isMeshnetSet = false
	return nil
}

func (netw *Combined) StatusMap() (map[string]string, error) {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.mesh.StatusMap()
}

func (netw *Combined) allowIncoming(publicKey string, address netip.Addr, lanAllowed bool) error {
	rules := []firewall.Rule{}

	ruleName := publicKey + allowIncomingRule + address.String()
	rule := firewall.Rule{
		Name:      ruleName,
		Direction: firewall.Inbound,
		RemoteNetworks: []netip.Prefix{
			netip.PrefixFrom(address, address.BitLen()),
		},
		Allow:   true,
		Comment: meshnetFirewallRuleComment,
	}
	rules = append(rules, rule)

	ruleIndex := slices.Index(netw.rules, ruleName)

	if ruleIndex != -1 {
		return fmt.Errorf("allow rule already exist for %s", ruleName)
	}

	if !lanAllowed {
		ruleName := publicKey + blockLanRule + address.String()
		rule := firewall.Rule{
			Name:      ruleName,
			Direction: firewall.Inbound,
			LocalNetworks: []netip.Prefix{
				netip.MustParsePrefix("10.0.0.0/8"),
				netip.MustParsePrefix("172.16.0.0/12"),
				netip.MustParsePrefix("192.168.0.0/16"),
				netip.MustParsePrefix("169.254.0.0/16"),
			},
			RemoteNetworks: []netip.Prefix{
				netip.PrefixFrom(address, address.BitLen()),
			},
			Allow:   false,
			Comment: meshnetFirewallRuleComment,
		}

		rules = append(rules, rule)
		netw.rules = append(netw.rules, ruleName)
	}

	if err := netw.fw.Add(rules); err != nil {
		return fmt.Errorf("adding allow-incoming rule to firewall: %w", err)
	}

	netw.rules = append(netw.rules, ruleName)
	return nil
}

func (netw *Combined) AllowFileshare(uniqueAddress meshnet.UniqueAddress) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.allowFileshare(uniqueAddress.UID, uniqueAddress.Address)
}

func (netw *Combined) allowFileshare(publicKey string, address netip.Addr) error {
	if !netw.isFilesharePermitted {
		log.Println(internal.WarningPrefix, "fileshare is not permitted, can't add allow rules")
		return nil
	}

	ruleName := publicKey + "-allow-fileshare-rule-" + address.String()
	rules := []firewall.Rule{{
		Name:           ruleName,
		Direction:      firewall.Inbound,
		Protocols:      []string{"tcp"},
		Ports:          []int{49111},
		PortsDirection: firewall.Destination,
		RemoteNetworks: []netip.Prefix{
			netip.PrefixFrom(address, address.BitLen()),
		},
		Allow:   true,
		Comment: meshnetFirewallRuleComment,
	}}

	ruleIndex := slices.Index(netw.rules, ruleName)

	if ruleIndex != -1 {
		return fmt.Errorf("allow rule already exist for %s", ruleName)
	}

	if err := netw.fw.Add(rules); err != nil {
		return fmt.Errorf("adding allow-fileshare rule to firewall: %w", err)
	}

	netw.rules = append(netw.rules, ruleName)
	return nil
}

func (netw *Combined) PermitFileshare() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if netw.isFilesharePermitted {
		return nil
	}
	netw.isFilesharePermitted = true
	return netw.allowFileshareAll()
}

func (netw *Combined) allowFileshareAll() error {
	var allErrors []error
	for _, peer := range netw.cfg.Peers {
		if peer.DoIAllowFileshare {
			err := netw.allowFileshare(peer.PublicKey, peer.Address)
			allErrors = append(allErrors, err)
		}
	}
	return errors.Join(allErrors...)
}

func (netw *Combined) undenyDNS() error {
	if !netw.dnsDenied {
		log.Println(internal.DebugPrefix, "attemtpt to undeny dns when it was not previously denied")
		return nil
	}

	if err := netw.fw.Delete([]string{denyPrivateDNSRule}); err != nil {
		return fmt.Errorf("deleting deny-private-dns dns rule: %w", err)
	}

	netw.dnsDenied = false

	return nil
}

func (netw *Combined) denyDNS() error {
	if netw.dnsDenied {
		log.Println(internal.DebugPrefix, "attemtpt to deny dns when it was already denied")
		return nil
	}

	rules := []firewall.Rule{{
		Name:           denyPrivateDNSRule,
		Direction:      firewall.Outbound,
		Protocols:      []string{"udp", "tcp"},
		Ports:          []int{53},
		PortsDirection: firewall.Destination,
		RemoteNetworks: []netip.Prefix{
			netip.MustParsePrefix("10.0.0.0/8"),
			netip.MustParsePrefix("172.16.0.0/12"),
			netip.MustParsePrefix("192.168.0.0/16"),
			netip.MustParsePrefix("169.254.0.0/16"),
		},
		Allow:    false,
		Physical: true,
	}}

	if err := netw.fw.Add(rules); err != nil {
		return fmt.Errorf("adding deny-private-dns rule to firewall: %w", err)
	}

	netw.dnsDenied = true

	return nil
}

func (netw *Combined) blockIncoming(uniqueAddress meshnet.UniqueAddress) error {
	lanRuleName := uniqueAddress.UID + blockLanRule + uniqueAddress.Address.String()
	if slices.Index(netw.rules, lanRuleName) != -1 {
		if err := netw.removeRule(lanRuleName); err != nil {
			return err
		}
	}

	ruleName := uniqueAddress.UID + allowIncomingRule + uniqueAddress.Address.String()
	return netw.removeRule(ruleName)
}

func (netw *Combined) blockFileshare(publicKey string, address netip.Addr) error {
	if !netw.isFilesharePermitted {
		log.Println(internal.WarningPrefix, "fileshare is already forbidden")
		return nil
	}
	ruleName := publicKey + "-allow-fileshare-rule-" + address.String()
	return netw.removeRule(ruleName)
}

func (netw *Combined) removeRule(ruleName string) error {
	ruleIndex := slices.Index(netw.rules, ruleName)

	if ruleIndex == -1 {
		return ErrNoSuchRule{ruleName}
	}

	if err := netw.fw.Delete([]string{ruleName}); err != nil {
		return err
	}
	netw.rules = slices.Delete(netw.rules, ruleIndex, ruleIndex+1)

	return nil
}

func (netw *Combined) ForbidFileshare() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if !netw.isFilesharePermitted {
		return nil
	}

	err := netw.blockFileshareAll()
	// NOTE: Mark fileshare as forbidden only when there was no error here, so it
	// can be tried again.
	if err == nil {
		netw.isFilesharePermitted = false
	}

	return err
}

func (netw *Combined) blockFileshareAll() error {
	var allErrors []error
	for _, peer := range netw.cfg.Peers {
		err := netw.blockFileshare(peer.PublicKey, peer.Address)
		// NOTE: It's fine to have the rule already removed which returns [ErrNoSuchRule].
		// It's not fine to have any other errors, so keep those.
		if !errors.Is(err, ErrNoSuchRule{}) {
			allErrors = append(allErrors, err)
		}
	}
	return errors.Join(allErrors...)
}

func getHostsFromConfig(peers mesh.MachinePeers) dns.Hosts {
	hosts := make(dns.Hosts, 0, len(peers))
	for _, peer := range peers {
		if peer.Address.IsValid() {
			var hostName string
			var domainNames []string

			if peer.Nickname != "" {
				hostName = peer.Nickname
				domainNames = []string{
					peer.Nickname + ".nord",
					peer.Hostname,
					strings.TrimSuffix(peer.Hostname, ".nord"),
				}
			} else {
				hostName = peer.Hostname
				domainNames = []string{strings.TrimSuffix(peer.Hostname, ".nord")}
			}

			hosts = append(hosts, dns.Host{
				IP:          peer.Address,
				FQDN:        hostName,
				DomainNames: domainNames,
			})
		}
	}
	return hosts
}

func (netw *Combined) refreshIncoming(peer mesh.MachinePeer) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	if !peer.DoIAllowInbound {
		return nil
	}

	address := meshnet.UniqueAddress{
		UID: peer.PublicKey, Address: peer.Address,
	}

	if slices.Index(netw.rules, peer.PublicKey+allowIncomingRule+peer.Address.String()) != -1 {
		if err := netw.blockIncoming(address); err != nil {
			return fmt.Errorf("blocking incoming traffic: %w", err)
		}
	}

	if err := netw.allowIncoming(address.UID, address.Address, peer.DoIAllowRouting && peer.DoIAllowLocalNetwork); err != nil {
		return fmt.Errorf("allowing incoming traffic: %w", err)
	}

	return nil
}

func (netw *Combined) ResetRouting(peer mesh.MachinePeer, peers mesh.MachinePeers) error {
	lanAvailable := netw.lanDiscovery || !netw.isNetworkSet
	if err := netw.exitNode.ResetPeers(peers,
		lanAvailable,
		netw.isKillSwitchSet,
		netw.allowlist); err != nil {
		return err
	}

	return netw.refreshIncoming(peer)
}

func (netw *Combined) defaultMeshBlock(ip netip.Addr) error {
	defaultMeshBlock := "default-mesh-block"
	defaultMeshAllowEstablished := "default-mesh-allow-established"
	if err := netw.fw.Add([]firewall.Rule{
		// Block all the inbound traffic for the meshnet peers
		{
			Name:           defaultMeshBlock,
			Direction:      firewall.Inbound,
			RemoteNetworks: []netip.Prefix{defaultMeshSubnet},
			Allow:          false,
			Comment:        meshnetFirewallRuleComment,
		},
		// Allow inbound traffic for the existing connections
		// E. g. this device is making some calls to another
		// peer. In such case it should be able to receive
		// the response.
		{
			Name:           defaultMeshAllowEstablished,
			Direction:      firewall.Inbound,
			RemoteNetworks: []netip.Prefix{defaultMeshSubnet},
			ConnectionStates: firewall.ConnectionStates{
				SrcAddr: ip,
				States: []firewall.ConnectionState{
					firewall.Related,
					firewall.Established,
				},
			},
			Allow:   true,
			Comment: meshnetFirewallRuleComment,
		},
	}); err != nil {
		return err
	}
	netw.rules = append(netw.rules, defaultMeshBlock)
	netw.rules = append(netw.rules, defaultMeshAllowEstablished)
	return nil
}

func (netw *Combined) SetLanDiscovery(enabled bool) {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	netw.lanDiscovery = enabled

	lanAvailable := netw.lanDiscovery || !netw.isNetworkSet

	// if routing rules were set - they will be adjusted as needed
	if netw.isMeshnetSet || netw.isVpnSet {
		if err := netw.policyRouter.SetupRoutingRules(
			netw.lastServer.IP.Is6(),
			netw.enableLocalTraffic,
			netw.lanDiscovery,
			netw.allowlist.Subnets,
		); err != nil {
			log.Println(
				internal.ErrorPrefix,
				"failed to set routing rules up after enabling lan discovery:",
				err,
			)
		}
	}

	if err := netw.exitNode.ResetFirewall(lanAvailable,
		netw.isKillSwitchSet,
		netw.allowlist); err != nil {
		log.Println(internal.ErrorPrefix,
			"failed to reset peers firewall rules after enabling lan discovery:",
			err)
	}
}

func (netw *Combined) GetConnectionParameters() (vpn.ServerData, bool) {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	return netw.vpnet.GetConnectionParameters()
}
