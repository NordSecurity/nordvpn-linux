package firewall

import (
	"fmt"
	"net"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestFirewallService(t *testing.T) {
	assert.Implements(t, (*Service)(nil), &Firewall{})
}

// IsEnabled reports firewall status.
func (fw *Firewall) isEnabled() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.enabled
}

type mockAgent struct {
	added   int
	deleted int
}

type failingAgent struct {
	added   int
	deleted int
}

func (m *mockAgent) Add(rule Rule) error {
	m.added++
	return nil
}

func (m *mockAgent) Delete(rule Rule) error {
	m.deleted++
	return nil
}

func (m *mockAgent) Flush() error {
	m.deleted++
	return nil
}

func (f *failingAgent) Add(rule Rule) error {
	f.added++
	return fmt.Errorf("adding")
}

func (f *failingAgent) Delete(rule Rule) error {
	f.deleted++
	return fmt.Errorf("deleting")
}

func (m *failingAgent) Flush() error {
	return nil
}

func TestFirewallAdd(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		rules    []Rule
		expected []Rule
		agent    Agent
		hasError bool
	}{
		{
			name:     "empty slice",
			rules:    []Rule{},
			expected: []Rule{},
			agent:    &mockAgent{},
			hasError: false,
		},
		{
			name:     "nil slice",
			rules:    nil,
			expected: nil,
			agent:    &mockAgent{},
			hasError: false,
		},
		{
			name: "one rule added without error",
			rules: []Rule{
				{
					Name: "block",
				},
			},
			expected: []Rule{
				{
					Name: "block",
				},
			},
			agent:    &mockAgent{},
			hasError: false,
		},
		{
			name: "multiple rules added without error",
			rules: []Rule{
				{
					Name: "allow",
				},
				{
					Name: "block",
				},
				{
					Name: "permit",
				},
			},
			expected: []Rule{
				{
					Name: "allow",
				},
				{
					Name: "block",
				},
				{
					Name: "permit",
				},
			},
			agent:    &mockAgent{},
			hasError: false,
		},
		{
			name:     "memory failure",
			rules:    []Rule{{}},
			expected: []Rule{},
			agent:    &mockAgent{},
			hasError: true,
		},
		{
			name:     "agent failure",
			rules:    []Rule{{}},
			expected: []Rule{},
			agent:    &failingAgent{},
			hasError: true,
		},
		{
			name: "duplicate failure",
			rules: []Rule{
				{
					Name: "allow",
				},
				{
					Name: "allow",
				},
			},
			expected: []Rule{
				{
					Name: "allow",
				},
			},
			agent:    &mockAgent{},
			hasError: true,
		},
		{
			name: "replace existing rule",
			rules: []Rule{
				{
					Name:       "block",
					Interfaces: []net.Interface{{Name: "lo0"}},
				},
				{
					Name:       "block",
					Interfaces: []net.Interface{{Name: "en0"}},
				},
			},
			expected: []Rule{
				{
					Name:       "block",
					Interfaces: []net.Interface{{Name: "en0"}},
				},
			},
			agent:    &mockAgent{},
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw := NewFirewall(test.agent, test.agent, &subs.Subject[string]{}, true)
			err := fw.Add(test.rules)
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*Error)
				assert.True(t, ok)
			}
			assert.ElementsMatch(t, test.expected, fw.rules.rules)
		})
	}
}

func TestFirewallDelete(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		rules    []Rule
		expected []Rule
		del      []string
		agent    Agent
		hasError bool
	}{
		{
			name:     "empty slice",
			rules:    []Rule{},
			expected: []Rule{},
			del:      []string{"block"},
			agent:    &mockAgent{},
			hasError: true,
		},
		{
			name:     "nil slice",
			rules:    nil,
			expected: []Rule{},
			del:      []string{"block"},
			agent:    &mockAgent{},
			hasError: true,
		},
		{
			name: "one rule deleted without error",
			rules: []Rule{
				{
					Name: "block",
				},
			},
			expected: []Rule{},
			del:      []string{"block"},
			agent:    &mockAgent{},
			hasError: false,
		},
		{
			name: "multiple rules deleted without error",
			rules: []Rule{
				{
					Name: "block",
				},
				{
					Name: "allow",
				},
			},
			expected: []Rule{},
			del:      []string{"block", "allow"},
			agent:    &mockAgent{},
			hasError: false,
		},
		{
			name:     "agent failure",
			rules:    []Rule{},
			expected: []Rule{},
			del:      []string{"block"},
			agent:    &failingAgent{},
			hasError: true,
		},
		{
			name:     "memory failure",
			rules:    []Rule{},
			expected: []Rule{},
			del:      []string{"block"},
			agent:    &mockAgent{},
			hasError: true,
		},
		{
			name: "no such rule",
			rules: []Rule{
				{
					Name: "block",
				},
			},
			expected: []Rule{
				{
					Name: "block",
				},
			},
			del:      []string{"allow"},
			agent:    &mockAgent{},
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw := NewFirewall(test.agent, test.agent, &subs.Subject[string]{}, false)
			fw.rules.rules = test.rules
			err := fw.Delete(test.del)
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*Error)
				assert.True(t, ok)
			}
			assert.ElementsMatch(t, test.expected, fw.rules.rules)
		})
	}
}

func TestFirewallEnable(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		enabled  bool
		expected bool
		hasError bool
	}{
		{
			name:     "enabled",
			enabled:  true,
			expected: true,
			hasError: true,
		},
		{
			name:     "disabled",
			enabled:  false,
			expected: true,
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw := NewFirewall(&mockAgent{}, &mockAgent{}, &subs.Subject[string]{}, test.enabled)
			err := fw.Enable()
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*Error)
				assert.True(t, ok)
			}
			assert.Equal(t, test.expected, fw.isEnabled())
		})
	}
}

func TestFirewallDisable(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		enabled  bool
		expected bool
		hasError bool
	}{
		{
			name:     "enabled",
			enabled:  true,
			expected: false,
			hasError: false,
		},
		{
			name:     "disabled",
			enabled:  false,
			expected: false,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw := NewFirewall(&mockAgent{}, &mockAgent{}, &subs.Subject[string]{}, test.enabled)
			err := fw.Disable()
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*Error)
				assert.True(t, ok)
			}
			assert.Equal(t, test.expected, fw.isEnabled())
		})
	}
}

func TestFirewallSwap(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name  string
		rules []Rule
	}{
		{
			name:  "empty slice",
			rules: []Rule{},
		},
		{
			name:  "nil slice",
			rules: nil,
		},
		{
			name: "one rule",
			rules: []Rule{
				{
					Name: "one",
				},
			},
		},
		{
			name: "many rules",
			rules: []Rule{
				{
					Name: "one",
				},
				{
					Name: "two",
				},
				{
					Name: "three",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent := &mockAgent{}
			fw := NewFirewall(agent, agent, &subs.Subject[string]{}, true)
			fw.rules.rules = test.rules
			err := fw.swap(agent, agent)
			assert.NoError(t, err)
			assert.Equal(t, agent.added, agent.deleted)
			assert.Equal(t, agent.added, len(test.rules))
		})
	}
}
