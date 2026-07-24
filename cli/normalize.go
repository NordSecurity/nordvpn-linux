package cli

import (
	"strings"

	"github.com/urfave/cli/v2"
)

// Names of the implicit help command that urfave/cli injects at runtime; it is
// not present in the static command tree, so normalization handles it by name.
const (
	helpName  = "help"
	helpAlias = "h"
)

// NormalizeCommandCase canonicalizes the case of command and subcommand name
// tokens in args so that commands can be invoked case-insensitively, e.g.
// "SET MESH on" and "Set Mesh On" both resolve to "set meshnet on".
//
// Only tokens that name a command/subcommand (matched case-insensitively against
// the command tree, including aliases) are rewritten to their canonical name.
// Every other token — flags, flag values, server tags, hostnames, file paths,
// on/off values — is left byte-for-byte untouched, so case-sensitive argument
// values are never corrupted.
//
// commands is the top-level command list (app.Commands); args is the full
// argument vector including the program name at index 0.
func NormalizeCommandCase(commands []*cli.Command, args []string) []string {
	if len(args) == 0 {
		return args
	}

	out := make([]string, 0, len(args))
	out = append(out, args[0]) // program name, verbatim

	level := commands
	matching := true
	for _, arg := range args[1:] {
		if !matching {
			out = append(out, arg)
			continue
		}

		// Flags may precede a subcommand; pass them through without treating
		// them as a command boundary or descending the tree.
		if strings.HasPrefix(arg, "-") {
			out = append(out, arg)
			continue
		}

		if cmd := findCommandFold(level, arg); cmd != nil {
			out = append(out, cmd.Name)
			level = cmd.Subcommands
			continue
		}

		// urfave/cli injects an implicit "help" command (alias "h") into the app
		// and into every command group that has subcommands. That injection
		// happens inside Run (after this normalization), so the help command is
		// never present in the tree above. Recognize it explicitly, mirroring
		// urfave's rule that help exists wherever the current level has commands.
		// help's argument is a sibling command name, so keep the current level.
		if len(level) > 0 && (strings.EqualFold(arg, helpName) || strings.EqualFold(arg, helpAlias)) {
			out = append(out, helpName)
			continue
		}

		// First token that isn't a command: it and everything after it are
		// arguments/values and must be preserved verbatim.
		out = append(out, arg)
		matching = false
	}

	return out
}

// findCommandFold returns the command in commands whose canonical name or any of
// its aliases matches name case-insensitively, or nil if none match.
func findCommandFold(commands []*cli.Command, name string) *cli.Command {
	for _, cmd := range commands {
		if strings.EqualFold(cmd.Name, name) {
			return cmd
		}
		for _, alias := range cmd.Aliases {
			if strings.EqualFold(alias, name) {
				return cmd
			}
		}
	}
	return nil
}
