//go:build !telio

package main

import (
	"errors"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/config"
	cesh "github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

func getVpnFactory(eventsDbPath string, fwmark uint32, enableNATTraversal, enableLana bool) daemon.FactoryFunc {
	return func(tech config.Technology) (vpn.VPN, error) {
		switch tech {
		case config.Technology_NORDLYNX:
			return nordlynx.NewKernelSpace(fwmark), nil
		case config.Technology_OPENVPN:
			return openvpn.New(fwmark), nil
		case config.Technology_UNKNOWN_TECHNOLOGY:
			fallthrough
		default:
			return nil, errors.New("no such technology")
		}
	}
}

// mockMesh is a mock implementation of meshnet. It is used when telio
// is not available and should be used only for development purposes
type mockMesh bool

func (mockMesh) Enable(netip.Addr, string) error { return nil }
func (mockMesh) Disable() error                  { return nil }
func (mockMesh) IsActive() bool                  { return false }
func (mockMesh) Refresh(cesh.MachineMap) error   { return nil }
func (mockMesh) Tun() tunnel.T                   { return &tunnel.Tunnel{} }
func (mockMesh) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}

func meshnetImplementation(fn daemon.FactoryFunc) (meshnet.Mesh, error) {
	return mockMesh(true), nil
}

type mockKeygen bool

func (mockKeygen) Private() string      { return "" }
func (mockKeygen) Public(string) string { return "" }

func keygenImplementation(daemon.FactoryFunc) (mockKeygen, error) { return mockKeygen(true), nil }
