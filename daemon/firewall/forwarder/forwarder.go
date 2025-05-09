// Package forwarder manages the FORWARD chain rules(meshnet and allowlist).
package forwarder

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/kernel"
)

const (
	Ipv4fwdKernelParamName = "net.ipv4.ip_forward"
)

// ForwardChainManager is responsible for managing rules in the FORWARD chain of iptables
type ForwardChainManager interface {
	Enable() error
	ResetPeers(peers mesh.MachinePeers,
		lanAvailable bool,
		killswitch bool,
		enableAllowlist bool,
		allowlist config.Allowlist) error
	ResetFirewall(lanAvailable bool, killswitch bool, enableAllowlist bool, allowlist config.Allowlist) error
	Disable() error
}

// Forwarder manages the FORWARD chain in iptables
type Forwarder struct {
	mu               sync.Mutex
	interfaceNames   []string // need to remember on which interface we started
	runCommandFunc   runCommandFunc
	sysctlSetter     kernel.SysctlSetter
	peers            mesh.MachinePeers
	allowlistManager allowlistManager
	enabled          bool
}

// NewForwarder create & initialize new Server
func NewForwarder(interfaceNames []string, commandFunc runCommandFunc, sysctlSetter kernel.SysctlSetter) *Forwarder {
	return &Forwarder{
		interfaceNames:   interfaceNames,
		runCommandFunc:   commandFunc,
		sysctlSetter:     sysctlSetter,
		allowlistManager: newAllowlist(commandFunc),
	}
}

// Enable adds meshnet related rules to the FORWARD chain.
func (en *Forwarder) Enable() error {
	en.mu.Lock()
	defer en.mu.Unlock()

	if err := en.sysctlSetter.Set(); err != nil {
		return fmt.Errorf("enabling ipv4 forwarding: %w", err)
	}

	// block traffic from unauthorized peers
	err := enableFiltering(en.runCommandFunc)
	if err != nil {
		return fmt.Errorf("enabling filtering: %w", err)
	}

	en.enabled = true
	return nil
}

// ResetFirewall resets forwarding rules using the stored peer list. If meshnet is not enabled, only allowlist related
// rules will be affected.
func (en *Forwarder) ResetFirewall(lanAvailable bool,
	killswitch bool,
	enableAllowlist bool,
	allowlist config.Allowlist) error {
	if !en.enabled {
		if err := en.allowlistManager.disableAllowlist(); err != nil {
			return fmt.Errorf("disabling peer allowlist: %w", err)
		}
		if err := resetAllowlistRules(en.runCommandFunc,
			en.interfaceNames,
			killswitch,
			enableAllowlist,
			allowlist.Subnets); err != nil {
			return fmt.Errorf("reseting allowlist rules: %w", err)
		}

		return nil
	}
	en.mu.Lock()
	defer en.mu.Unlock()

	if err := en.resetPeers(lanAvailable, killswitch, enableAllowlist, allowlist); err != nil {
		return fmt.Errorf("reseting peers: %w", err)
	}

	return nil
}

// ResetPeers resets forwarding rules to respect settings in the provided peer list.
func (en *Forwarder) ResetPeers(peers mesh.MachinePeers,
	lanAvailable bool,
	killswitch bool,
	enableAllowlist bool,
	allowlist config.Allowlist) error {
	en.mu.Lock()
	defer en.mu.Unlock()

	en.peers = peers
	return en.resetPeers(lanAvailable, killswitch, enableAllowlist, allowlist)
}

func (en *Forwarder) resetPeers(lanAvailable bool, killswitch bool, enableAllowlist bool, allowlist config.Allowlist) error {
	trafficPeers := make([]TrafficPeer, 0, len(en.peers))
	for _, peer := range en.peers {
		if peer.Address.IsValid() {
			trafficPeers = append(trafficPeers, TrafficPeer{
				netip.PrefixFrom(peer.Address, peer.Address.BitLen()),
				peer.DoIAllowRouting,
				// TODO: Remove '&& lanAvailable'
				// According to the user-facing documentation meshnet peer local access does not depend on
				// host VPN lan discovery or allowlists settings
				peer.DoIAllowLocalNetwork && lanAvailable,
			})
		}
	}

	if err := resetForwardTraffic(trafficPeers,
		en.interfaceNames,
		en.runCommandFunc,
		killswitch,
		enableAllowlist,
		allowlist.Subnets); err != nil {
		return err
	}

	// TODO: Peer local access should not depend on host VPN allowlists settings
	if err := en.allowlistManager.disableAllowlist(); err != nil {
		return err
	}

	en.allowlistManager.setAllowlist(allowlist)

	// If exit node doesn't have full access to its own LAN, we need to ensure access to
	// allowlisted destinations
	if !lanAvailable {
		en.allowlistManager.setPeers(en.peers)
		return en.allowlistManager.enableAllowlist()
	}

	return nil
}

// Disable removes meshnet related rules from the FORWARD and NAT chains.
func (en *Forwarder) Disable() error {
	en.mu.Lock()
	defer en.mu.Unlock()

	var err error
	err = clearFiltering(en.runCommandFunc)
	if err != nil {
		return fmt.Errorf("clearing filtering: %w", err)
	}

	err = clearMasquerading(en.runCommandFunc)
	if err != nil {
		return fmt.Errorf("clearing masquerading: %w", err)
	}

	if err := en.sysctlSetter.Unset(); err != nil {
		return fmt.Errorf(
			"unsetting the forwarding value: %w",
			err,
		)
	}

	if err := en.allowlistManager.disableAllowlist(); err != nil {
		return fmt.Errorf("disabling allowlist: %w", err)
	}

	en.enabled = false

	return nil
}
