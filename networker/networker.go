/*
Package networker abstracts network configuration from the rest of the system.
*/
package networker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/netip"
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
	"github.com/NordSecurity/nordvpn-linux/kernel"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	mapset "github.com/deckarep/golang-set/v2"
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
	ArpIgnoreParamName = "net.ipv4.conf.all.arp_ignore"
	IpForwardParamName = "net.ipv4.ip_forward" // used for meshnet to have routing thru current machine
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
		events.DisconnectCallback, // callback provided by the caller in case Networker disconnects internally
	) error
	// Cancel is created instead of using context.Context because `Start` is shared between VPN
	// and meshnet networkers
	Stop() error      // stop vpn
	UnSetMesh() error // stop meshnet
	SetDNS(nameservers []string) error
	UnsetDNS() error
	IsVPNActive() bool
	IsMeshnetActive() bool
	EnableFirewall() error
	DisableFirewall() error
	EnableRouting()
	DisableRouting()
	SetAllowlist(allowlist config.Allowlist) error
	IsNetworkSet() bool
	SetKillSwitch() error
	UnsetKillSwitch() error
	SetVPN(vpn.VPN)
	LastServerName() string
	SetLanDiscovery(bool)
	UnsetFirewall() error
	GetConnectionParameters() (vpn.ServerData, bool)
	SetARPIgnore(bool) error
}

// Combined configures networking for VPN connections.
//
// It is implemented in such a way, that all public methods
// use sync.Mutex and all private ones don't.
type Combined struct {
	vpnet           vpn.VPN
	mesh            meshnet.Mesh
	gateway         routes.GatewayRetriever
	publisher       events.Publisher[string]
	allowlistRouter routes.Service
	dnsSetter       dns.Setter
	fw              firewall.Service
	devices         device.ListFunc
	policyRouter    routes.PolicyService
	dnsHostSetter   dns.HostnameSetter
	router          routes.Service
	peerRouter      routes.Service
	isNetworkSet    bool // used during cleanup
	isVpnSet        bool // used during cleanup
	isMeshnetSet    bool
	nextVPN         vpn.VPN
	cfg             mesh.MachineMap
	lastServer      vpn.ServerData
	lastCreds       vpn.Credentials
	lastNameservers []string
	lastPrivateKey  string
	fwmark          uint32
	mu              sync.Mutex
	lanDiscovery    bool
	// need to memorize route to remote LAN state set on mesh peer connect
	// according how remote peer has set its permission, for later when
	// doing mesh refresh which may happen in background e.g. when network
	// change event happens
	enableLocalTraffic bool
	// list with the existing OS interfaces when VPN was connected.
	// This is used at network changes to know when a new interface was inserted
	interfaces mapset.Set[string]
	// dnsDenied            bool
	ipv6Blocker     ipv6.Blocker
	ignoreARP       bool
	arpIgnoreSetter kernel.SysctlSetter
	allowlist       config.Allowlist
	isKillSwitchSet bool
	fwConfig        firewall.Config
	ipForwardSetter kernel.SysctlSetter
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
	fw firewall.Service,
	devices device.ListFunc,
	policyRouter routes.PolicyService,
	dnsHostSetter dns.HostnameSetter,
	router routes.Service,
	peerRouter routes.Service,
	fwmark uint32,
	lanDiscovery bool,
	ipv6Blocker ipv6.Blocker,
	ignoreARP bool,
	arpIgnoreSetter kernel.SysctlSetter,
	allowlist config.Allowlist,
	ipForwardSetter kernel.SysctlSetter,
) *Combined {
	return &Combined{
		vpnet:              vpnet,
		mesh:               mesh,
		gateway:            gateway,
		publisher:          publisher,
		allowlistRouter:    allowlistRouter,
		dnsSetter:          dnsSetter,
		fw:                 fw,
		devices:            devices,
		policyRouter:       policyRouter,
		dnsHostSetter:      dnsHostSetter,
		router:             router,
		peerRouter:         peerRouter,
		fwmark:             fwmark,
		lanDiscovery:       lanDiscovery,
		enableLocalTraffic: true,
		interfaces:         mapset.NewSet[string](),
		ipv6Blocker:        ipv6Blocker,
		ignoreARP:          ignoreARP,
		arpIgnoreSetter:    arpIgnoreSetter,
		allowlist:          allowlist,
		fwConfig:           firewall.Config{Allowlist: allowlist},
		ipForwardSetter:    ipForwardSetter,
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
	disconnectCallback events.DisconnectCallback,
) (err error) {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	netw.allowlist = allowlist
	netw.enableLocalTraffic = enableLocalTraffic
	if netw.isConnectedToVPN() {
		return netw.restart(ctx, creds, serverData, nameservers, disconnectCallback)
	}
	return netw.start(ctx, creds, serverData, allowlist, nameservers)
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

	cfg := netw.fwConfig.CopyWith(
		firewall.WithTunnelInterface(""),
	)
	if err := netw.configureFirewall(cfg); err != nil {
		log.Println(internal.DeferPrefix, err)
	}

	netw.unblockIPv6()

	if err := netw.arpIgnoreSetter.Unset(); err != nil {
		log.Println(internal.DebugPrefix, "unsetting arp ignore when recovering from failure:", err)
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

	// Always disable IPv6 with sysctl in the system
	// We also do that when there is a change in network interfaces
	if err = netw.ipv6Blocker.Block(); err != nil {
		log.Println(internal.ErrorPrefix, "Failed to block ipv6 during start using sysctl ", err)
		return err
	}

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
		netw.enableLocalTraffic,
		netw.lanDiscovery,
		allowlist.Subnets,
	); err != nil {
		return err
	}

	if err = netw.configureNetwork(serverData, nameservers); err != nil {
		return err
	}

	if netw.ignoreARP {
		if err := netw.arpIgnoreSetter.Set(); err != nil {
			return fmt.Errorf("setting arp ignore: %w", err)
		}
	}

	tunnelInterface := netw.vpnet.Tun().Interface().Name
	// configure firewall
	// update tunnel name and allowlist
	// use netw.allowlist because it is populated in setAllowlist and there will be
	// updated to also contain LAN addresses for LAN discovery enabled
	newCfg := netw.fwConfig.CopyWith(
		firewall.WithTunnelInterface(tunnelInterface),
		firewall.WithAllowlist(netw.allowlist),
	)
	if err := netw.configureFirewall(newCfg); err != nil {
		return fmt.Errorf("configuring firewall: %w", err)
	}

	netw.isVpnSet = true
	netw.lastServer = serverData
	netw.lastCreds = creds
	netw.lastNameservers = nameservers

	netw.interfaces = device.OutsideCapableTrafficIfNames(mapset.NewSet(tunnelInterface))
	return nil
}

func (netw *Combined) configureNetwork(
	serverData vpn.ServerData,
	nameservers config.DNS,
) error {
	netw.publisher.Publish("starting network configuration")

	if err := netw.setNetwork(netw.allowlist); err != nil {
		if !netw.isNetworkSet {
			return err
		}
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

	return nil
}

func (netw *Combined) configureDNS(serverData vpn.ServerData, nameservers config.DNS) error {
	dnsGetter := &dns.NameServers{}

	if netw.isMeshnetSet && internal.MeshSubnet.Contains(serverData.IP) {
		return netw.setDNS(dnsGetter.Get(false))
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

func (netw *Combined) restart(
	ctx context.Context,
	creds vpn.Credentials,
	serverData vpn.ServerData,
	nameservers config.DNS,
	disconnectCallback events.DisconnectCallback,
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

	stopStartTime := time.Now()
	err = netw.vpnet.Stop()
	disconnectCallback(stopStartTime, err)
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

	// Always disable IPv6 with sysctl in the system
	// We also do that when there is a change in network interfaces
	if err = netw.ipv6Blocker.Block(); err != nil {
		log.Println(internal.ErrorPrefix, "Failed to block ipv6 during restart using sysctl ", err)
		return err
	}

	if netw.ignoreARP {
		if err := netw.arpIgnoreSetter.Set(); err != nil {
			return fmt.Errorf("setting arp ignore: %w", err)
		}
	}

	// configure firewall
	// update only the interface name, in case there is a different VPN technology used
	newCfg := netw.fwConfig.CopyWith(
		firewall.WithTunnelInterface(netw.vpnet.Tun().Interface().Name),
	)
	if err := netw.configureFirewall(newCfg); err != nil {
		return fmt.Errorf("configuring firewall: %w", err)
	}

	netw.lastServer = serverData
	netw.lastCreds = creds
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

		netw.interfaces = mapset.NewSet[string]()
	}
	return nil
}

func (netw *Combined) stop() error {
	if netw.vpnet == nil {
		return errNilVPN
	}
	netw.publisher.Publish("stopping network configuration")

	netw.unblockIPv6()

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
	netw.isVpnSet = false
	if !netw.isKillSwitchSet {
		if err = netw.unsetNetwork(); err != nil {
			return fmt.Errorf("unsetting network: %w", err)
		}
	}

	if err := netw.arpIgnoreSetter.Unset(); err != nil {
		return fmt.Errorf("unsetting arp ignore: %w", err)
	}

	// configure firewall
	// remove tunnel interface and use the KS state, in case is set internally, e.g. at reconnect
	newCfg := netw.fwConfig.CopyWith(
		firewall.WithTunnelInterface(""),
		firewall.WithKillSwitch(netw.isKillSwitchSet),
	)
	if err := netw.configureFirewall(newCfg); err != nil {
		return fmt.Errorf("configuring firewall at stop: %w", err)
	}

	netw.switchToNextVpn()

	return nil
}

// switchToNextVpn check if VPN technology was changed when connect was in progress
func (netw *Combined) switchToNextVpn() {
	if netw.nextVPN != nil {
		netw.vpnet = netw.nextVPN
		netw.nextVPN = nil
	}
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

// Will configure the firewall based on the config parameter.
// If config doesn't have VPN, KS, or meshnet enabled it will delete the firewall rules from the system
func (netw *Combined) configureFirewall(config firewall.Config) error {
	if config.IsEmpty() {
		if err := netw.fw.Flush(); err != nil {
			return fmt.Errorf("removing all the rules from firewall: %w", err)
		}
	} else if err := netw.fw.Configure(config); err != nil {
		return fmt.Errorf("configuring firewall: %w", err)
	}
	netw.fwConfig = config

	return nil
}

func (netw *Combined) resetAllowlist() error {
	// this is done in order to maintain the order of the firewall rules
	log.Println(internal.InfoPrefix, "reset allow list")
	if err := netw.allowlistRouter.Flush(); err != nil {
		return fmt.Errorf("flushing allowlist router: %w", err)
	}

	if err := netw.setAllowlist(netw.allowlist); err != nil {
		return fmt.Errorf("re-setting allowlist: %w", err)
	}
	return nil
}

// EnableFirewall activates the firewall and applies the rules
// according to the user's settings. (killswitch, allowlist, meshnet)
func (netw *Combined) EnableFirewall() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	if err := netw.fw.Enable(); err != nil {
		return fmt.Errorf("enabling firewall: %w", err)
	}

	// use the existing information to configure the firewall
	if err := netw.configureFirewall(netw.fwConfig); err != nil {
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

func (netw *Combined) blockIPv6() {
	// Always disable IPv6 with sysctl in the system
	err := netw.ipv6Blocker.Block()
	if err != nil {
		log.Println(internal.WarningPrefix, "Failed to block ipv6 using sysctl", err)
	}
}

func (netw *Combined) unblockIPv6() {
	// Unblock from sysctl
	err := netw.ipv6Blocker.Unblock()
	if err != nil {
		log.Println(internal.WarningPrefix, "Failed to unblock ipv6 using sysctl", err)
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
	} else {
		netw.allowlist = allowlist
		if netw.lanDiscovery {
			netw.allowlist = addLANPermissions(allowlist)
		}
	}

	// configure firewall
	// use netw.allowlist because it will be updated with LAN ranges if needed
	newCfg := netw.fwConfig.CopyWith(
		firewall.WithAllowlist(netw.allowlist),
	)

	if err := netw.configureFirewall(newCfg); err != nil {
		return fmt.Errorf("firewall at allowlist: %w", err)
	}

	return nil
}

func (netw *Combined) setAllowlist(allowlist config.Allowlist) error {
	// allow traffic to LAN - only when user enabled lan-discovery
	if netw.lanDiscovery {
		allowlist = addLANPermissions(allowlist)
	}

	// adjust allow subnet routing rules
	if err := netw.policyRouter.SetupRoutingRules(
		netw.enableLocalTraffic,
		netw.lanDiscovery,
		allowlist.Subnets,
	); err != nil {
		return fmt.Errorf(
			"setting routing rules: %w",
			err,
		)
	}

	netw.allowlist = allowlist
	return nil
}

func (netw *Combined) unsetAllowlist() error {
	log.Println(internal.InfoPrefix, "unset allow list")
	if err := netw.allowlistRouter.Flush(); err != nil {
		log.Println(internal.WarningPrefix, "flushing allowlist router:", err)
	}
	return nil
}

func (netw *Combined) IsNetworkSet() bool {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.isNetworkSet
}

func (netw *Combined) setNetwork(allowlist config.Allowlist) error {
	if err := netw.setAllowlist(allowlist); err != nil {
		return err
	}

	netw.isNetworkSet = true
	return nil
}

func (netw *Combined) UnsetFirewall() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	// just refresh the firewall because the netw.fwConfig already contains the correct values
	if err := netw.configureFirewall(netw.fwConfig); err != nil {
		return fmt.Errorf("unset firewall: %w", err)
	}
	return nil
}

func (netw *Combined) unsetNetwork() error {
	if err := netw.unsetAllowlist(); err != nil {
		return err
	}

	netw.isNetworkSet = false
	return nil
}

func (netw *Combined) SetKillSwitch() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.setKillSwitch()
}

func (netw *Combined) setKillSwitch() error {
	// configure firewall
	// use netw.allowlist to ensure it has the correct value
	// because LAN ranges are added later for LAN discovery
	newCfg := netw.fwConfig.CopyWith(
		firewall.WithKillSwitch(true),
		firewall.WithAllowlist(netw.allowlist),
	)

	if err := netw.configureFirewall(newCfg); err != nil {
		return fmt.Errorf("enabling kill switch: %w", err)
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

	// configure firewall
	newCfg := netw.fwConfig.CopyWith(
		firewall.WithKillSwitch(false),
	)

	if err := netw.configureFirewall(newCfg); err != nil {
		return fmt.Errorf("disabling kill switch: %w", err)
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
	if netw.isMeshnetSet {
		if err := netw.refresh(c); err != nil {
			return fmt.Errorf("refresh meshnet: %w", err)
		}

		if err := netw.ipForwardSetter.Set(); err != nil {
			return fmt.Errorf("IP forwarding enabling: %w", err)
		}

		// configure firewall
		newCfg := netw.fwConfig.CopyWith(
			firewall.WithMeshnetInfo(firewall.NewMeshInfo(netw.cfg, netw.mesh.Tun().Interface().Name)),
		)
		if err := netw.configureFirewall(newCfg); err != nil {
			return fmt.Errorf("configure firewall for meshnet refresh: %w", err)
		}
	}

	return meshnet.ErrMeshnetNotEnabled
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

			if err := netw.ipForwardSetter.Unset(); err != nil {
				log.Println(internal.DeferPrefix, err)
			}

			cfg := netw.fwConfig.CopyWith(
				firewall.WithMeshnetInfo(nil),
			)
			if err := netw.configureFirewall(cfg); err != nil {
				log.Println(internal.DeferPrefix, err)
			}

			if err := netw.dnsHostSetter.UnsetHosts(); err != nil {
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
		netw.enableLocalTraffic,
		netw.lanDiscovery,
		netw.allowlist.Subnets,
	); err != nil {
		return fmt.Errorf("setting routing rules: %w", err)
	}
	routingRulesSet = true

	// add routes for new peers and remove for the old ones
	netw.publisher.Publish("adding mesh route")
	if err := netw.peerRouter.Add(routes.Route{
		Subnet:  internal.MeshSubnet,
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

	if err := netw.ipForwardSetter.Set(); err != nil {
		return fmt.Errorf("IP forwarding setting: %w", err)
	}

	// configure firewall
	newCfg := netw.fwConfig.CopyWith(
		firewall.WithMeshnetInfo(firewall.NewMeshInfo(netw.cfg, netw.mesh.Tun().Interface().Name)),
	)
	if err := netw.configureFirewall(newCfg); err != nil {
		return fmt.Errorf("configure firewall for set meshnet: %w", err)
	}

	netw.isMeshnetSet = true
	netw.lastPrivateKey = privateKey

	return nil
}

func (netw *Combined) refresh(cfg mesh.MachineMap) error {
	if err := netw.dnsHostSetter.UnsetHosts(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	if err := netw.mesh.Refresh(cfg); err != nil {
		return fmt.Errorf("refreshing mesh: %w", err)
	}
	netw.cfg = cfg

	// TODO (LVPN-4031): detect which peer we are connected (if connected)
	// to and check if maybe allowLocalAccess permission has changed and
	// if so, change routing to route to local LAN

	var hostName string
	var domainNames []string

	//nolint:staticcheck
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

	//nolint:staticcheck
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

	netw.publisher.Publish("refreshing mesh")
	return nil
}

func (netw *Combined) UnSetMesh() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	// clear this flag only when user turns mesh off, cannot do that in internal func
	// because it is used during refresh, and during refresh we need to know what
	// was set before i.e. during mesh peer connect
	netw.enableLocalTraffic = true
	if err := netw.unSetMesh(); err != nil {
		return fmt.Errorf("unsetting meshnet: %w", err)
	}

	if err := netw.ipForwardSetter.Unset(); err != nil {
		return fmt.Errorf("IP forwarding unset: %w", err)
	}

	return nil
}

func (netw *Combined) unSetMesh() error {
	if !netw.isMeshnetSet {
		return ErrMeshNotActive
	}
	if err := netw.dnsHostSetter.UnsetHosts(); err != nil {
		return fmt.Errorf("unsetting hosts: %w", err)
	}

	newCfg := netw.fwConfig.CopyWith(
		firewall.WithMeshnetInfo(nil),
	)
	if err := netw.configureFirewall(newCfg); err != nil {
		return fmt.Errorf("configure firewall for set meshnet: %w", err)
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

func (netw *Combined) PermitFileshare() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.configureFileshareAccess(false)
}

func (netw *Combined) ForbidFileshare() error {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	return netw.configureFileshareAccess(true)
}

func (netw *Combined) configureFileshareAccess(block bool) error {
	if !netw.isMeshnetSet {
		return nil
	}

	cfg := netw.fwConfig.CopyWith(
		firewall.WithBlockFileshare(block),
	)
	return netw.configureFirewall(cfg)
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

func (netw *Combined) SetLanDiscovery(enabled bool) {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	netw.lanDiscovery = enabled

	// if routing rules were set - they will be adjusted as needed
	if netw.isMeshnetSet || netw.isVpnSet {
		if err := netw.policyRouter.SetupRoutingRules(
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
}

func (netw *Combined) GetConnectionParameters() (vpn.ServerData, bool) {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	return netw.vpnet.GetConnectionParameters()
}

// SetARPIgnore sets arp ignore to the desired value if VPN connection is active. Networker will set arp ignore
// accordingly upon subsequent connections. Setting arp ignore to value previously configured is a noop.
func (netw *Combined) SetARPIgnore(ignoreARP bool) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	if !netw.isConnectedToVPN() {
		netw.ignoreARP = ignoreARP
		return nil
	}

	if ignoreARP {
		if err := netw.arpIgnoreSetter.Set(); err != nil {
			return fmt.Errorf("setting arp ignore: %w", err)
		}
	} else {
		if err := netw.arpIgnoreSetter.Unset(); err != nil {
			return fmt.Errorf("unsetting arp ignore: %w", err)
		}
	}

	netw.ignoreARP = ignoreARP

	return nil
}
