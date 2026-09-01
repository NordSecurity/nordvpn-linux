package mock

import "github.com/NordSecurity/nordvpn-linux/daemon/routes"

type PolicyRouter struct {
	EnableLocalTraffic bool
}

func (r *PolicyRouter) SetupRoutingRules(enableLan bool, _ bool, _ []string) error {
	r.EnableLocalTraffic = enableLan
	return nil
}
func (*PolicyRouter) CleanupRouting() error { return nil }
func (*PolicyRouter) TableID() uint         { return 0 }
func (*PolicyRouter) Enable() error         { return nil }
func (*PolicyRouter) Disable() error        { return nil }
func (*PolicyRouter) IsEnabled() bool       { return true }

type Router struct{}

func (Router) Add(routes.Route) error { return nil }
func (Router) Flush() error           { return nil }
func (Router) Enable(uint) error      { return nil }
func (Router) Disable() error         { return nil }
func (Router) IsEnabled() bool        { return true }
