package cli

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func (c *cmd) MeshMount(ctx *cli.Context) error {
	resp, err := c.fileshareClient.Mount(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}
	if err := getFileshareResponseToError(resp); err != nil {
		return formatError(err)
	}
	color.Green(MsgMeshnetMountSuccess)
	return nil
}

func (c *cmd) MeshUnmount(ctx *cli.Context) error {
	resp, err := c.fileshareClient.Unmount(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}
	if err := getFileshareResponseToError(resp); err != nil {
		return formatError(err)
	}
	color.Green(MsgMeshnetUnmountSuccess)
	return nil
}
