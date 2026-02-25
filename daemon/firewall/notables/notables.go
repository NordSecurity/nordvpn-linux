// Package notables implements noop firewall agent.
package notables

import (
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
)

type Facade struct{}

func (*Facade) Configure(vpnInfo *firewall.VpnInfo, meshMap *mesh.MachineMap) error { return nil }
func (*Facade) Flush() error                                                        { return nil }
