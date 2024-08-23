package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/exp/slices"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Allowlist add port help text
const (
	AllowlistAddPortUsageText     = "Adds port to the allowlist"
	AllowlistAddPortArgsUsageText = `<port> [protocol <protocol>]`
	AllowlistAddPortDescription   = `Use this command to allowlist the UDP and TCP port.

Example: 'nordvpn allowlist add port 22'

Optionally, protocol can be provided to specify which protocol should be allowlisted.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn allowlist add port 22 protocol TCP'`
)

func (c *cmd) AllowlistAddPort(ctx *cli.Context) error {
	args := ctx.Args()
	if !(args.Len() == 1 || (args.Len() == 3 && args.Get(1) == AllowlistProtocol)) {
		return formatError(argsCountError(ctx))
	}

	port, err := strconv.ParseInt(args.First(), 10, 64)
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	if port < AllowlistMinPort || port > AllowlistMaxPort {
		return formatError(fmt.Errorf(
			AllowlistPortRangeError,
			port,
			AllowlistMinPort,
			AllowlistMaxPort,
		))
	}

	isUDP := false
	isTCP := false
	if args.Len() == 1 {
		isUDP = true
		isTCP = true
	} else {
		switch args.Get(2) {
		case config.Protocol_UDP.String():
			isUDP = true
		case config.Protocol_TCP.String():
			isTCP = true
		default:
			return formatError(argsParseError(ctx))
		}
	}

	settings, err := c.getSettings()
	if err != nil {
		return formatError(err)
	}
	allowlist := settings.Settings.GetAllowlist()
	if isTCP && slices.Contains(allowlist.Ports.Tcp, port) ||
		isUDP && slices.Contains(allowlist.Ports.Udp, port) {
		return formatError(fmt.Errorf(
			AllowlistAddPortExistsError,
			port,
			getProtocolStr(isTCP, isUDP),
		))
	}
	if isTCP {
		allowlist.Ports.Tcp = append(allowlist.Ports.Tcp, port)
	}
	if isUDP {
		allowlist.Ports.Udp = append(allowlist.Ports.Udp, port)
	}
	resp, err := c.client.SetAllowlist(context.Background(), &pb.SetAllowlistRequest{
		Allowlist: allowlist,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(
			AllowlistAddPortExistsError,
			port,
			getProtocolStr(isTCP, isUDP),
		))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(
			AllowlistAddPortSuccess,
			port,
			getProtocolStr(isTCP, isUDP),
		))
	}
	return nil
}

func (c *cmd) AllowlistAddPortAutoComplete(ctx *cli.Context) {
	switch ctx.NArg() {
	case 1:
		// show one word for completion
		fmt.Println(stringProtocol)
	case 2:
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

// getProtocolStr returns one of:
// * TCP
// * UDP
// * UDP|TCP
func getProtocolStr(isTCP bool, isUDP bool) string {
	if isTCP && !isUDP {
		return config.Protocol_TCP.String()
	} else if isUDP && !isTCP {
		return config.Protocol_UDP.String()
	}
	return fmt.Sprintf("%s|%s", config.Protocol_UDP, config.Protocol_TCP)
}
