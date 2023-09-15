// Package exitnode provides meshnet-related firewall management functionality.
package exitnode

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/kernel"
)

const (
	ipv4fwdKernelParamName = "net.ipv4.ip_forward"
)

// Node is exit node server side interface
type Node interface {
	Enable() error
	ResetPeers(mesh.MachinePeers, bool) error
	Disable() error
}

// Server struct for server side
type Server struct {
	mu             sync.Mutex
	interfaceNames []string // need to remember on which interface we started
	runCommandFunc runCommandFunc
	sysctlSetter   *kernel.SysctlSetter
}

// NewServer create & initialize new Server
func NewServer(interfaceNames []string, commandFunc runCommandFunc) *Server {
	// To allow peer routing when connected to OpenVPN as well
	interfaceNames = append(interfaceNames, openvpn.InterfaceName)
	return &Server{
		interfaceNames: interfaceNames,
		runCommandFunc: commandFunc,
		sysctlSetter: kernel.NewSysctlSetter(
			ipv4fwdKernelParamName,
			1,
			0,
		),
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

	err = enableMasquerading(en.interfaceNames, en.runCommandFunc)
	if err != nil {
		return fmt.Errorf("enabling masquerading: %w", err)
	}

	return nil
}

// EnablePeer enables masquerading for peer
func (en *Server) ResetPeers(peers mesh.MachinePeers, lanDiscovery bool) error {
	en.mu.Lock()
	defer en.mu.Unlock()

	trafficPeers := make([]TrafficPeer, 0, len(peers))
	for _, peer := range peers {
		if peer.Address.IsValid() {
			trafficPeers = append(trafficPeers, TrafficPeer{
				netip.PrefixFrom(peer.Address, peer.Address.BitLen()),
				peer.DoIAllowRouting,
				peer.DoIAllowLocalNetwork && lanDiscovery,
			})
		}
	}
	return resetPeersTraffic(trafficPeers, en.runCommandFunc)
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

	err = clearMasquerading(en.interfaceNames, en.runCommandFunc)
	if err != nil {
		return fmt.Errorf("clearing masquerading: %w", err)
	}

	if err := en.sysctlSetter.Unset(); err != nil {
		return fmt.Errorf(
			"unsetting the forwarding value: %w",
			err,
		)
	}

	return nil
}
