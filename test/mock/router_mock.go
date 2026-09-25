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

type ToggleRouter struct {
	Enabled bool
	Err     error
}

func (*ToggleRouter) Add(routes.Route) error { return nil }
func (*ToggleRouter) Flush() error           { return nil }

func (r *ToggleRouter) Enable(uint) error {
	if r.Err != nil {
		return r.Err
	}
	r.Enabled = true
	return nil
}

func (r *ToggleRouter) Disable() error {
	if r.Err != nil {
		return r.Err
	}
	r.Enabled = false
	return nil
}

func (r *ToggleRouter) IsEnabled() bool { return r.Enabled }

type TogglePolicyRouter struct {
	Enabled bool
}

func (*TogglePolicyRouter) SetupRoutingRules(bool, bool, []string) error { return nil }
func (*TogglePolicyRouter) CleanupRouting() error                        { return nil }
func (*TogglePolicyRouter) TableID() uint                                { return 0 }
func (r *TogglePolicyRouter) IsEnabled() bool                            { return r.Enabled }

func (r *TogglePolicyRouter) Enable() error {
	r.Enabled = true
	return nil
}

func (r *TogglePolicyRouter) Disable() error {
	r.Enabled = false
	return nil
}
