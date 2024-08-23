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

// Allowlist remove port help text
const (
	AllowlistRemovePortUsageText       = "Removes port from the allowlist"
	AllowlistRemovePortArgsUsageText   = `<port> [protocol <protocol>]`
	AllowlistRemovePortArgsDescription = `Use this command to remove a port from the allowlist.

Example: 'nordvpn allowlist remove port 22'

Optionally, protocol can be provided to specify which protocol should be removed from the allowlist.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn allowlist remove port 22 protocol TCP'`
)

func (c *cmd) AllowlistRemovePort(ctx *cli.Context) error {
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

	var (
		udpIndex int
		tcpIndex int
	)
	if isUDP {
		udpIndex = slices.Index(allowlist.Ports.Udp, port)
		if udpIndex >= 0 {
			allowlist.Ports.Udp = slices.Delete(allowlist.Ports.Udp, udpIndex, udpIndex+1)
		}
	}
	if isTCP {
		tcpIndex = slices.Index(allowlist.Ports.Tcp, port)
		if tcpIndex >= 0 {
			allowlist.Ports.Tcp = slices.Delete(allowlist.Ports.Tcp, tcpIndex, tcpIndex+1)
		}
	}

	if isUDP && udpIndex < 0 || isTCP && tcpIndex < 0 {
		return formatError(fmt.Errorf(
			AllowlistRemovePortExistsError,
			port,
			getProtocolStr(isTCP && tcpIndex < 0, isUDP && udpIndex < 0),
		))
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
			AllowlistRemovePortExistsError,
			port,
			getProtocolStr(isTCP, isUDP),
		))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(
			AllowlistRemovePortSuccess,
			port,
			getProtocolStr(isTCP, isUDP),
		))
	}
	return nil
}

func (c *cmd) AllowlistRemovePortAutoComplete(ctx *cli.Context) {
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
		return
	case 1:
		fmt.Println(stringProtocol)
	case 2:
		port, _ := strconv.ParseInt(ctx.Args().First(), 10, 64)
		if slices.Contains(allowlist.Ports.Udp, port) {
			fmt.Println(config.Protocol_UDP.String())
		}
		if slices.Contains(allowlist.Ports.Tcp, port) {
			fmt.Println(config.Protocol_TCP.String())
		}
	default:
		return
	}
}
