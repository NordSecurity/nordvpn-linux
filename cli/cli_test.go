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
			expected: ctx.App.Name,
		},
		{
			name:     "only app name is into the list",
			appArgs:  []string{"nordvpn"},
			expected: ctx.App.Name,
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

func Test_removeFlagFromArgs(t *testing.T) {
	category.Set(t, category.Unit)

	const flagName = "flag"
	const argName = "arg1"
	const otherArgName1 = "arg2"
	const otherArgName2 = "arg2"

	tests := []struct {
		name         string
		args         []string
		expectedArgs []string
	}{
		{
			name:         "only flag args",
			args:         []string{"--" + flagName, argName},
			expectedArgs: []string{},
		},
		{
			name:         "flag arg and preceding arg",
			args:         []string{otherArgName1, "--" + flagName, argName},
			expectedArgs: []string{otherArgName1},
		},
		{
			name:         "flag arg and succeeding arg",
			args:         []string{"--" + flagName, argName, otherArgName1},
			expectedArgs: []string{otherArgName1},
		},
		{
			name:         "flag arg and succeeding/preceding arg",
			args:         []string{otherArgName1, "--" + flagName, argName, otherArgName2},
			expectedArgs: []string{otherArgName1, otherArgName2},
		},
		{
			name:         "flag arg and preceding arg of the same name",
			args:         []string{argName, "--" + flagName, argName},
			expectedArgs: []string{argName},
		},
		{
			name:         "flag arg and succeeding arg of the same name",
			args:         []string{"--" + flagName, argName, argName},
			expectedArgs: []string{argName},
		},
		{
			name:         "flag arg and succeeding/preceding arg of the same name",
			args:         []string{argName, "--" + flagName, argName, argName},
			expectedArgs: []string{argName, argName},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := removeFlagFromArgs(test.args, flagName)
			assert.Equal(t, test.expectedArgs, result)
		})
	}
}
