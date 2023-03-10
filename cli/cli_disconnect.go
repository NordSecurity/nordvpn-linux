package cli

import (
	"context"
	"io"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// DisconnectUsageText is shown next to disconnect command by nordvpn --help
const DisconnectUsageText = "Disconnects you from VPN"

func (c *cmd) Disconnect(ctx *cli.Context) error {
	resp, err := c.client.Disconnect(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return formatError(err)
		}

		switch out.Type {
		case internal.CodeVPNNotRunning:
			color.Yellow(DisconnectNotConnected)
		case internal.CodeDisconnected:
			color.Green(internal.DisconnectSuccess)
			color.Yellow(DisconnectConnectionRating, ctx.App.Name)
		}
	}
	return nil
}
