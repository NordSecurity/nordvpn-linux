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
	ResetPeers(mesh.MachinePeers, bool) error
	ResetFirewall(lanAvailable bool) error
	Disable() error
	SetAllowlist(config config.Allowlist, lanAvailable bool) error
}

// Server struct for server side
type Server struct {
	mu               sync.Mutex
	runCommandFunc   runCommandFunc
	sysctlSetter     kernel.SysctlSetter
	peers            mesh.MachinePeers
	allowlistManager allowlistManager
	enabled          bool
}

// NewServer create & initialize new Server
func NewServer(commandFunc runCommandFunc, allowlist config.Allowlist, sysctlSetter kernel.SysctlSetter) *Server {
	return &Server{
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
func (en *Server) ResetFirewall(lanAvailable bool) error {
	if !en.enabled {
		return nil
	}
	en.mu.Lock()
	defer en.mu.Unlock()

	return en.resetPeers(lanAvailable)
}

// EnablePeer enables masquerading for peer
func (en *Server) ResetPeers(peers mesh.MachinePeers, lanAvailable bool) error {
	en.mu.Lock()
	defer en.mu.Unlock()

	en.peers = peers
	return en.resetPeers(lanAvailable)
}

func (en *Server) resetPeers(lanAvailable bool) error {
	trafficPeers := make([]TrafficPeer, 0, len(en.peers))
	for _, peer := range en.peers {
		if peer.Address.IsValid() {
			trafficPeers = append(trafficPeers, TrafficPeer{
				netip.PrefixFrom(peer.Address, peer.Address.BitLen()),
				peer.DoIAllowRouting,
				peer.DoIAllowLocalNetwork && lanAvailable,
			})
		}
	}

	if err := resetPeersTraffic(trafficPeers, en.runCommandFunc); err != nil {
		return err
	}

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
