/*
Package firewall provides firewall service to the caller
*/
package firewall

import (
	"fmt"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/exp/slices"
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

	ruleNamesFromOS, err := fw.current.GetActiveRules()
	if err != nil {
		log.Printf("%v, unable to get already active rules: %v\n", internal.WarningPrefix, err)
	}
	trafficWasDropped := false

	for _, rule := range rules {
		fw.publisher.Publish(fmt.Sprintf("adding rule %s", rule.Name))
		if rule.Name == "" {
			return NewError(ErrRuleWithoutName)
		}

		// check if rule exists already
		existingRule, err := fw.rules.Get(rule.Name)
		if err == nil {
			// rule with the given name exists, check if the rules are equal => replace or return error
			if existingRule.Equal(rule) {
				return NewError(ErrRuleAlreadyExists)
			}
			fw.publisher.Publish(fmt.Sprintf("replacing existing rule %s", rule.Name))
		} else {
			existingRule = Rule{}
			// Check if rule exists in the OS, but not in the app
			existsInOS := slices.ContainsFunc(ruleNamesFromOS, func(ruleName string) bool {
				return ruleName == rule.Name || ruleName == rule.SimplifiedName
			})
			if existsInOS {
				// block traffic until rules are swapped
				if !trafficWasDropped {
					blockRule := Rule{Name: "drop-all", Direction: TwoWay, Allow: false}
					if err := fw.current.Add(blockRule); err != nil {
						log.Printf("%s failed to temporarily block traffic: %v\n", internal.ErrorPrefix, err)
					} else {
						trafficWasDropped = true
						defer func() {
							if err := fw.current.Delete(blockRule); err != nil {
								log.Printf("%s failed to unblock temporarily blocked traffic: %v\n", internal.ErrorPrefix, err)
							}
						}()
					}
				}
				// delete it from the firewall because later it will be inserted
				log.Printf("%s rule already exists in OS: %v\n", internal.WarningPrefix, rule.Name)
				if err := fw.current.Delete(rule); err != nil {
					log.Printf("%s failed to delete rule %s from OS %v\n", internal.ErrorPrefix, rule.Name, err)
				}
			}
		}

		if err := fw.current.Add(rule); err != nil {
			return NewError(fmt.Errorf("adding %s: %w", rule.Name, err))
		}

		if err := fw.rules.Add(rule); err != nil {
			return NewError(fmt.Errorf("adding %s to memory: %w", rule.Name, err))
		}

		if !existingRule.IsEmpty() {
			// remove older rule because the new one was added
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

func (fw *Firewall) Flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	return fw.current.Flush()
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
