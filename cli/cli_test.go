package cli

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCLICommands(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name: "account",
		},
		{
			Name: "cities",
		},
		{
			Name: "set",
			Subcommands: []*cli.Command{
				{
					Name:    "threatprotectionlite",
					Aliases: []string{"tplite", "tpl", "cybersec"},
				},
				{
					Name: "defaults",
				},
				{
					Name: "dns",
				},
				{
					Name: "firewall",
				},
				{
					Name: "technology",
				},
				{
					Name: "autoconnect",
					Subcommands: []*cli.Command{
						{
							Name: "arg",
						}},
				},
			},
		},
	}
	app.EnableBashCompletion = true

	ctx := &cli.Context{
		Context: context.Background(),
		App:     &cli.App{},
		Command: &cli.Command{},
	}
	set := flag.NewFlagSet(flagFilter, flag.ContinueOnError)
	ctx = cli.NewContext(app, set, ctx)

	tests := []struct {
		name     string
		appArgs  []string
		expected string
	}{
		{
			name:     "empty arguments",
			appArgs:  []string{},
			expected: "",
		},
		{
			name:     "one argument",
			appArgs:  []string{"nordvpn"},
			expected: "",
		},
		{
			name:     "nordvpn set",
			appArgs:  []string{"nordvp", "set"},
			expected: "set",
		},
		{
			name:     "nordvpn set dns",
			appArgs:  []string{"nordvp", "set", "dns"},
			expected: "set dns",
		},
		{
			name:     "nordvpn set dns 1234",
			appArgs:  []string{"nordvp", "set", "dns", "1234"},
			expected: "set dns",
		},
		{
			name:     "nordvpn set dns 1 2 3 4",
			appArgs:  []string{"nordvp", "set", "dns", "1", "2", "3", "4"},
			expected: "set dns",
		},
		{
			name:     "works with aliases",
			appArgs:  []string{"nordvp", "set", "tpl"},
			expected: "set threatprotectionlite",
		},
		{
			name:     "works with more subcommands levels",
			appArgs:  []string{"nordvp", "set", "autoconnect", "arg", "a"},
			expected: "set autoconnect arg",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, commandFullName(ctx, test.appArgs))
		})
	}
}
