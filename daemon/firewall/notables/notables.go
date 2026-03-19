// Package notables implements noop firewall agent.
package notables

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
)

// to be deleted
type Facade struct{}

func (*Facade) Configure(config firewall.Config) error { return nil }
func (*Facade) Flush() error                           { return nil }
