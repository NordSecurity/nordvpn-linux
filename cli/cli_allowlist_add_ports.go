package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Allowlist add ports help text
const (
	AllowlistAddPortsUsageText     = "Adds port range to the allowlist"
	AllowlistAddPortsArgsUsageText = `<port_from> <port_to> [protocol <protocol>]`
	AllowlistAddPortsDescription   = `Use this command to allowlist the UDP and TCP ports.

Example: 'nordvpn allowlist add ports 3000 8000'

Optionally, protocol can be provided to specify which protocol should be allowlisted.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn allowlist add ports 3000 8000 protocol TCP'`
)

func (c *cmd) AllowlistAddPorts(ctx *cli.Context) error {
	args := ctx.Args()
	if !(args.Len() == 2 || (args.Len() == 4 && args.Get(2) == AllowlistProtocol)) {
		return formatError(argsCountError(ctx))
	}

	startPort, err := strconv.ParseInt(args.First(), 10, 64)
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	endPort, err := strconv.ParseInt(args.Get(1), 10, 64)
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	isUDP := false
	isTCP := false

	if args.Len() == 2 {
		isUDP = true
		isTCP = true
	} else {
		switch args.Get(3) {
		case config.Protocol_UDP.String():
			isUDP = true
		case config.Protocol_TCP.String():
			isTCP = true
		default:
			return formatError(argsParseError(ctx))
		}
	}

	resp, err := c.client.SetAllowlist(context.Background(), &pb.SetAllowlistRequest{
		Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
			SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
				IsUdp: isUDP,
				IsTcp: isTCP,
				PortRange: &pb.PortRange{
					Start: startPort,
					Stop:  endPort,
				},
			},
		},
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(
			AllowlistAddPortsExistsError,
			startPort,
			endPort,
			getProtocolStr(isTCP, isUDP),
		))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeAllowlistPortOutOfRange:
		return formatError(fmt.Errorf(
			AllowlistPortsRangeError,
			startPort,
			endPort,
			internal.AllowlistMinPort,
			internal.AllowlistMaxPort,
		))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(
			AllowlistAddPortsSuccess,
			startPort,
			endPort,
			getProtocolStr(isTCP, isUDP),
		))
	}
	return nil
}

func (c *cmd) AllowlistAddPortsAutoComplete(ctx *cli.Context) {
	switch ctx.NArg() {
	case 2:
		// show one word for completion
		fmt.Println(stringProtocol)
	case 3:
		// show available protocols
		resp, err := c.client.SettingsProtocols(context.Background(), &pb.Empty{})
		if err != nil {
			return
		}

		for _, item := range resp.Data {
			fmt.Println(item)
		}
	default:
		return
	}
}
