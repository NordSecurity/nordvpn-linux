package networker

import (
	"errors"
	"fmt"
	"log"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
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
		log.Println("reconfigure meshnet")
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
			log.Println("reconfigure VPN")

			return netw.vpnet.NetworkChanged()
		}
	}

	return nil
}

// refreshVPN will handle network changes
// 1. try to let each VPN implementation to handle
// 2. as fallback, fully re-creates the VPN tunnel but keeps the firewall rules
// Thread unsafe.
func (netw *Combined) refreshVPN() (err error) {
	errNetChanged := netw.handleNetworkChanged()
	if errNetChanged == nil {
		return nil
	}

	log.Println(internal.ErrorPrefix, "failed to handle network changes, reinit the tunnel", errNetChanged)

	started := netw.isVpnSet
	var ip netip.Addr
	var vpnErr, meshErr error
	defer func() { err = errors.Join(vpnErr, meshErr) }()

	if started {
		if !netw.isKillSwitchSet {
			if err := netw.setKillSwitch(netw.allowlist); err != nil {
				return fmt.Errorf("setting killswitch: %w", err)
			}
			defer func() {
				if vpnErr != nil {
					// Keep iptables rules to not expose user after background connect failure
					netw.isKillSwitchSet = false
				} else {
					vpnErr = netw.unsetKillSwitch()
				}
			}()
		}

		if netw.vpnet.Tun() != nil && len(netw.vpnet.Tun().IPs()) > 0 {
			ip = netw.vpnet.Tun().IPs()[0]
		}

		if vpnErr = netw.stop(); vpnErr != nil {
			vpnErr = fmt.Errorf("stopping networker: %w", vpnErr)
			return
		}
	}

	if netw.isMeshnetSet {
		if netw.mesh.Tun() != nil && len(netw.mesh.Tun().IPs()) > 0 {
			ip = netw.mesh.Tun().IPs()[0]
		}

		// Don't return on mesh errors yet, still have to try to start VPN
		meshErr = netw.unSetMesh()
		if meshErr != nil {
			meshErr = fmt.Errorf("unsetting mesh: %w", meshErr)
		} else {
			meshErr = netw.setMesh(netw.cfg, ip, netw.lastPrivateKey)
			if meshErr != nil {
				meshErr = fmt.Errorf("setting mesh: %w", meshErr)
			}
		}
	}

	if started {
		if vpnErr = netw.start(
			netw.lastCreds,
			netw.lastServer,
			netw.allowlist,
			netw.lastNameservers,
		); vpnErr != nil {
			vpnErr = fmt.Errorf("starting networker: %w", vpnErr)
			return
		}
	}

	return nil
}

// Thread unsafe.
func (netw *Combined) isConnectedToVPN() bool {
	if netw.vpnet == nil || netw.vpnet.Tun() == nil {
		return false
	}
	return netw.vpnet.IsActive()
}
