package cli

import (
	"context"
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
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
		{
			name:         "no flag value given",
			args:         []string{"--" + flagName},
			expectedArgs: []string{},
		},
		{
			name:         "no flag value given preceding arg",
			args:         []string{argName, "--" + flagName},
			expectedArgs: []string{argName},
		},
		{
			name:         "no flag value given preceding arg same name as flag name",
			args:         []string{flagName, "--" + flagName},
			expectedArgs: []string{flagName},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := removeFlagFromArgs(test.args, flagName)
			assert.Equal(t, test.expectedArgs, result)
		})
	}
}

func TestReadForConfirmation(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		input          io.Reader
		expectedOutput bool
		expectedStatus bool
	}{
		{
			name:           "no input",
			input:          strings.NewReader(""),
			expectedOutput: false,
			expectedStatus: false,
		},
		{
			name:           "a newline",
			input:          strings.NewReader("\n"),
			expectedOutput: false,
			expectedStatus: false,
		},
		{
			name:           "a number",
			input:          strings.NewReader("5"),
			expectedOutput: false,
			expectedStatus: false,
		},
		{
			name:           "predicting Anton's input",
			input:          strings.NewReader("\\"),
			expectedOutput: false,
			expectedStatus: false,
		},
		{
			name:           "lowercase n",
			input:          strings.NewReader("n"),
			expectedOutput: false,
			expectedStatus: true,
		},
		{
			name:           "uppercase n",
			input:          strings.NewReader("N"),
			expectedOutput: false,
			expectedStatus: true,
		},
		{
			name:           "lowercase y",
			input:          strings.NewReader("y"),
			expectedOutput: true,
			expectedStatus: true,
		},
		{
			name:           "uppercase y",
			input:          strings.NewReader("Y"),
			expectedOutput: true,
			expectedStatus: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			answer, ok := readForConfirmation(test.input, test.name)
			assert.Equal(t, test.expectedOutput, answer)
			assert.Equal(t, test.expectedStatus, ok)
		})
	}
}

func TestReadForConfirmationDefaultValue(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		input        io.Reader
		defaultValue bool
	}{
		{
			name:         "no input, default to true",
			input:        strings.NewReader(""),
			defaultValue: true,
		},
		{
			name:         "no input, default to false",
			input:        strings.NewReader(""),
			defaultValue: false,
		},
		{
			name:         "a newline, default to true",
			input:        strings.NewReader("\n"),
			defaultValue: true,
		},
		{
			name:         "a newline, default to false",
			input:        strings.NewReader("\n"),
			defaultValue: false,
		},
		{
			name:         "a number, default to true",
			input:        strings.NewReader("5"),
			defaultValue: true,
		},
		{
			name:         "a number, default to false",
			input:        strings.NewReader("5"),
			defaultValue: false,
		},
		{
			name:         "predicting Anton's input, default to true",
			input:        strings.NewReader("\\"),
			defaultValue: true,
		},
		{
			name:         "predicting Anton's input, default to false",
			input:        strings.NewReader("\\"),
			defaultValue: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			answer := readForConfirmationDefaultValue(test.input, test.name, test.defaultValue)
			assert.Equal(t, test.defaultValue, answer)
		})
	}
}

// TestVersion checks how the version member for App.Version is constructed
func TestComposeAppVersion(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		buildVersion string
		env          string
		isUnderSnap  bool
		expected     string
	}{
		{
			name:         "version for deb/rpm production build",
			buildVersion: "1.2.3",
			env:          string(internal.Production),
			isUnderSnap:  false,
			expected:     "1.2.3",
		},
		{
			name:         "version for deb/rpm development build",
			buildVersion: "1.2.3",
			env:          string(internal.Development),
			isUnderSnap:  false,
			expected:     "1.2.3 - dev",
		},
		{
			name:         "version for snap production build",
			buildVersion: "1.2.3",
			env:          string(internal.Production),
			isUnderSnap:  true,
			expected:     "1.2.3 [snap]",
		},
		{
			name:         "version for snap development build",
			buildVersion: "1.2.3",
			env:          string(internal.Development),
			isUnderSnap:  true,
			expected:     "1.2.3 [snap] - dev",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := composeAppVersion(test.buildVersion, test.env, test.isUnderSnap)
			assert.Equal(t, test.expected, version)
		})
	}
}
