package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetKillSwitchUsageText is shown next to killswitch command by nordvpn set --help
const SetKillSwitchUsageText = "Enables or disables Kill Switch. This security feature blocks your device from accessing the Internet while not connected to the VPN or in case connection with a VPN server is lost."

func (c *cmd) SetKillSwitch(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetKillSwitch(context.Background(), &pb.SetKillSwitchRequest{
		KillSwitch: flag,
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
	case internal.CodeVPNMisconfig, internal.CodeKillSwitchError, internal.CodeFailure:
		return formatError(internal.ErrUnhandled)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Kill Switch", nstrings.GetBoolLabel(flag)))
	case internal.CodeDependencyError:
		color.Yellow(fmt.Sprintf(FirewallRequired, "killswitch"))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Kill Switch", nstrings.GetBoolLabel(flag)))
	}
	return nil
}
