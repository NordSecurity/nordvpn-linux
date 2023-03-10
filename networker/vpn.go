package networker

import (
	"fmt"
	"net/netip"
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

// refreshVPN fully re-creates the VPN tunnel but keeps the firewall
// rules
// This is needed since libtelio's NotifyNetworkChange does not well
// therefore, full tunnel must be re-created
//
// Thread unsafe.
func (netw *Combined) refreshVPN() error {
	meshnetSet := netw.isMeshnetSet
	started := netw.isVpnSet
	killswitch := netw.isKillSwitchSet
	var ip netip.Addr

	if started {
		if !killswitch {
			if err := netw.setKillSwitch(netw.whitelist); err != nil {
				return fmt.Errorf("setting killswitch: %w", err)
			}
		}

		if netw.vpnet.Tun() != nil && len(netw.vpnet.Tun().IPs()) > 0 {
			ip = netw.vpnet.Tun().IPs()[0]
		}

		if err := netw.stop(); err != nil {
			return fmt.Errorf("stopping networker: %w", err)
		}
	}

	if meshnetSet {
		if netw.mesh.Tun() != nil && len(netw.mesh.Tun().IPs()) > 0 {
			ip = netw.mesh.Tun().IPs()[0]
		}

		if err := netw.unSetMesh(); err != nil {
			return fmt.Errorf("stopping meshnet: %w", err)
		}

		if err := netw.setMesh(netw.cfg, ip, netw.lastPrivateKey); err != nil {
			return fmt.Errorf("starting meshnet: %w", err)
		}
	}

	if started {
		if err := netw.start(
			netw.lastCreds,
			netw.lastServer,
			netw.whitelist,
			netw.lastNameservers,
		); err != nil {
			return fmt.Errorf("starting networker: %w", err)
		}
	}

	if !killswitch {
		if err := netw.unsetKillSwitch(); err != nil {
			return fmt.Errorf("unsetting killswitch: %w", err)
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
