package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetObfuscateUsageText is shown next to obfuscate command by nordvpn set --help
const SetObfuscateUsageText = "Enables or disables obfuscation. When enabled, this feature allows to bypass network traffic sensors which aim to detect usage of the protocol and log, throttle or block it."

func (c *cmd) SetObfuscate(ctx *cli.Context) error {
	if err := c.BeforeSetObfuscate(ctx); err != nil {
		return formatError(err)
	}

	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetObfuscate(context.Background(), &pb.SetGenericRequest{Enabled: flag})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Obfuscation", nstrings.GetBoolLabel(flag)))
		return nil
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeAutoConnectServerNotObfuscated:
		return formatError(errors.New(ObfuscateOnServerNotObfuscated))
	case internal.CodeAutoConnectServerObfuscated:
		return formatError(errors.New(ObfuscateOffServerObfuscated))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Obfuscation", nstrings.GetBoolLabel(flag)))
		flag, _ := strconv.ParseBool(resp.Data[0])
		if flag {
			color.Yellow(SetReconnect)
		}
	}
	return nil
}

func (c *cmd) BeforeSetObfuscate(ctx *cli.Context) error {
	resp, err := c.client.Settings(context.Background(), &pb.SettingsRequest{Uid: int64(os.Getuid())})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return ErrConfig
	case internal.CodeSuccess:
		break
	default:
		return internal.ErrUnhandled
	}

	if resp.Data.Technology != config.Technology_OPENVPN {
		return formatError(errors.New(SetObfuscateUnavailable))
	}
	return nil
}
