package iptablesmanager

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	iptablesmock "github.com/NordSecurity/nordvpn-linux/test/mock/firewall/iptables_manager"
	"github.com/stretchr/testify/assert"
)

func TestIptablesManager(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		rules           []string
		newRulePriority RulePriority
		expectedCommand string
	}{
		{
			name: "insert rule with lowest priority",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
			},
			newRulePriority: 0,
			expectedCommand: "-I INPUT 4 -j DROP -m comment --comment nordvpn-0",
		},
		{
			name: "insert rule with highest priority",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
			},
			newRulePriority: 4,
			expectedCommand: "-I INPUT 1 -j DROP -m comment --comment nordvpn-4",
		},
		{
			name: "insert rule in between",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-4 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
			},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 2 -j DROP -m comment --comment nordvpn-3",
		},
		{
			name:            "insert rule in empty iptables",
			rules:           []string{},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 1 -j DROP -m comment --comment nordvpn-3",
		},
		{
			name: "insert rule no nordvpn rules",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* other-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-2 */"},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 1 -j DROP -m comment --comment nordvpn-3",
		},
		{
			name: "insert with highest priority non-nordvpn rules at the bottom",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-2 */"},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 1 -j DROP -m comment --comment nordvpn-3",
		},
		{
			name: "insert with lowest priority non-nordvpn rules at the bottom",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-2 */"},
			newRulePriority: 0,
			expectedCommand: "-I INPUT 4 -j DROP -m comment --comment nordvpn-0",
		},
		{
			name: "insert in between non-nordvpn rules at the bottom",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-2 */"},
			newRulePriority: 2,
			expectedCommand: "-I INPUT 2 -j DROP -m comment --comment nordvpn-2",
		},
		{
			name: "insert with highest priority non-nordvpn in between",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* other-0 */",   // (1)
				"DROP       all  --  anywhere             anywhere             /* other-1 */",   // (2)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */", // nordvpn (3)
				"DROP       all  --  anywhere             anywhere             /* other-2 */",   // (4)
				"DROP       all  --  anywhere             anywhere             /* other-3 */",   // (5)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */", // nordvpn (6)
				"DROP       all  --  anywhere             anywhere             /* other-4 */",   // (7)
				"DROP       all  --  anywhere             anywhere             /* other-5 */",   // (8)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-0 */", // nordvpn (9)
			},
			newRulePriority: 4,
			expectedCommand: "-I INPUT 3 -j DROP -m comment --comment nordvpn-4",
		},
		{
			name: "insert with highest priority non-nordvpn in between",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* other-0 */",   // (1)
				"DROP       all  --  anywhere             anywhere             /* other-1 */",   // (2)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */", // nordvpn (3)
				"DROP       all  --  anywhere             anywhere             /* other-2 */",   // (4)
				"DROP       all  --  anywhere             anywhere             /* other-3 */",   // (5)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */", // nordvpn (6)
				"DROP       all  --  anywhere             anywhere             /* other-4 */",   // (7)
				"DROP       all  --  anywhere             anywhere             /* other-5 */",   // (8)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */", // nordvpn (9)
			},
			newRulePriority: 0,
			expectedCommand: "-I INPUT 10 -j DROP -m comment --comment nordvpn-0",
		},
		{
			name: "insert in between non-nordvpn in between",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* other-0 */",   // (1)
				"DROP       all  --  anywhere             anywhere             /* other-1 */",   // (2)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-4 */", // nordvpn (3)
				"DROP       all  --  anywhere             anywhere             /* other-2 */",   // (4)
				"DROP       all  --  anywhere             anywhere             /* other-3 */",   // (5)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */", // nordvpn (6)
				"DROP       all  --  anywhere             anywhere             /* other-4 */",   // (7)
				"DROP       all  --  anywhere             anywhere             /* other-5 */",   // (8)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */", // nordvpn (9)
			},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 6 -j DROP -m comment --comment nordvpn-3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chain := iptablesmock.NewIptablesOutput(iptablesmock.InputChainName)
			chain.AddRules(test.rules...)

			commandRunnerMock := iptablesmock.NewCommandRunnerMock()
			commandRunnerMock.AddIptablesListOutput(iptablesmock.InputChainName, chain.Get())

			iptablesManager := NewIPTablesManager(&commandRunnerMock, true, true)
			// nolint:errcheck // Tested in other uts
			iptablesManager.InsertRule(NewFwRule(
				Input,
				IPv4,
				"-j DROP",
				test.newRulePriority))

			commands := commandRunnerMock.PopIPv4Commands()
			assert.Len(t, commands, 1, "Only one command per rule insertion should be executed.")
			assert.Equal(t, test.expectedCommand, commands[0], "Invalid command executed when inserting a rule.")
		})
	}
}
