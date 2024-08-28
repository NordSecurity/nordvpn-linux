package cli

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/exp/slices"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Allowlist remove ports help text
const (
	AllowlistRemovePortsUsageText       = "Removes port range from the allowlist"
	AllowlistRemovePortsArgsUsageText   = `<port_from> <port_to> [protocol <protocol>]`
	AllowlistRemovePortsArgsDescription = `Use this command to remove ports from the allowlist.

Example: 'nordvpn allowlist remove ports 3000 8000'

Optionally, protocol can be provided to specify which protocol should be removed from the allowlist.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn allowlist remove ports 3000 8000 protocol TCP'`
)

func (c *cmd) AllowlistRemovePorts(ctx *cli.Context) error {
	args := ctx.Args()

	if !(args.Len() == 2 || (args.Len() == 4 && args.Get(2) == AllowlistProtocol)) {
		return formatError(argsCountError(ctx))
	}

	startPort, err := strconv.ParseInt(args.First(), 10, 64)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return formatError(argsParseError(ctx))
	}

	endPort, err := strconv.ParseInt(args.Get(1), 10, 64)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
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

	resp, err := c.client.UnsetAllowlist(context.Background(), &pb.SetAllowlistRequest{
		Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
			SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
				IsUdp: isUDP,
				IsTcp: isTCP,
				PortRange: &pb.PortRange{
					StartPort: startPort,
					EndPort:   startPort,
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
			AllowlistRemovePortsExistsError,
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
	case internal.CodeAllowlistPortNoop:
		return formatError(fmt.Errorf(
			AllowlistRemovePortsExistsError,
			startPort,
			endPort,
			getProtocolStr(isTCP, isUDP),
		))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(
			AllowlistRemovePortsSuccess,
			startPort,
			endPort,
			getProtocolStr(isTCP, isUDP),
		))
	}
	return nil
}

func (c *cmd) AllowlistRemovePortsAutoComplete(ctx *cli.Context) {
	settings, err := c.client.Settings(context.Background(), &pb.SettingsRequest{})
	if err != nil {
		return
	}
	allowlist := settings.GetData().Settings.GetAllowlist()

	switch ctx.NArg() {
	case 0:
		// create config after auth
		ports := append(allowlist.Ports.Udp, allowlist.Ports.Tcp...)
		slices.Sort(ports)
		ports = slices.Compact(ports)
		for _, port := range ports {
			fmt.Println(port)
		}
	case 1:
		startPort, err := strconv.ParseInt(ctx.Args().First(), 10, 64)
		if err != nil {
			return
		}

		ports := append(allowlist.Ports.Udp, allowlist.Ports.Tcp...)
		slices.Sort(ports)
		ports = slices.Compact(ports)
		for _, port := range ports {
			if startPort <= port {
				fmt.Println(port)
			}
		}
	case 2:
		fmt.Println(stringProtocol)
	case 3:
		startPort, err := strconv.ParseInt(ctx.Args().First(), 10, 64)
		if err != nil {
			return
		}

		if slices.Contains(allowlist.Ports.Udp, startPort) {
			fmt.Println(config.Protocol_UDP.String())
		}
		if slices.Contains(allowlist.Ports.Tcp, startPort) {
			fmt.Println(config.Protocol_TCP.String())
		}
	default:
		return
	}
}
