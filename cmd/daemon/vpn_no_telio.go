//go:build !telio

package main

import (
	"fmt"
	"net/netip"

	cesh "github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

func getNordlynxVPN(envIsDev bool,
	eventsDbPath string,
	fwmark uint32,
	cfg vpn.LibConfigGetter,
	appVersion string,
	eventsPublisher *vpn.Events) (*nordlynx.KernelSpace, error) {
	return nordlynx.NewKernelSpace(fwmark, eventsPublisher), nil
}

// noopMesh is a noop implementation of meshnet. It is used when telio
// is not available and should be used only for development purposes
type noopMesh bool

func (noopMesh) Enable(netip.Addr, string) error { return nil }
func (noopMesh) Disable() error                  { return nil }
func (noopMesh) IsActive() bool                  { return false }
func (noopMesh) Refresh(cesh.MachineMap) error   { return nil }
func (noopMesh) Tun() tunnel.T                   { return &tunnel.Tunnel{} }
func (noopMesh) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}

func (noopMesh) NetworkChanged() error {
	return fmt.Errorf("not supported")
}

func meshnetImplementation(fn daemon.FactoryFunc) (meshnet.Mesh, error) {
	return noopMesh(true), nil
}

type noopKeygen bool

func (noopKeygen) Private() string      { return "" }
func (noopKeygen) Public(string) string { return "" }

func keygenImplementation(daemon.FactoryFunc) (noopKeygen, error) { return noopKeygen(true), nil }
