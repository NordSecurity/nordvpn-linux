package cli

import (
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/urfave/cli/v2"
)

var UpdateMessagePrinted bool

func (c cmd) BeforeGeneral(ctx *cli.Context) error {
	if internal.IsProdEnv(string(c.environment)) {
		return fmt.Errorf(NoSuchCommand, strings.Join(ctx.Lineage()[1].Args().Slice(), " "))
	}
	return nil
}

func (c *cmd) HiddenGeneral() bool {
	return internal.IsProdEnv(string(c.environment))
}

func (c *cmd) Except(tech config.Technology) bool {
	return c.config.Technology != tech
}

func (c *cmd) ExceptDevelopment(tech config.Technology) bool {
	return c.Except(tech) || c.HiddenGeneral()
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
