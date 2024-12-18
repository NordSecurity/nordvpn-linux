package cli

import (
	"context"
	"io"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// RegisterUsageText is shown next to register command by nordvpn --help
const RegisterUsageText = "Registers a new user account"

func (c *cmd) Register(ctx *cli.Context) error {
	cl, err := c.client.LoginOAuth2(
		context.Background(),
		&pb.LoginOAuth2Request{
			Type: pb.LoginType_LoginType_SIGNUP,
		},
	)
	if err != nil {
		return formatError(err)
	}

	for {
		resp, err := cl.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return formatError(err)
		}
		if url := resp.GetData(); url != "" {
			color.Green("Continue in the browser: %s", url)
		}
	}

	return nil
}
