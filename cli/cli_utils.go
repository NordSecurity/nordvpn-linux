package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/urfave/cli/v2"
)

// settings returns the daemon Settings, fetching them once and memoizing the result on the
// cmd for reuse. The CLI evaluates several Hidden: cmd.Except(...) gates while building the
// command tree
func (c *cmd) settings() (*pb.Settings, error) {
	if c.settingsCache != nil {
		return c.settingsCache, nil
	}
	resp, err := c.client.Settings(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	c.settingsCache = resp.GetData()
	return c.settingsCache, nil
}

// Except returns true if technology in the app configuration is different than tech.
func (c *cmd) Except(tech config.Technology) bool {
	settings, err := c.settings()
	if err != nil {
		return false
	}
	return settings.GetTechnology() != tech
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
