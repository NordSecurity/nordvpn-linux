// Package routes provides route setting functionality.
package routes

import (
	"fmt"
	"net"
	"net/netip"
	"sync"
)

var (
	// ErrRouteToOtherDestinationExists defines that route for specified network already exists but not to a specified destination
	ErrRouteToOtherDestinationExists = fmt.Errorf("route to differ")
)

// Route defines a route to Subnet through the specified Gateway
type Route struct {
	Gateway netip.Addr
	Subnet  netip.Prefix
	Device  net.Interface
	TableID uint
}

// IsEqual compares to routes for equality.
func (r *Route) IsEqual(to Route) bool {
	return r.Gateway == to.Gateway &&
		r.Subnet == to.Subnet &&
		r.Device.Name == to.Device.Name &&
		r.TableID == to.TableID
}

// Agent is stateless and is responsible for creating and deleting source based
// routes.
//
// Used by implementers.
type Agent interface {
	// Add route to a router
	Add(route Route) error
	// Flush all existing routes for this router
	Flush() error
}

// Service is stateful and updates system routing configuration by using the
// appropriate agent.
//
// Used by callers.
type Service interface {
	// Add route to a router
	Add(route Route) error
	// Flush all existing routes for this router
	Flush() error
	// Enable adds previously remembered routes.
	Enable(tableID uint) error
	// Disable remembers previously added routes before flushing them.
	Disable() error
	// IsEnabled reports route setting status
	IsEnabled() bool
}

// Router is responsible for changing one routing agent over another.
//
// Thread-safe.
type Router struct {
	current   Agent
	noop      Agent
	working   Agent
	applied   []Route
	isEnabled bool
	mu        sync.Mutex
}

func NewRouter(noop, working Agent, enabled bool) *Router {
	var current Agent
	if enabled {
		current = working
	} else {
		current = noop
	}

	return &Router{
		current:   current,
		noop:      noop,
		working:   working,
		isEnabled: enabled,
	}
}

func (r *Router) Add(route Route) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.current.Add(route); err != nil {
		return err
	}
	r.applied = append(r.applied, route) // append to nil slice allocates a new slice
	return nil
}

func (r *Router) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.current.Flush(); err != nil {
		return err
	}
	r.applied = nil
	return nil
}

func (r *Router) Enable(tableID uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.isEnabled {
		for _, route := range r.applied { // noop if r.applied is nil
			if route.TableID != 0 {
				route.TableID = tableID
			}
			if err := r.working.Add(route); err != nil {
				return err
			}
		}
		r.isEnabled = true
		r.current = r.working
	}
	return nil
}

func (r *Router) Disable() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isEnabled {
		if err := r.current.Flush(); err != nil {
			return err
		}
		r.isEnabled = false
		r.current = r.noop
	}
	return nil
}

func (r *Router) IsEnabled() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isEnabled
}
