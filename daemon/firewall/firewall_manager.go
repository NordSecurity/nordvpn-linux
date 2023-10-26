package firewall

import "fmt"

type Args []string

type FirewallRule interface {
	ToArgs() []Args
	ToUndoArgs() []Args
}

type FirewallManager struct {
	rules           map[string]FirewallRule
	commandExecutor CommandExecutor
	enabled         bool
}

func NewFirewallManager(commandExecutor CommandExecutor, enabled bool) FirewallManager {
	return FirewallManager{
		rules:           make(map[string]FirewallRule),
		commandExecutor: commandExecutor,
		enabled:         enabled,
	}
}

func (f *FirewallManager) addRule(rule FirewallRule) error {
	for _, args := range rule.ToArgs() {
		if err := f.commandExecutor.ExecuteCommand(args...); err != nil {
			return fmt.Errorf("executing add rule command: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) removeRule(rule FirewallRule) error {
	for _, args := range rule.ToUndoArgs() {
		if err := f.commandExecutor.ExecuteCommand(args...); err != nil {
			return fmt.Errorf("executing remove rule command: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) AddRule(name string, rule FirewallRule) error {
	if _, ok := f.rules[name]; ok {
		return fmt.Errorf("rule %s already exists", name)
	}

	if f.enabled {
		if err := f.addRule(rule); err != nil {
			return fmt.Errorf("adding rule: %w", err)
		}
	}

	f.rules[name] = rule
	return nil
}

func (f *FirewallManager) RemoveRule(name string) error {
	rule, ok := f.rules[name]

	if !ok {
		return fmt.Errorf("rule %s does not exist", name)
	}

	if f.enabled {
		if err := f.removeRule(rule); err != nil {
			return fmt.Errorf("removing rule: %w", err)
		}
	}

	delete(f.rules, name)

	return nil
}

func (f *FirewallManager) Enable() error {
	if f.enabled {
		return fmt.Errorf("firewall is already enabled")
	}

	for _, rule := range f.rules {
		if err := f.addRule(rule); err != nil {
			return fmt.Errorf("enabling firewall: %w", err)
		}
	}

	f.enabled = true

	return nil
}

func (f *FirewallManager) Disable() error {
	if !f.enabled {
		return fmt.Errorf("firewall is already disabled")
	}

	for _, rule := range f.rules {
		if err := f.removeRule(rule); err != nil {
			return fmt.Errorf("disabling firewall: %w", err)
		}
	}

	f.enabled = false

	return nil
}
