package cli

import (
	"context"
	"flag"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCLICommands(t *testing.T) {
	category.Set(t, category.Unit)

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
			name:     "only app name is into the list",
			appArgs:  []string{"nordvpn"},
			expected: "",
		},
		{
			name:     "app name and one subcommand works",
			appArgs:  []string{"nordvp", "set"},
			expected: ctx.App.Name + " set",
		},
		{
			name:     "app name and 2 subcommands",
			appArgs:  []string{"nordvp", "set", "dns"},
			expected: ctx.App.Name + " set dns",
		},
		{
			name:     "one argument for command is not returned",
			appArgs:  []string{"nordvp", "set", "dns", "1234"},
			expected: ctx.App.Name + " set dns",
		},
		{
			name:     "multiple arguments are the command are not returned",
			appArgs:  []string{"nordvp", "set", "firewall", "1", "2", "3", "4"},
			expected: ctx.App.Name + " set firewall",
		},
		{
			name:     "works with aliases",
			appArgs:  []string{"nordvp", "set", "tpl"},
			expected: ctx.App.Name + " set threatprotectionlite",
		},
		{
			name:     "works with more subcommands levels",
			appArgs:  []string{"nordvp", "set", "autoconnect", "arg", "a"},
			expected: ctx.App.Name + " set autoconnect arg",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, commandFullName(ctx, test.appArgs))
		})
	}
}
