package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func (c *cmd) SetARPIgnore(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetARPIgnore(context.Background(), &pb.SetGenericRequest{
		Enabled: flag,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(SetARPIgnoreNothingToSet, nstrings.GetBoolLabel(flag)))
	case internal.CodeSuccess:
		if !flag {
			color.Yellow(SetARPIgnoreWarning)
		}
		color.Green(fmt.Sprintf(SetARPIgnoreSuccess, nstrings.GetBoolLabel(flag)))
	}

	return nil
}
