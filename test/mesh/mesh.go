package mesh

import (
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/errors"
	testtunnel "github.com/NordSecurity/nordvpn-linux/test/tunnel"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// Working stub of a github.com/NordSecurity/nordvpn-linux/meshnet.Mesh interface.
type Working struct{}

func (Working) Enable(netip.Addr, string) error       { return nil }
func (Working) Disable() error                        { return nil }
func (Working) Refresh(mesh.MachineMap) error         { return nil }
func (Working) StatusMap() (map[string]string, error) { return map[string]string{}, nil }
func (Working) Tun() tunnel.T                         { return testtunnel.Working{} }

// Failing stub of a github.com/NordSecurity/nordvpn-linux/meshnet.Mesh interface.
type Failing struct{}

func (Failing) Enable(netip.Addr, string) error       { return errors.ErrOnPurpose }
func (Failing) Disable() error                        { return errors.ErrOnPurpose }
func (Failing) Refresh(mesh.MachineMap) error         { return errors.ErrOnPurpose }
func (Failing) StatusMap() (map[string]string, error) { return nil, errors.ErrOnPurpose }
func (Failing) Tun() tunnel.T                         { return testtunnel.Working{} }
