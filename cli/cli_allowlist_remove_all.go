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

// AllowlistRemoveAllUsageText is shown next to all command by nordvpn allowlist remove --help
const AllowlistRemoveAllUsageText = "Removes all ports and subnets from the allowlist"

func (c *cmd) AllowlistRemoveAll(ctx *cli.Context) error {
	c.config.Allowlist.Ports.UDP.Clear()
	c.config.Allowlist.Ports.TCP.Clear()
	c.config.Allowlist.Subnets.Clear()

	resp, err := c.client.SetAllowlist(context.Background(), &pb.SetAllowlistRequest{
		Allowlist: &pb.Allowlist{
			Ports: &pb.Ports{
				Udp: client.SetToInt64s(c.config.Allowlist.Ports.UDP),
				Tcp: client.SetToInt64s(c.config.Allowlist.Ports.TCP),
			},
			Subnets: internal.SetToStrings(c.config.Allowlist.Subnets),
		},
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(AllowlistRemoveAllError))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(AllowlistRemoveAllSuccess))
	}

	return nil
}
