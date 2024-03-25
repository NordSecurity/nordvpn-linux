package cli

import "github.com/urfave/cli/v2"

// User is a dummy command, called by an autostart desktop to start up norduser process in snap environment.
// norduser is started on any cli command, so there is no need for this function do do anything.
func (c *cmd) User(ctx *cli.Context) error {
	return nil
}
