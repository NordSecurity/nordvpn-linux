package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedRulesAdd(t *testing.T) {
	tests := []struct {
		testName string
		rule     Rule
		given    OrderedRules
		expected OrderedRules
		hasError bool
	}{
		{
			testName: "nameless rule",
			rule:     Rule{},
			given:    OrderedRules{},
			expected: OrderedRules{},
			hasError: true,
		},
		{
			testName: "rule with a name",
			rule:     Rule{Name: "block"},
			given:    OrderedRules{},
			expected: OrderedRules{
				rules: []Rule{
					{Name: "block"},
				},
			},
			hasError: false,
		},
		{
			testName: "duplicate rule",
			rule:     Rule{Name: "block"},
			given: OrderedRules{
				rules: []Rule{
					{Name: "block"},
				},
			},
			expected: OrderedRules{
				rules: []Rule{
					{Name: "block"},
				},
			},
			hasError: true,
		},
		{
			testName: "replace existing rule",
			rule:     Rule{Name: "block", Interfaces: []net.Interface{{Name: "lo0"}}},
			given: OrderedRules{
				rules: []Rule{
					{Name: "block", Interfaces: []net.Interface{{Name: "en0"}}},
				},
			},
			expected: OrderedRules{
				rules: []Rule{
					{Name: "block", Interfaces: []net.Interface{{Name: "lo0"}}},
				},
			},
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := test.given.Add(test.rule)
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*Error)
				assert.True(t, ok)
			}
			assert.ElementsMatch(t, test.expected.rules, test.given.rules)
		})
	}
}

func TestOrderedRulesGet(t *testing.T) {
	tests := []struct {
		testName string
		given    OrderedRules
		ruleName string
		expected Rule
		hasError bool
	}{
		{
			testName: "empty slice",
			given:    OrderedRules{rules: []Rule{}},
			ruleName: "block",
			expected: Rule{},
			hasError: true,
		},
		{
			testName: "nil slice",
			given:    OrderedRules{},
			ruleName: "block",
			expected: Rule{},
			hasError: true,
		},
		{
			testName: "existing rule",
			given: OrderedRules{
				rules: []Rule{
					{Name: "block"},
				},
			},
			ruleName: "block",
			expected: Rule{Name: "block"},
			hasError: false,
		},
		{
			testName: "missing rule",
			given: OrderedRules{
				rules: []Rule{
					{Name: "block"},
				},
			},
			ruleName: "allow",
			expected: Rule{},
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			rule, err := test.given.Get(test.ruleName)
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*Error)
				assert.True(t, ok)
			}
			assert.Equal(t, test.expected, rule)
		})
	}
}

func TestOrderedRulesDelete(t *testing.T) {
	tests := []struct {
		testName string
		given    OrderedRules
		ruleName string
		expected OrderedRules
		hasError bool
	}{
		{
			testName: "empty slice",
			given:    OrderedRules{rules: []Rule{}},
			ruleName: "block",
			expected: OrderedRules{rules: []Rule{}},
			hasError: true,
		},
		{
			testName: "nil slice",
			given:    OrderedRules{rules: nil},
			ruleName: "block",
			expected: OrderedRules{rules: []Rule{}},
			hasError: true,
		},
		{
			testName: "single element found",
			given: OrderedRules{
				rules: []Rule{
					{Name: "block"},
				},
			},
			ruleName: "block",
			expected: OrderedRules{rules: []Rule{}},
			hasError: false,
		},
		{
			testName: "single element not found",
			given: OrderedRules{
				rules: []Rule{
					{Name: "block"},
				},
			},
			ruleName: "allow",
			expected: OrderedRules{
				rules: []Rule{
					{Name: "block"},
				},
			},
			hasError: true,
		},
		{
			testName: "resize on delete",
			given: OrderedRules{
				rules: []Rule{
					{Name: "block"},
					{Name: "allow"},
					{Name: "filter"},
				},
			},
			ruleName: "allow",
			expected: OrderedRules{
				rules: []Rule{
					{Name: "block"},
					{Name: "filter"},
				},
			},
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := test.given.Delete(test.ruleName)
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*Error)
				assert.True(t, ok)
			}
			assert.ElementsMatch(t, test.expected.rules, test.given.rules)
		})
	}
}
