package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetTechnologyUsageText is shown next to technology command by nordvpn set --help
const SetTechnologyUsageText = "Sets the technology"

// SetTechnologyArgsUsageText is shown by nordvpn set technology --help
const SetTechnologyArgsUsageText = `<technology>

Use this command to set the technology.
Supported values for <technology>: OpenVPN or NordLynx.

Example: 'nordvpn set technology OpenVPN'`

func (c *cmd) SetTechnology(ctx *cli.Context) error {
	args := ctx.Args()

	switch args.Len() {
	case 0:
		return formatError(argsCountError(ctx))
	case 1:
	default:
		return formatError(argsParseError(ctx))
	}

	var tech config.Technology
	switch strings.ToUpper(args.First()) {
	case config.Technology_OPENVPN.String():
		tech = config.Technology_OPENVPN
	case config.Technology_NORDLYNX.String():
		tech = config.Technology_NORDLYNX
	default:
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetTechnology(context.Background(), &pb.SetTechnologyRequest{
		Technology: tech,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFormatError:
		return formatError(argsParseError(ctx))
	case internal.CodeDependencyError:
		return formatError(fmt.Errorf(SetTechnologyDepsError, internal.StringsToInterfaces(resp.Data)...))
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Technology", strings.Join(resp.Data, " ")))
	case internal.CodeSuccessWithoutAC:
		// must be right before CodeSuccess
		color.Yellow(SetAutoConnectForceOff)
		fallthrough
	case internal.CodeSuccess:
		flag, _ := strconv.ParseBool(resp.Data[0])
		color.Green(fmt.Sprintf(MsgSetSuccess, "Technology", strings.Join(resp.Data[1:], " ")))
		c.config.Technology = tech
		c.config.Obfuscate = false
		if flag {
			color.Yellow(SetReconnect)
		}
	}
	return c.configManager.Save(c.config)
}

func (c *cmd) SetTechnologyAutoComplete(ctx *cli.Context) {
	resp, err := c.client.SettingsTechnologies(context.Background(), &pb.Empty{})
	if err != nil {
		return
	}

	for _, item := range resp.Data {
		fmt.Println(item)
	}
}
