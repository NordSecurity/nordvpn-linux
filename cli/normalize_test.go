package cli

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func testCommandTree() []*cli.Command {
	return []*cli.Command{
		{
			Name:    "connect",
			Aliases: []string{"c"},
		},
		{
			Name:    "set",
			Aliases: []string{"s"},
			Subcommands: []*cli.Command{
				{Name: "killswitch"},
				{Name: "autoconnect"},
			},
		},
		{
			Name:    "meshnet",
			Aliases: []string{"mesh"},
			Subcommands: []*cli.Command{
				{
					Name: "peer",
					Subcommands: []*cli.Command{
						{Name: "list"},
						{Name: "remove"},
					},
				},
				{Name: "set"},
			},
		},
		{
			Name: "login",
		},
	}
}

func TestNormalizeCommandCase(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "already lowercase",
			input:    []string{"nordvpn", "set", "killswitch", "on"},
			expected: []string{"nordvpn", "set", "killswitch", "on"},
		},
		{
			name:     "all uppercase commands",
			input:    []string{"nordvpn", "SET", "KILLSWITCH", "on"},
			expected: []string{"nordvpn", "set", "killswitch", "on"},
		},
		{
			name:     "mixed case commands",
			input:    []string{"nordvpn", "Set", "Killswitch", "On"},
			expected: []string{"nordvpn", "set", "killswitch", "On"},
		},
		{
			name:     "alias canonicalization",
			input:    []string{"nordvpn", "S", "killswitch", "on"},
			expected: []string{"nordvpn", "set", "killswitch", "on"},
		},
		{
			name:     "meshnet alias case-insensitive",
			input:    []string{"nordvpn", "MESH", "peer", "LIST"},
			expected: []string{"nordvpn", "meshnet", "peer", "list"},
		},
		{
			name:     "deep nesting",
			input:    []string{"nordvpn", "Mesh", "Peer", "Remove", "MyPeer-Host"},
			expected: []string{"nordvpn", "meshnet", "peer", "remove", "MyPeer-Host"},
		},
		{
			name:     "value case preserved after command stops matching",
			input:    []string{"nordvpn", "CONNECT", "United_States"},
			expected: []string{"nordvpn", "connect", "United_States"},
		},
		{
			name:     "flag before value on command without subcommands",
			input:    []string{"nordvpn", "Connect", "--group", "Double_VPN", "United_States"},
			expected: []string{"nordvpn", "connect", "--group", "Double_VPN", "United_States"},
		},
		{
			name:     "login token stays case-sensitive",
			input:    []string{"nordvpn", "LOGIN", "--token", "AbCdEf123XYZ"},
			expected: []string{"nordvpn", "login", "--token", "AbCdEf123XYZ"},
		},
		{
			name:     "unknown command left untouched",
			input:    []string{"nordvpn", "Foobar", "Baz"},
			expected: []string{"nordvpn", "Foobar", "Baz"},
		},
		{
			name:     "only program name",
			input:    []string{"nordvpn"},
			expected: []string{"nordvpn"},
		},
		{
			name:     "empty args",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "argument matching a nested command name is not descended past a value",
			input:    []string{"nordvpn", "SET", "autoconnect", "ON", "set"},
			expected: []string{"nordvpn", "set", "autoconnect", "ON", "set"},
		},
		{
			name:     "help as trailing subcommand uppercase",
			input:    []string{"nordvpn", "mesh", "HELP"},
			expected: []string{"nordvpn", "meshnet", "help"},
		},
		{
			name:     "help mixed case on command group",
			input:    []string{"nordvpn", "MESH", "Help"},
			expected: []string{"nordvpn", "meshnet", "help"},
		},
		{
			name:     "help alias h canonicalized",
			input:    []string{"nordvpn", "meshnet", "H"},
			expected: []string{"nordvpn", "meshnet", "help"},
		},
		{
			name:     "top-level help uppercase",
			input:    []string{"nordvpn", "HELP"},
			expected: []string{"nordvpn", "help"},
		},
		{
			name:     "help followed by sibling command is canonicalized",
			input:    []string{"nordvpn", "HELP", "MESH"},
			expected: []string{"nordvpn", "help", "meshnet"},
		},
		{
			name:     "help on nested group canonicalizes following subcommand",
			input:    []string{"nordvpn", "MESH", "HELP", "PEER"},
			expected: []string{"nordvpn", "meshnet", "help", "peer"},
		},
		{
			name:     "help not recognized on leaf command without subcommands",
			input:    []string{"nordvpn", "LOGIN", "HELP"},
			expected: []string{"nordvpn", "login", "HELP"},
		},
	}

	tree := testCommandTree()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeCommandCase(tree, tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
