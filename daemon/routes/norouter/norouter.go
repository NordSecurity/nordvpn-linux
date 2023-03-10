// Package norouter implements noop router.
package norouter

import "github.com/NordSecurity/nordvpn-linux/daemon/routes"

type Facade struct{}

func (*Facade) Add(routes.Route) error { return nil }
func (*Facade) Flush() error           { return nil }
