package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	mapset "github.com/deckarep/golang-set"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// WhitelistAddPortsUsageText is show next to ports command by nordvpn whitelist add --help
const WhitelistAddPortsUsageText = "Adds port range to a whitelist"

// WhitelistAddPortsArgsUsageText is shown by nordvpn whitelist add ports --help
const WhitelistAddPortsArgsUsageText = `[port from] [port to] [protocol [protocol]]

Use this command to whitelist the UDP and TCP ports.

Example: 'nordvpn whitelist add ports 3000 8000'

Optionally, protocol can be provided to specify which protocol should be whitelisted.
Supported values for [protocol]: TCP, UDP

Example: 'nordvpn whitelist add ports 3000 8000 protocol TCP'`

func (c *cmd) WhitelistAddPorts(ctx *cli.Context) error {
	args := ctx.Args()
	if !(args.Len() == 2 || (args.Len() == 4 && args.Get(2) == WhitelistProtocol)) {
		return formatError(argsCountError(ctx))
	}

	startPort, err := strconv.Atoi(args.First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	endPort, err := strconv.Atoi(args.Get(1))
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	if startPort > endPort {
		return formatError(argsParseError(ctx))
	}

	if !(WhitelistMinPort <= startPort && startPort <= WhitelistMaxPort && WhitelistMinPort <= endPort && endPort <= WhitelistMaxPort) {
		return formatError(fmt.Errorf(WhitelistPortsRangeError, args.First(), args.Get(1), strconv.Itoa(WhitelistMinPort), strconv.Itoa(WhitelistMaxPort)))
	}

	var (
		data    = []interface{}{args.First(), args.Get(1)}
		success bool
		UDPSet  = mapset.NewSet()
		TCPSet  = mapset.NewSet()
	)
	if args.Len() == 2 {
		for port := startPort; port <= endPort; port++ {
			portJSONNumber := json.Number(strconv.Itoa(port))
			success = UDPSet.Add(portJSONNumber) || success
			success = TCPSet.Add(portJSONNumber) || success
		}
		data = append(data, fmt.Sprintf("%s|%s", config.Protocol_UDP.String(), config.Protocol_TCP.String()))
	} else {
		var set = mapset.NewSet()
		switch args.Get(3) {
		case config.Protocol_UDP.String():
			set = UDPSet
			data = append(data, config.Protocol_UDP.String())
		case config.Protocol_TCP.String():
			set = TCPSet
			data = append(data, config.Protocol_TCP.String())
		default:
			return formatError(argsParseError(ctx))
		}

		for port := startPort; port <= endPort; port++ {
			portJSONNumber := json.Number(strconv.Itoa(port))
			success = set.Add(portJSONNumber) || success
		}
	}

	if !success {
		return formatError(fmt.Errorf(WhitelistAddPortsExistsError, data...))
	}

	UDPSet = c.config.Whitelist.Ports.UDP.Union(UDPSet)
	TCPSet = c.config.Whitelist.Ports.TCP.Union(TCPSet)
	resp, err := c.client.SetWhitelist(context.Background(), &pb.SetWhitelistRequest{
		Whitelist: &pb.Whitelist{
			Ports: &pb.Ports{
				Udp: client.SetToInt64s(UDPSet),
				Tcp: client.SetToInt64s(TCPSet),
			},
			Subnets: internal.SetToStrings(c.config.Whitelist.Subnets),
		},
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(WhitelistAddPortsExistsError, data...))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Whitelist.Ports.UDP = UDPSet
		c.config.Whitelist.Ports.TCP = TCPSet
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(WhitelistAddPortsSuccess, data...))
	}
	return nil
}

func (c *cmd) WhitelistAddPortsAutoComplete(ctx *cli.Context) {
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
