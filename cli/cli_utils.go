package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/urfave/cli/v2"
)

// Except returns true if technology in the app configuration is different than tech.
func (c *cmd) Except(tech config.Technology) bool {
	settings, err := c.client.Settings(context.Background(), &pb.Empty{})
	if err != nil {
		return false
	}
	return settings.GetData().Technology != tech
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
