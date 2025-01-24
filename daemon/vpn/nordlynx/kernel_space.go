package nordlynx

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

type KernelSpace struct {
	state           vpn.State
	active          bool
	fwmark          uint32
	tun             *tunnel.Tunnel
	eventsPublisher *vpn.Events
	serverData      vpn.ServerData
	sync.Mutex
}

func NewKernelSpace(fwmark uint32, eventsPublisher *vpn.Events) *KernelSpace {
	return &KernelSpace{
		state:           vpn.ExitedState,
		fwmark:          fwmark,
		eventsPublisher: eventsPublisher,
	}
}

func (k *KernelSpace) Start(
	_ context.Context,
	creds vpn.Credentials,
	serverData vpn.ServerData,
) (err error) {
	k.Lock()
	defer k.Unlock()
	if k.active {
		return vpn.ErrVPNAIsAlreadyStarted
	}

	event := events.DataConnect{
		EventStatus:         events.StatusAttempt,
		TargetServerIP:      serverData.IP.String(),
		TargetServerCountry: serverData.Country,
		TargetServerCity:    serverData.City,
		IsVirtualLocation:   serverData.VirtualLocation,
	}

	k.eventsPublisher.Connected.Publish(event)
	defer func() {
		if err != nil {
			k.eventsPublisher.Disconnected.Publish(events.DataDisconnect{})
			return
		}
		event.EventStatus = events.StatusSuccess
		k.eventsPublisher.Connected.Publish(event)
	}()

	conf := wgQuickConfig(
		creds.NordLynxPrivateKey,
		k.fwmark,
		serverData.NordLynxPublicKey,
		serverData.IP,
	)

	k.serverData = serverData

	// check if wireguard is not up already
	if _, err := exec.Command("ip", "link", "show", "dev", InterfaceName).Output(); err == nil {
		return vpn.ErrTunnelAlreadyExists
	}

	// add wireguard interface
	if err := upWGInterface(InterfaceName); err != nil {
		return fmt.Errorf("turning on nordlynx: %w", err)
	}

	iface, err := net.InterfaceByName(InterfaceName)
	if err != nil {
		if err := k.Stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return err
	}

	interfaceIps := []netip.Addr{netip.MustParseAddr("10.5.0.2")}
	ipv6, err := vpn.InterfaceIPv6(serverData.IP, interfaceID())
	if err == nil {
		interfaceIps = append(interfaceIps, ipv6)
	}

	tun := tunnel.New(*iface, interfaceIps, netip.Prefix{})
	k.tun = tun
	if err := pushConfig(tun.Interface(), conf); err != nil {
		if err := k.stop(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
		return fmt.Errorf("setting nordlynx server to connect to: %w", err)
	}

	if err := tun.AddAddrs(); err != nil {
		if err := k.stop(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
		return err
	}

	if err := tun.Up(); err != nil {
		if err := k.stop(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
		return err
	}

	if err := SetMTU(tun.Interface()); err != nil {
		if err := k.stop(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
		return fmt.Errorf("setting MTU for nordlynx interface: %w", err)
	}

	k.active = true
	k.state = vpn.ConnectedState
	return nil
}

// Stop is used by disconnect command
func (k *KernelSpace) Stop() error {
	k.Lock()
	defer k.Unlock()
	k.eventsPublisher.Disconnected.Publish(events.DataDisconnect{})
	if k.state == vpn.ConnectingState {
		k.state = vpn.ExitingState
		return nil
	}
	return k.stop()
}

func (k *KernelSpace) IsActive() bool {
	k.Lock()
	defer k.Unlock()
	return k.active
}

func (k *KernelSpace) Tun() tunnel.T {
	k.Lock()
	defer k.Unlock()
	return k.tun
}

func (k *KernelSpace) State() vpn.State {
	k.Lock()
	defer k.Unlock()
	return k.state
}

func (k *KernelSpace) GetConnectionParameters() (vpn.ServerData, bool) {
	k.Lock()
	defer k.Unlock()
	return k.serverData, k.active
}

// stop is used on errors
func (k *KernelSpace) stop() error {
	if k.tun != nil {
		err := deleteInterface(k.tun.Interface())
		if err != nil {
			return err
		}
	}

	k.active = false
	k.tun = nil
	k.state = vpn.ExitedState
	return nil
}

func (k *KernelSpace) NetworkChanged() error {
	return fmt.Errorf("not supported")
}

func pushConfig(iface net.Interface, wgconf string) error {
	// fill temp file with generated config
	tmp, err := internal.FileTemp(iface.Name, []byte(wgconf))
	if err != nil {
		return err
	}
	defer internal.FileDelete(tmp.Name())

	// pass config file to interface
	debug("wg", "setconf", iface.Name, tmp.Name())
	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command("wg", "setconf", iface.Name, tmp.Name()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting wireguard config: %s: %w", string(out), err)
	}

	return nil
}

// wgQuickTemplate is a template for WG-Quick
const wgQuickTemplate = `[Interface]
PrivateKey = %s
Fwmark = %#x
[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/0,::/0
Endpoint = %s
PersistentKeepalive = 25`

func wgQuickConfig(
	privateKey string,
	fwmark uint32,
	publicKey string,
	serverIP netip.Addr,
) string {
	return fmt.Sprintf(
		wgQuickTemplate,
		privateKey,
		fwmark,
		publicKey,
		net.JoinHostPort(
			serverIP.String(),
			strconv.Itoa(defaultPort),
		),
	)
}
