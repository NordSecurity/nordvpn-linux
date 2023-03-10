package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// Click is hidden cmd used to open some info when clicking on desktop icon
func (c *cmd) Click(ctx *cli.Context) error {
	if ctx.NArg() >= 1 {
		// if arg is given
		// run the same as: login --callback %arg
		return c.oauth2(ctx)
	}

	if err := cli.ShowAppHelp(ctx); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Press 'Enter' to close this window.")
	fmt.Println("To use nordvpn - open new terminal and type command e.g. nordvpn connect")

	_, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	return err
}
