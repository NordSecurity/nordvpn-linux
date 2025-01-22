package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func (c *cmd) SetPostquantumVpn(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetPostQuantum(context.Background(), &pb.SetGenericRequest{Enabled: flag})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Post-quantum VPN", nstrings.GetBoolLabel(flag)))
		return nil
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodePqAndMeshnetSimultaneously:
		return formatError(errors.New(SetPqAndMeshnet))
	case internal.CodePqWithoutNordlynx:
		return formatError(fmt.Errorf(SetPqUnavailable, resp.Data[0]))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Post-quantum VPN", nstrings.GetBoolLabel(flag)))
		flag, _ := strconv.ParseBool(resp.Data[0])
		if flag {
			color.Yellow(SetReconnect)
		}
	}

	return nil
}
