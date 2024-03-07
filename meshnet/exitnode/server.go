// Package exitnode provides meshnet-related firewall management functionality.
package exitnode

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

// Node is exit node server side interface
type Node interface {
	Enable() error
	ResetPeers(mesh.MachinePeers, bool, bool) error
	ResetFirewall(lanAvailable bool, killswitch bool) error
	Disable() error
	SetAllowlist(config config.Allowlist, lanAvailable bool) error
}

// Server struct for server side
type Server struct {
	mu               sync.Mutex
	interfaceNames   []string // need to remember on which interface we started
	runCommandFunc   runCommandFunc
	sysctlSetter     kernel.SysctlSetter
	peers            mesh.MachinePeers
	allowlistManager allowlistManager
	enabled          bool
}

// NewServer create & initialize new Server
func NewServer(interfaceNames []string, commandFunc runCommandFunc, allowlist config.Allowlist, sysctlSetter kernel.SysctlSetter) *Server {
	return &Server{
		interfaceNames:   interfaceNames,
		runCommandFunc:   commandFunc,
		sysctlSetter:     sysctlSetter,
		allowlistManager: newAllowlist(commandFunc, allowlist),
	}
}

// Enable backup current state and enable fwd+msq
func (en *Server) Enable() error {
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

// ResetFirewall resets peer rules when peers don't change
func (en *Server) ResetFirewall(lanAvailable bool, killswitch bool) error {
	if !en.enabled {
		return nil
	}
	en.mu.Lock()
	defer en.mu.Unlock()

	return en.resetPeers(lanAvailable, killswitch)
}

// EnablePeer enables masquerading for peer
func (en *Server) ResetPeers(peers mesh.MachinePeers, lanAvailable bool, killswitch bool) error {
	en.mu.Lock()
	defer en.mu.Unlock()

	en.peers = peers
	return en.resetPeers(lanAvailable, killswitch)
}

func (en *Server) resetPeers(lanAvailable bool, killswitch bool) error {
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

	if err := resetPeersTraffic(trafficPeers, en.interfaceNames, en.runCommandFunc, killswitch); err != nil {
		return err
	}

	// TODO: Peer local access should not depend on host VPN allowlists settings
	if err := en.allowlistManager.disableAllowlist(); err != nil {
		return err
	}
	// If exit node doesn't have full access to its own LAN, we need to ensure access to
	// allowlisted destinations
	if !lanAvailable {
		en.allowlistManager.setPeers(en.peers)
		return en.allowlistManager.enableAllowlist()
	}

	return nil
}

// Disable restore current state and disable fwd+msq
func (en *Server) Disable() error {
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

func (en *Server) SetAllowlist(allowlist config.Allowlist, lanAvailable bool) error {
	en.mu.Lock()
	defer en.mu.Unlock()

	if err := en.allowlistManager.disableAllowlist(); err != nil {
		return err
	}

	en.allowlistManager.setAllowlist(allowlist)

	if en.enabled && !lanAvailable {
		return en.allowlistManager.enableAllowlist()
	}

	return nil
}
