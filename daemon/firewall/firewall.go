/*
Package firewall provides firewall service to the caller
*/
package firewall

import (
	"fmt"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/events"
)

// Firewall is responsible for correctly changing one firewall agent over another.
//
// Thread-safe.
type Firewall struct {
	rules     OrderedRules
	current   Agent
	noop      Agent
	working   Agent
	publisher events.Publisher[string]
	enabled   bool
	mu        sync.Mutex
}

// NewFirewall produces an instance of Firewall.
func NewFirewall(noop, working Agent, publisher events.Publisher[string], enabled bool) *Firewall {
	var current Agent

	if enabled {
		current = working
	} else {
		current = noop
	}

	return &Firewall{
		working:   working,
		noop:      noop,
		enabled:   enabled,
		current:   current,
		publisher: publisher,
	}
}

// Add rules to the firewall.
func (fw *Firewall) Add(rules []Rule) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, rule := range rules {
		fw.publisher.Publish(fmt.Sprintf("adding rule %s", rule.Name))
		if rule.Name == "" {
			return NewError(ErrRuleWithoutName)
		}

		existingRule, err := fw.rules.Get(rule.Name)
		if err == nil {
			// rule with the given name exists, check if the rules are equal
			if existingRule.Equal(rule) {
				return NewError(ErrRuleAlreadyExists)
			}
			fw.publisher.Publish(fmt.Sprintf("replacing existing rule %s", rule.Name))
		}

		if err := fw.current.Add(rule); err != nil {
			return NewError(fmt.Errorf("adding %s: %w", rule.Name, err))
		}

		if err := fw.rules.Add(rule); err != nil {
			return NewError(fmt.Errorf("adding %s to memory: %w", rule.Name, err))
		}

		if err == nil {
			// remove older rule
			if err := fw.current.Delete(existingRule); err != nil {
				return NewError(fmt.Errorf("removing replaced rule %s to memory: %w", rule.Name, err))
			}
		}
	}
	return nil
}

// Delete rules from firewall by their names.
func (fw *Firewall) Delete(names []string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, name := range names {
		fw.publisher.Publish(fmt.Sprintf("deleting rule %s", name))
		rule, err := fw.rules.Get(name)
		if err != nil {
			return NewError(fmt.Errorf("getting %s: %w", name, err))
		}
		err = fw.current.Delete(rule)
		if err != nil {
			return NewError(fmt.Errorf("deleting %s: %w", name, err))
		}
		err = fw.rules.Delete(rule.Name)
		if err != nil {
			return NewError(fmt.Errorf("deleting %s from memory: %s", name, err))
		}
	}
	return nil
}

// Enable restores firewall operations from no-ops.
func (fw *Firewall) Enable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.enabled {
		return NewError(ErrFirewallAlreadyEnabled)
	}
	fw.enabled = true
	fw.current = fw.working
	return fw.swap(fw.noop, fw.current)
}

// Disable turns all firewall operations into no-ops.
func (fw *Firewall) Disable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if !fw.enabled {
		return NewError(ErrFirewallAlreadyDisabled)
	}
	fw.enabled = false
	fw.current = fw.noop
	return fw.swap(fw.working, fw.current)
}

func (fw *Firewall) swap(current Agent, next Agent) error {
	for _, rule := range fw.rules.rules {
		if err := current.Delete(rule); err != nil {
			return NewError(fmt.Errorf("deleting rule %s: %w", rule.Name, err))
		}
		if err := next.Add(rule); err != nil {
			return NewError(fmt.Errorf("adding rule %s: %w", rule.Name, err))
		}
	}
	return nil
}
