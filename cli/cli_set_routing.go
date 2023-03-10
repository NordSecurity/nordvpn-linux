package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const SetRoutingUsageText = "Allows routing traffic through VPN " +
	"servers and peer devices in Meshnet. This setting must be " +
	"enabled to send your traffic through a VPN server or a peer " +
	"device. If the setting is disabled, the app will only " +
	"initiate necessary connections to a VPN server or a peer " +
	"device but won’t start traffic routing."

func (c *cmd) SetRouting(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetRouting(context.Background(), &pb.SetGenericRequest{Enabled: flag})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Routing", nstrings.GetBoolLabel(flag)))
	case internal.CodeDependencyError:
		color.Yellow(fmt.Sprintf(MsgInUse, "Routing", "meshnet"))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Routing", nstrings.GetBoolLabel(flag)))
	}
	return nil
}
