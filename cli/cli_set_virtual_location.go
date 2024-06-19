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

const (
	// SetVirtualLocationUsageText is shown next to defaults command by nordvpn set --help and for autocomplete
	MsgSetVirtualLocationUsageText   = "TODO: MsgSetVirtualLocationUsageText"
	MsgSetVirtualLocationDescription = "TODO: MsgSetVirtualLocationDescription"
)

func (c *cmd) SetVirtualLocation(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	response, err := c.client.SetVirtualLocation(context.Background(), &pb.SetGenericRequest{Enabled: flag})
	if err != nil {
		return formatError(err)
	}

	switch response.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		// TODO: approve by UX
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Virtual location", nstrings.GetBoolLabel(flag)))
	case internal.CodeSuccess:
		// TODO: approve by UX
		color.Green(fmt.Sprintf(MsgSetSuccess, "Virtual location", nstrings.GetBoolLabel(flag)))
	}
	return nil
}
