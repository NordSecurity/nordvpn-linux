// Package notables implements noop firewall agent.
package notables

import "github.com/NordSecurity/nordvpn-linux/daemon/firewall"

type Facade struct{}

func (*Facade) Add(firewall.Rule) error    { return nil }
func (*Facade) Delete(firewall.Rule) error { return nil }
