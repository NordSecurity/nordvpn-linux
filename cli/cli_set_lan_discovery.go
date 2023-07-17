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

func SetLANDiscoveryErrorCodeToError(code pb.SetErrorCode, args ...any) error {
	switch code {
	case pb.SetErrorCode_FAILURE:
		return formatError(internal.ErrUnhandled)
	case pb.SetErrorCode_CONFIG_ERROR:
		return formatError(ErrConfig)
	case pb.SetErrorCode_ALREADY_SET:
		return formatError(
			errors.New(color.YellowString(fmt.Sprintf(SetLANDiscoveryAlreadyEnabled, args...))))
	}
	return nil
}

func (c *cmd) SetLANDiscovery(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsCountError(ctx))
	}

	arg := ctx.Args().First()
	flag, err := nstrings.BoolFromString(arg)
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetLANDiscovery(context.Background(), &pb.SetLANDiscoveryRequest{
		Enabled: flag,
	})

	switch resp.Response.(type) {
	case *pb.SetLANDiscoveryResponse_ErrorCode:
		return SetLANDiscoveryErrorCodeToError(resp.GetErrorCode(), arg)
	case *pb.SetLANDiscoveryResponse_SetLanDiscoveryStatus:
		color.Green(fmt.Sprintf(MsgSetSuccess, "LAN Discovery", nstrings.GetBoolLabel(flag)))
	}

	return nil
}
