package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func (c *cmd) SetSanitization(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetSanitization(context.Background(), &pb.SetGenericRequest{
		Enabled: flag,
	})
	if err != nil {
		return formatError(err)
	}

	sanitizationState := "disabled"
	if flag {
		sanitizationState = "enabled"
	}

	switch resp.Type {
	case internal.CodeSuccess:
		if !flag {
			color.Yellow(MsgDisableSanitizationWarning)
		}
		color.Green(MsgDisableSanitizationSuccess, sanitizationState)
	case internal.CodeNothingToDo:
		return fmt.Errorf(MsgDisableSanitizationNothingToDo, sanitizationState)
	case internal.CodeDisableSanitizationNonRoot:
		return errors.New(MsgDisableSanitizationNonRoot)
	case internal.CodeInternalError:
		return errors.New(internal.UnhandledMessage)
	}

	return nil
}
