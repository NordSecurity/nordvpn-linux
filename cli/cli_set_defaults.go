package cli

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetDefaultsUsageText is shown next to defaults command by nordvpn set --help
const SetDefaultsUsageText = "Restores settings to their default values."

func (c *cmd) SetDefaults(ctx *cli.Context) error {
	resp, err := c.client.SetDefaults(context.Background(), &pb.SetDefaultsRequest{NoLogout: false})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeSuccess:
		color.Green(SetDefaultsSuccess)
	}
	return nil
}
