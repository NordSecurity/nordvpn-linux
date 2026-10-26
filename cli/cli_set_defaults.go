package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	// SetDefaultsUsageText is shown next to defaults command by nordvpn set --help
	SetDefaultsUsageText = "Restores settings to their default values."
	flagLogout           = "logout"
	flagOffKillswitch    = "off-killswitch"
)

func (c *cmd) SetDefaults(ctx *cli.Context) error {
	logout := ctx.Bool(flagLogout)
	offKillswitch := ctx.Bool(flagOffKillswitch)

	resp, err := c.client.SetDefaults(context.Background(), &pb.SetDefaultsRequest{NoLogout: !logout, OffKillswitch: offKillswitch})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeSuccess:
		color.Green(SetDefaultsSuccess)
	case internal.CodeSetDefaultsNotApplied:
		if len(resp.Data) == 0 {
			return formatError(errors.New(SetDefaultsNetworkSettingsNotApplied))
		}
		return formatError(fmt.Errorf(SetDefaultsPartialSuccess, strings.Join(resp.Data, ", ")))
	case internal.CodeCleanRecentConnectionError:
		return formatError(errors.New(client.RecentConnectionErrorMessage))
	default:
		return formatError(internal.ErrUnhandled)
	}
	return nil
}
