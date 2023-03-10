package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// WhitelistRemoveAllUsageText is shown next to all command by nordvpn whitelist remove --help
const WhitelistRemoveAllUsageText = "Removes all ports and subnets from the whitelist"

func (c *cmd) WhitelistRemoveAll(ctx *cli.Context) error {
	c.config.Whitelist.Ports.UDP.Clear()
	c.config.Whitelist.Ports.TCP.Clear()
	c.config.Whitelist.Subnets.Clear()

	resp, err := c.client.SetWhitelist(context.Background(), &pb.SetWhitelistRequest{
		Whitelist: &pb.Whitelist{
			Ports: &pb.Ports{
				Udp: client.SetToInt64s(c.config.Whitelist.Ports.UDP),
				Tcp: client.SetToInt64s(c.config.Whitelist.Ports.TCP),
			},
			Subnets: internal.SetToStrings(c.config.Whitelist.Subnets),
		},
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(WhitelistRemoveAllError))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(WhitelistRemoveAllSuccess))
	}

	return nil
}
