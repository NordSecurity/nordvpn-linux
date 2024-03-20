package cli

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func (c *cmd) MeshSet(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	action := c.meshClient.DisableMeshnet
	if flag {
		action = c.meshClient.EnableMeshnet
	}

	resp, err := action(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}
	if err := MeshnetResponseToError(resp); err != nil {
		return formatError(err)
	}

	color.Green(MsgSetMeshnetSuccess, nstrings.GetBoolLabel(flag))

	return nil
}
