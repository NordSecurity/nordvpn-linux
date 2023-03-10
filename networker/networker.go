/*
Package networker abstracts network configuration from the rest of the system.
*/
package networker

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/ipv6"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/meshnet/exitnode"
	"github.com/NordSecurity/nordvpn-linux/slices"

	"github.com/kofalt/go-memoize"
)

var (
	// errNilVPN is returned when there is a bug in program logic.
	errNilVPN      = errors.New("vpn is nil")
	errInactiveVPN = errors.New("not connected to vpn")
	// ErrMeshNotActive to report to outside
	ErrMeshNotActive = errors.New("mesh is not active")
	// ErrMeshPeerIsNotRoutable to report to outside
	ErrMeshPeerIsNotRoutable = errors.New("mesh peer is not routable")
	// ErrMeshPeerNotFound to report to outside
	ErrMeshPeerNotFound = errors.New("mesh peer not found")
	defaultMeshSubnet   = netip.MustParsePrefix("100.64.0.0/10")
)

// ConnectionStatus of a currently active connection
type ConnectionStatus struct {
	// State of the vpn. OpenVPN specific.
	State vpn.State
	// Technology, which may or may not match what's in the config
	Technology config.Technology
	// Protocol, which may or may not match what's in the config
	Protocol config.Protocol
	// IP of the other end of the connection
	IP netip.Addr
	// Hostname of the other end of the connection
	Hostname string
	// Country of the other end of the connection
	Country string
	// City of the other end of the connection
	City string
	// Download is the amount of data received through the connection
	Download uint64
	// Upload is the amount of data sent through the connection
	Upload uint64
	// Uptime since the connection start
	Uptime *time.Duration
}

// Networker configures networking for connections.
//
// At the moment interface is designed to support only VPN connections.
type Networker interface {
	Start(
		vpn.Credentials,
		vpn.ServerData,
		config.Whitelist,
		config.DNS,
	) error
	Stop() error      // stop vpn
	UnSetMesh() error // stop meshnet
	SetDNS(nameservers []string) error
	UnsetDNS() error
	IsVPNActive() bool
	ConnectionStatus() (ConnectionStatus, error)
	EnableFirewall() error
	DisableFirewall() error
	EnableRouting()
	DisableRouting()
	SetWhitelist(config.Whitelist) error
	UnsetWhitelist() error
	IsNetworkSet() bool
	SetKillSwitch(config.Whitelist) error
	UnsetKillSwitch() error
	PermitIPv6() error
	DenyIPv6() error
	SetVPN(vpn.VPN)
	LastServerName() string
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
	whitelistRouter    routes.Service
	dnsSetter          dns.Setter
	ipv6               ipv6.Blocker
	fw                 firewall.Service
	devices            device.ListFunc
	policyRouter       routes.PolicyService
	dnsHostSetter      dns.HostnameSetter
	router             routes.Service
	peerRouter         routes.Service
	exitNode           exitnode.Node
	isNetworkSet       bool // used during cleanup
	isKillSwitchSet    bool // used during cleanup
	isV6TrafficAllowed bool // used during cleanup
	isVpnSet           bool // used during cleanup
	isMeshnetSet       bool
	rules              []string // firewall rule names
	nextVPN            vpn.VPN
	cfg                mesh.MachineMap
	whitelist          config.Whitelist
	lastServer         vpn.ServerData
	lastCreds          vpn.Credentials
	startTime          *time.Time
	lastNameservers    []string
	lastPrivateKey     string
	ipv6Enabled        bool
	fwmark             uint32
	mu                 sync.Mutex
}

// NewCombined returns a ready made version of
// Combined.
func NewCombined(
	vpnet vpn.VPN,
	mesh meshnet.Mesh,
	gateway routes.GatewayRetriever,
	publisher events.Publisher[string],
	whitelistRouter routes.Service,
	dnsSetter dns.Setter,
	ipv6 ipv6.Blocker,
	fw firewall.Service,
	devices device.ListFunc,
	policyRouter routes.PolicyService,
	dnsHostSetter dns.HostnameSetter,
	router routes.Service,
	peerRouter routes.Service,
	exitNode exitnode.Node,
	fwmark uint32,
) *Combined {
	return &Combined{
		vpnet:           vpnet,
		mesh:            mesh,
		gateway:         gateway,
		publisher:       publisher,
		whitelistRouter: whitelistRouter,
		dnsSetter:       dnsSetter,
		ipv6:            ipv6,
		fw:              fw,
		devices:         devices,
		policyRouter:    policyRouter,
		dnsHostSetter:   dnsHostSetter,
		router:          router,
		peerRouter:      peerRouter,
		exitNode:        exitNode,
		rules:           []string{},
		fwmark:          fwmark,
	}
}

// Start VPN connection after preparing the network.
func (netw *Combined) Start(
	creds vpn.Credentials,
	serverData vpn.ServerData,
	whitelist config.Whitelist,
	nameservers config.DNS,
) (err error) {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if netw.isConnectedToVPN() {
		return netw.restart(creds, serverData, nameservers)
	}
	return netw.start(creds, serverData, whitelist, nameservers)
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
	creds vpn.Credentials,
	serverData vpn.ServerData,
	whitelist config.Whitelist,
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
	if err = netw.vpnet.Start(creds, serverData); err != nil {
		if err := netw.vpnet.Stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return err
	}

	if !netw.isMeshnetSet {
		netw.publisher.Publish("Setting the routing rules up")
		err = netw.policyRouter.SetupRoutingRules(
			netw.vpnet.Tun().Interface(),
			serverData.IP.Is6(),
		)

		if err != nil {
			return err
		}
	}

	netw.publisher.Publish("starting network configuration")
	// if KillSwitch is turned on, connection is already dropped
	if !netw.isNetworkSet && !netw.isKillSwitchSet {
		err = netw.setNetwork(whitelist)
		if err != nil {
			return err
		}
	}

	if err = netw.resetWhitelist(); err != nil {
		return err
	}

	err = netw.router.Add(routes.Route{
		Subnet:  netip.MustParsePrefix("0.0.0.0/0"),
		Device:  netw.vpnet.Tun().Interface(),
		TableID: netw.policyRouter.TableID(),
	})

	if err != nil {
		return fmt.Errorf("adding the default route: %w", err)
	}

	dnsGetter := &dns.NameServers{}

	if netw.isMeshnetSet && defaultMeshSubnet.Contains(serverData.IP) {
		err = netw.setDNS(dnsGetter.Get(false, false))
	} else {
		err = netw.setDNS(nameservers)
	}
	if err != nil {
		return err
	}

	if netw.isMeshnetSet {
		if err = netw.refresh(netw.cfg); err != nil {
			return fmt.Errorf(
				"refreshing meshnet: %w",
				err,
			)
		}
	}

	netw.isVpnSet = true
	netw.lastServer = serverData
	netw.lastCreds = creds
	netw.lastNameservers = nameservers
	start := time.Now()
	netw.startTime = &start
	return nil
}

func (netw *Combined) restart(
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
	if err = netw.vpnet.Start(creds, serverData); err != nil {
		if err := netw.vpnet.Stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return err
	}

	// after restarting need to restore routing - because tun interface was recreated
	// assuming all other routing rules are left as it was before restart
	if err = netw.router.Add(routes.Route{
		Subnet:  netip.MustParsePrefix("0.0.0.0/0"),
		Device:  netw.vpnet.Tun().Interface(),
		TableID: netw.policyRouter.TableID(),
	}); err != nil {
		return fmt.Errorf("adding the default route: %w", err)
	}

	dnsGetter := &dns.NameServers{}
	if netw.isMeshnetSet && defaultMeshSubnet.Contains(serverData.IP) {
		err = netw.setDNS(dnsGetter.Get(false, false))
	} else {
		err = netw.setDNS(nameservers)
	}
	if err != nil {
		return err
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
	if netw.isVpnSet {
		err := netw.stop()
		if err != nil && !errors.Is(err, errNilVPN) {
			return err
		}
	}
	return nil
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
func (netw *Combined) ConnectionStatus() (ConnectionStatus, error) {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	if !netw.isConnectedToVPN() {
		return ConnectionStatus{}, errInactiveVPN
	}

	stats, err := netw.vpnet.Tun().TransferRates()
	if err != nil {
		return ConnectionStatus{}, err
	}

	tech := config.Technology_OPENVPN
	if netw.vpnet.Tun().Interface().Name == "nordlynx" {
		tech = config.Technology_NORDLYNX
	}

	var uptime *time.Duration
	if netw.startTime != nil {
		dur := time.Since(*netw.startTime)
		uptime = &dur
	}

	return ConnectionStatus{
		State:      vpn.ConnectedState,
		Technology: tech,
		Protocol:   netw.lastServer.Protocol,
		IP:         netw.lastServer.IP,
		Hostname:   netw.lastServer.Hostname,
		Country:    netw.lastServer.Country,
		City:       netw.lastServer.City,
		Download:   stats.Rx,
		Upload:     stats.Tx,
		Uptime:     uptime,
	}, nil
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

	return netw.fw.Add([]firewall.Rule{
		{
			Name:       "drop",
			Direction:  firewall.TwoWay,
			Interfaces: ifaces,
			Allow:      false,
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
			Name:        "vpn_whitelist_icmp6_errors",
			Interfaces:  ifaces,
			Protocols:   []string{"ipv6-icmp"},
			Direction:   firewall.TwoWay,
			Allow:       true,
			Ipv6Only:    true,
			Icmpv6Types: []int{1, 2, 3, 4, 128, 129},
		},
		{
			Name:        "vpn_whitelist_icmp6_address",
			Interfaces:  ifaces,
			Protocols:   []string{"ipv6-icmp"},
			Direction:   firewall.TwoWay,
			Allow:       true,
			Ipv6Only:    true,
			Icmpv6Types: []int{133, 134, 135, 136, 141, 142, 148, 149},
			HopLimit:    255,
		},
		{
			Name:       "vpn_whitelist_icmp6_multicast",
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
		},
		{
			Name:       "vpn_whitelist_dhcp6_in",
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
		},
		{
			Name:       "vpn_whitelist_dhcp6_out",
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
		"vpn_whitelist_icmp6_errors",
		"vpn_whitelist_icmp6_address",
		"vpn_whitelist_icmp6_multicast",
		"vpn_whitelist_dhcp6_in",
		"vpn_whitelist_dhcp6_out",
	})

	if err != nil {
		return err
	}
	netw.isV6TrafficAllowed = false
	return nil
}

func (netw *Combined) resetWhitelist() error {
	// this is done in order to maintain the order of the firewall
	// rules
	if err := netw.unsetWhitelist(); err != nil {
		return fmt.Errorf("unsetting whitelist: %w", err)
	}

	if err := netw.setWhitelist(netw.whitelist); err != nil {
		return fmt.Errorf("re-setting whitelist: %w", err)
	}
	return nil
}

// EnableFirewall activates the firewall and applies the rules
// according to the user's settings. (killswitch, whitelist)
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
	if err := netw.whitelistRouter.Enable(tableID); err != nil {
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
	if err := netw.whitelistRouter.Disable(); err != nil {
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

func (netw *Combined) SetWhitelist(whitelist config.Whitelist) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.setWhitelist(whitelist)
}

func (netw *Combined) setWhitelist(whitelist config.Whitelist) error {
	ifaces, err := netw.devices()
	if err != nil {
		return err
	}

	rules := []firewall.Rule{}
	var subnets []netip.Prefix

	cache := memoize.NewMemoizer(30*time.Second, 1*time.Minute)
	type cachedGateway struct {
		gatewayIP        netip.Addr
		defaultInterface net.Interface
	}

	for cidr := range whitelist.Subnets {
		subnet, err := netip.ParsePrefix(cidr)
		if err != nil {
			// TODO: after Go 1.20, rewrite using error joining
			return fmt.Errorf("parsing subnet CIDR: %w", err)
		}

		// for private network we add only firewall exception
		if subnet.Addr().IsPrivate() {
			subnets = append(subnets, subnet)
			continue
		}

		gatewayRoute, err, _ := cache.Memoize(strconv.FormatBool(subnet.Addr().Is6()),
			func() (interface{}, error) {
				gwIP, gwName, gwErr := netw.gateway.Default(subnet.Addr().Is6())
				return cachedGateway{gwIP, gwName}, gwErr
			})

		if err != nil {
			// if gateway does not exist, we still honour users choice
			subnets = append(subnets, subnet)
			log.Println(internal.WarningPrefix, "whitelisting routes gateway not found for", subnet.String(), err)
			continue
		}

		route := routes.Route{
			Gateway: gatewayRoute.(cachedGateway).gatewayIP,
			Subnet:  subnet,
			Device:  gatewayRoute.(cachedGateway).defaultInterface,
			TableID: netw.policyRouter.TableID(),
		}

		err = netw.whitelistRouter.Add(route)
		if errors.Is(err, routes.ErrRouteToOtherDestinationExists) {
			log.Println(internal.WarningPrefix, "route(s) for whitelisted subnet(s) via non-default gateway already exist in the system")
		}
		if err != nil {
			// TODO: after Go 1.20, rewrite using error joining
			return fmt.Errorf("adding route for subnet %s: %w", route.Subnet, err)
		}

		subnets = append(subnets, subnet)
	}
	if subnets != nil {
		rules = append(rules, firewall.Rule{
			Name:           "whitelist_subnets",
			Interfaces:     ifaces,
			RemoteNetworks: subnets,
			Direction:      firewall.TwoWay,
			Allow:          true,
		})
	}

	for _, pair := range []struct {
		name  string
		ports map[int64]bool
	}{
		{name: "tcp", ports: whitelist.Ports.TCP},
		{name: "udp", ports: whitelist.Ports.UDP},
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
				Name:       "whitelist_ports_" + pair.name,
				Interfaces: ifaces,
				Protocols:  []string{pair.name},
				Direction:  firewall.TwoWay,
				Ports:      ports,
				Allow:      true,
			})
		}
	}

	if err := netw.fw.Add(rules); err != nil {
		return err
	}
	netw.whitelist = whitelist
	return nil
}

func (netw *Combined) UnsetWhitelist() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.unsetWhitelist()
}

func (netw *Combined) unsetWhitelist() error {
	if err := netw.whitelistRouter.Flush(); err != nil {
		return fmt.Errorf("flushing the whitelist router: %w", err)
	}

	for _, rule := range []string{
		"whitelist_subnets",
		"whitelist_ports_tcp",
		"whitelist_ports_udp",
	} {
		err := netw.fw.Delete([]string{rule})
		if err != nil && !errors.Is(err, firewall.ErrRuleNotFound) {
			// TODO: after Go 1.20, rewrite using error joining
			return err
		}
	}
	return nil
}

func (netw *Combined) IsNetworkSet() bool {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.isNetworkSet
}

func (netw *Combined) setNetwork(whitelist config.Whitelist) error {
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
			Name:       "api_whitelist",
			Interfaces: ifaces,
			Direction:  firewall.TwoWay,
			Marks:      []uint32{netw.fwmark},
			Allow:      true,
		},
	}); err != nil {
		return err
	}

	if err := netw.setWhitelist(whitelist); err != nil {
		return err
	}

	netw.isNetworkSet = true
	return nil
}

func (netw *Combined) unsetNetwork() error {
	if err := netw.fw.Delete([]string{"api_whitelist"}); err != nil {
		return err
	}

	err := netw.unblockTraffic()
	if err != nil && !errors.Is(err, firewall.ErrRuleNotFound) {
		return err
	}

	if err := netw.unsetWhitelist(); err != nil {
		return err
	}

	netw.isNetworkSet = false
	return nil
}

func (netw *Combined) SetKillSwitch(whitelist config.Whitelist) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.setKillSwitch(whitelist)
}

func (netw *Combined) setKillSwitch(whitelist config.Whitelist) error {
	if !netw.isNetworkSet {
		if err := netw.setNetwork(whitelist); err != nil {
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
	netw.mu.Lock()
	defer netw.mu.Unlock()
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
		if err = netw.router.Add(routes.Route{
			Subnet:  netip.MustParsePrefix("0.0.0.0/0"),
			Device:  netw.vpnet.Tun().Interface(),
			TableID: netw.policyRouter.TableID(),
		}); err != nil {
			return fmt.Errorf(
				"re-creating the default route: %w",
				err,
			)
		}
	}

	if !netw.isVpnSet {
		if err = netw.policyRouter.SetupRoutingRules(
			netw.mesh.Tun().Interface(),
			false,
		); err != nil {
			return fmt.Errorf(
				"setting routing rules: %w",
				err,
			)
		}
		routingRulesSet = true
	}

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
	if err = netw.defaultMeshBlock(); err != nil {
		return fmt.Errorf("adding default block rule: %w", err)
	}

	if err = netw.allow(cfg.Machine.PublicKey, cfg.Machine.Address); err != nil {
		return fmt.Errorf("allowing to reach self via meshnet: %w", err)
	}

	for _, peer := range cfg.Peers {
		if peer.DoIAllowInbound {
			err = netw.allow(peer.PublicKey, peer.Address)
			if err != nil {
				return err
			}
		}
	}

	err = netw.exitNode.ResetPeers(cfg.Peers)
	if err != nil {
		return err
	}

	hosts := dns.Hosts{dns.Host{
		IP:         cfg.Machine.Address,
		FQDN:       cfg.Machine.Hostname,
		DomainName: strings.TrimSuffix(cfg.Machine.Hostname, ".nord"),
	}}
	hosts = append(hosts, getHostsFromConfig(cfg.Peers)...)
	netw.publisher.Publish("updating mesh dns")
	if err := netw.dnsHostSetter.SetHosts(hosts); err != nil {
		return err
	}

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
		return fmt.Errorf("clearing peer routes: %w", err)
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
		if err := netw.router.Add(routes.Route{
			Subnet:  netip.MustParsePrefix("0.0.0.0/0"),
			Device:  netw.vpnet.Tun().Interface(),
			TableID: netw.policyRouter.TableID(),
		}); err != nil {
			return fmt.Errorf(
				"re-creating the default route: %w",
				err,
			)
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

// AllowIncoming traffic from the uniqueAddress.
func (netw *Combined) AllowIncoming(uniqueAddress meshnet.UniqueAddress) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.allow(uniqueAddress.UID, uniqueAddress.Address)
}

func (netw *Combined) allow(publicKey string, address netip.Addr) error {
	ruleName := publicKey + "-allow-rule-" + address.String()
	rules := []firewall.Rule{{
		Name:      ruleName,
		Direction: firewall.Inbound,
		RemoteNetworks: []netip.Prefix{
			netip.PrefixFrom(address, address.BitLen()),
		},
		Allow: true,
	}}

	for _, rule := range netw.rules {
		if rule == ruleName {
			return fmt.Errorf("allow rule already exist for %s", ruleName)
		}
	}
	if err := netw.fw.Add(rules); err != nil {
		return err
	}

	netw.rules = append(netw.rules, ruleName)
	return nil
}

// Unblock address.
func (netw *Combined) BlockIncoming(uniqueAddress meshnet.UniqueAddress) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.block(uniqueAddress.UID, uniqueAddress.Address)
}

func (netw *Combined) block(publicKey string, address netip.Addr) error {
	ruleName := publicKey + "-allow-rule-" + address.String()
	for index, rule := range netw.rules {
		if rule == ruleName {
			if err := netw.fw.Delete([]string{ruleName}); err != nil {
				return err
			}
			netw.rules = slices.Delete(netw.rules, index, index+1)
			return nil
		}
	}

	return fmt.Errorf("Allow rule does not exist for %s", ruleName)
}

func getHostsFromConfig(peers mesh.MachinePeers) dns.Hosts {
	hosts := make(dns.Hosts, 0, len(peers))
	for _, peer := range peers {
		hosts = append(hosts, dns.Host{
			IP:         peer.Address,
			FQDN:       peer.Hostname,
			DomainName: strings.TrimSuffix(peer.Hostname, ".nord"),
		})
	}
	return hosts
}

func (netw *Combined) ResetRouting(peers mesh.MachinePeers) error {
	return netw.exitNode.ResetPeers(peers)
}

func (netw *Combined) defaultMeshBlock() error {
	defaultMeshBlock := "default-mesh-block"
	defaultMeshAllowEstablished := "default-mesh-allow-established"
	if err := netw.fw.Add([]firewall.Rule{
		// Block all the inbound traffic for the meshnet peers
		{
			Name:           defaultMeshBlock,
			Direction:      firewall.Inbound,
			RemoteNetworks: []netip.Prefix{defaultMeshSubnet},
			Allow:          false,
		},
		// Allow inbound traffic for the existing connections
		// E. g. this device is making some calls to another
		// peer. In such case it should be able to receive
		// the response.
		{
			Name:           defaultMeshAllowEstablished,
			Direction:      firewall.Inbound,
			RemoteNetworks: []netip.Prefix{defaultMeshSubnet},
			ConnectionStates: []firewall.ConnectionState{
				firewall.Related,
				firewall.Established,
			},
			Allow: true,
		},
	}); err != nil {
		return err
	}
	netw.rules = append(netw.rules, defaultMeshBlock)
	netw.rules = append(netw.rules, defaultMeshAllowEstablished)
	return nil
}
