package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetNotifyUsageText is shown next to notify command by nordvpn set --help
const SetNotifyUsageText = "Enables or disables notifications"

func (c *cmd) SetNotify(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetNotify(context.Background(), &pb.SetNotifyRequest{
		Uid:    int64(os.Getuid()),
		Notify: flag,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(SetNotifyNothingToSet, nstrings.GetBoolLabel(flag)))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(SetNotifySuccess, nstrings.GetBoolLabel(flag)))
	}
	return nil
}
