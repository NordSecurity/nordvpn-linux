package networker

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	mapset "github.com/deckarep/golang-set/v2"
)

// IsVPNActive returns true when connection to VPN server is established.
// Otherwise false is returned.
//
// Thread safe.
func (netw *Combined) IsVPNActive() bool {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.isConnectedToVPN()
}

// IsMeshnetActive returns true when meshnet was activated.
// Otherwise false is returned.
//
// Thread safe.
func (netw *Combined) IsMeshnetActive() bool {
	netw.mu.Lock()
	defer netw.mu.Unlock()
	return netw.isMeshnetSet
}

func (netw *Combined) handleNetworkChanged() error {
	if netw.isMeshnetSet {
		log.Println(internal.InfoPrefix, "handle network changes for meshnet")
		if err := netw.mesh.NetworkChanged(); err != nil {
			return err
		}
	}

	if netw.isVpnSet {
		// for Nordlynx VPN + Meshnet NetworkChanged was already executed, so skip
		vpn, ok := netw.mesh.(vpn.VPN)
		if netw.isMeshnetSet && ok && vpn == netw.vpnet {
			log.Println(internal.InfoPrefix, "skip network changed for VPN, already executed for meshnet")
		} else {
			log.Println(internal.InfoPrefix, "handle network changes for VPN")

			if err := netw.vpnet.NetworkChanged(); err != nil {
				return err
			}
		}
		if err := netw.fixForLinuxMint20(); err != nil {
			return err
		}

		// at network changes, even if the same interfaces still exist in the system,
		// the routes might not be configured anymore, because the OS will delete them when interfaces are gone.
		// Reset the allow list to be sure that allowed routes still work.
		if err := netw.resetAllowlist(); err != nil {
			return err
		}
	}

	return nil
}

// during network changes in Linux Mint 20.03, systemd removes the tunnel interface from DNS resolver.
func (netw *Combined) fixForLinuxMint20() error {
	if err := netw.setDNS(netw.lastNameservers); err != nil {
		return err
	}
	// It needs to be set with delay to be sure systemd finishes its internal setup at network changes,
	// otherwise systemd will remove again the tunnel from DNS resolver.
	// In this way nordvpn will be the last changing the DNS resolvers list.
	time.Sleep(1 * time.Second)
	if err := netw.setDNS(netw.lastNameservers); err != nil {
		return err
	}
	return nil
}

// refreshVPN will handle network changes
// 1. try to let each VPN implementation to handle, if the system interfaces didn't change
// 2. fully re-creates the VPN tunnel but keeps the firewall rules
// Thread unsafe.
func (netw *Combined) refreshVPN(ctx context.Context) error {
	isVPNStarted := netw.isVpnSet
	isMeshStarted := netw.isMeshnetSet

	if !isVPNStarted && !isMeshStarted {
		return nil
	}

	tunnelName := ""
	if netw.vpnet != nil && netw.vpnet.Tun() != nil {
		tunnelName = netw.vpnet.Tun().Interface().Name
	}
	newInterfaces := netw.devices(mapset.NewSet(tunnelName))
	newInterfaceDetected := !newInterfaces.IsSubset(netw.interfaces)
	log.Println(internal.InfoPrefix,
		"refresh VPN, new interface detected[]:",
		newInterfaceDetected,
		netw.interfaces, "->", newInterfaces)

	if err := netw.handleNetworkChanged(); err == nil {
		return nil
	} else {
		log.Println(internal.ErrorPrefix, "failed to handle network changes, reinit the tunnel", err)
	}

	netw.interfaces = newInterfaces

	return netw.reconnect(ctx, func() error { return nil })
}

// XXX: probably something can be reused
func (netw *Combined) RecreateVPN(recreateFn func() error) error {
	netw.mu.Lock()
	defer netw.mu.Unlock()

	// XXX: checks etc.

	return netw.reconnect(context.Background(), recreateFn)
}

func (netw *Combined) reconnect(ctx context.Context, hook func() error) (err error) {
	// XXX: Almost same as refreshVPN, but call `hook` in the middle
	return
}

// Thread unsafe.
func (netw *Combined) isConnectedToVPN() bool {
	if netw.vpnet == nil || netw.vpnet.Tun() == nil {
		return false
	}
	return netw.vpnet.IsActive()
}
