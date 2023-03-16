package cli

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/urfave/cli/v2"
)

func (c *cmd) Except(tech config.Technology) bool {
	return c.config.Technology != tech
}

// SetBoolAutocomplete shows booleans suggestions
func (c *cmd) SetBoolAutocomplete(ctx *cli.Context) {
	if ctx.NArg() > 0 {
		return
	}
	for _, v := range nstrings.GetBools() {
		fmt.Println(v)
	}
}
