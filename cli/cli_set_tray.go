package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetTrayUsageText is shown next to defaults command by nordvpn set --help
const SetTrayUsageText = "Enables or disables the NordVPN icon in the system tray. The icon provides quick access to basic controls and your VPN status details."

func (c *cmd) SetTray(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	daemonResp, err := c.client.SetTray(context.Background(), &pb.SetTrayRequest{
		Uid:  int64(os.Getuid()),
		Tray: flag,
	})
	if err != nil {
		return formatError(err)
	}

	switch daemonResp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(SetTrayNothingToSet, nstrings.GetBoolLabel(flag)))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(SetTraySuccess, nstrings.GetBoolLabel(flag)))
	}

	return nil
}
