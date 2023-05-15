package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetProtocolUsageText is shown next to protocol command by nordvpn set --help
const SetProtocolUsageText = "Sets the protocol"

// SetProtocolArgsUsageText is shown by nordvpn set protocol --help
const SetProtocolArgsUsageText = `<protocol>

Use this command to set the protocol to TCP or UDP.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn set protocol TCP'`

// SetProtocol
func (c *cmd) SetProtocol(ctx *cli.Context) error {
	if err := c.BeforeSetProtocol(ctx); err != nil {
		return formatError(err)
	}

	switch ctx.NArg() {
	case 0:
		return formatError(argsCountError(ctx))
	case 1:
	default:
		return formatError(argsParseError(ctx))
	}

	args := ctx.Args()
	var proto config.Protocol
	switch strings.ToUpper(args.First()) {
	case config.Protocol_UDP.String():
		proto = config.Protocol_UDP
	case config.Protocol_TCP.String():
		proto = config.Protocol_TCP
	default:
		return formatError(argsParseError(ctx))
	}

	if c.config.Protocol == proto {
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Protocol", proto.String()))
		return nil
	}

	resp, err := c.client.SetProtocol(context.Background(), &pb.SetProtocolRequest{
		Protocol: proto,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeSuccess:
		c.config.Protocol = proto
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(MsgSetSuccess, "Protocol", proto.String()))
		flag, _ := strconv.ParseBool(resp.Data[0])
		if flag {
			color.Yellow(SetReconnect)
		}
	}
	return nil
}

func (c *cmd) BeforeSetProtocol(ctx *cli.Context) error {
	resp, err := c.client.Settings(context.Background(), &pb.SettingsRequest{Uid: int64(os.Getuid())})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return ErrConfig
	case internal.CodeSuccessWithoutAC:
		// must be right before CodeSuccess
		color.Yellow(SetAutoConnectForceOff)
		fallthrough
	case internal.CodeSuccess:
		break
	default:
		return internal.ErrUnhandled
	}

	if resp.Data.Technology != config.Technology_OPENVPN {
		return fmt.Errorf(SetProtocolUnavailable)
	}
	return nil
}

func (c *cmd) SetProtocolAutoComplete(ctx *cli.Context) {
	resp, err := c.client.SettingsProtocols(context.Background(), &pb.Empty{})
	if err != nil {
		return
	}

	for _, item := range resp.Data {
		fmt.Println(item)
	}
}
