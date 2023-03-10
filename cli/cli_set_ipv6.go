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

const SetIpv6UsageText = "Enables or disables use of the ipv6."

func (c *cmd) SetIpv6(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetIpv6(context.Background(), &pb.SetGenericRequest{Enabled: flag})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "IPv6", nstrings.GetBoolLabel(flag)))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "IPv6", nstrings.GetBoolLabel(flag)))
	}
	return nil
}
