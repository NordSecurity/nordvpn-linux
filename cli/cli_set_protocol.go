package cli

import (
	"context"
	"errors"
	"fmt"
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

func setProtocolCommonErrorCodeToError(code pb.SetErrorCode, args ...any) error {
	switch code {
	case pb.SetErrorCode_FAILURE:
		return formatError(internal.ErrUnhandled)
	case pb.SetErrorCode_CONFIG_ERROR:
		return formatError(ErrConfig)
	case pb.SetErrorCode_ALREADY_SET:
		return formatError(
			errors.New(color.YellowString(fmt.Sprintf(SetProtocolAlreadySet, args...))))
	}
	return nil
}

func handleSetProtocolStatus(code pb.SetProtocolStatus, protocol config.Protocol) error {
	switch code {
	case pb.SetProtocolStatus_INVALID_TECHNOLOGY:
		return fmt.Errorf(SetProtocolUnavailable)
	case pb.SetProtocolStatus_PROTOCOL_CONFIGURED_VPN_ON:
		color.Yellow(SetReconnect)
		fallthrough
	case pb.SetProtocolStatus_PROTOCOL_CONFIGURED:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Protocol", protocol.String()))
	}
	return nil
}

// SetProtocol
func (c *cmd) SetProtocol(ctx *cli.Context) error {
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

	resp, err := c.client.SetProtocol(context.Background(), &pb.SetProtocolRequest{
		Protocol: proto,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Response.(type) {
	case *pb.SetProtocolResponse_ErrorCode:
		return setProtocolCommonErrorCodeToError(resp.GetErrorCode(), proto.String())
	case *pb.SetProtocolResponse_SetProtocolStatus:
		return handleSetProtocolStatus(resp.GetSetProtocolStatus(), proto)
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
