package routes

import (
	"sync"
)

const (
	defaultCustomRoutingTableID uint = 205
)

func TableID() uint { return defaultCustomRoutingTableID }

// PolicyAgent is stateless and is responsible for creating and deleting policy
// based routes.
//
// Used by implementers.
type PolicyAgent interface {
	SetupRoutingRules(bool, bool, bool) error
	CleanupRouting() error
	TableID() uint
}

// Service is stateful and updates system routing configuration by using the
// appropriate agent.
//
// Used by callers.
type PolicyService interface {
	SetupRoutingRules(bool, bool, bool) error
	CleanupRouting() error
	// TableID of the routing table.
	TableID() uint
	// Enable sets up previously remembered rules.
	Enable() error
	// Disable remembers previously added rules before clearing them.
	Disable() error
	IsEnabled() bool
}

// PolicyRouter is responsible for changing one routing agent over another.
//
// Thread-safe.
type PolicyRouter struct {
	current     PolicyAgent
	noop        PolicyAgent
	working     PolicyAgent
	appliedRule *struct {
		ipv6         bool
		enableLocal  bool
		lanDiscovery bool
	}
	isEnabled bool
	mu        sync.Mutex
}

func NewPolicyRouter(noop, working PolicyAgent, enabled bool) *PolicyRouter {
	var current PolicyAgent
	if enabled {
		current = working
	} else {
		current = noop
	}
	return &PolicyRouter{
		current:   current,
		noop:      noop,
		working:   working,
		isEnabled: enabled,
	}
}

func (p *PolicyRouter) SetupRoutingRules(
	ipv6,
	enableLocal,
	lanDiscovery bool,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.current.SetupRoutingRules(ipv6, enableLocal, lanDiscovery); err != nil {
		return err
	}
	p.appliedRule = &struct {
		ipv6         bool
		enableLocal  bool
		lanDiscovery bool
	}{ipv6, enableLocal, lanDiscovery}
	return nil
}

func (p *PolicyRouter) CleanupRouting() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.current.CleanupRouting(); err != nil {
		return err
	}
	p.appliedRule = nil
	return nil
}

func (p *PolicyRouter) TableID() uint {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current.TableID()
}

func (p *PolicyRouter) Enable() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isEnabled {
		if p.appliedRule != nil {
			if err := p.working.SetupRoutingRules(p.appliedRule.ipv6, p.appliedRule.enableLocal, p.appliedRule.lanDiscovery); err != nil {
				return err
			}
		}
		p.current = p.working
		p.isEnabled = true
	}
	return nil
}

func (p *PolicyRouter) Disable() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isEnabled {
		if err := p.current.CleanupRouting(); err != nil {
			return err
		}
		p.current = p.noop
		p.isEnabled = false
	}
	return nil
}

func (p *PolicyRouter) IsEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isEnabled
}
