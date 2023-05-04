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

// WhitelistAddPortUsageText is shown next to port command by nordvpn whitelist add --help
const WhitelistAddPortUsageText = "Adds port to a whitelist"

// WhitelistAddPortArgsUsageText is shown on nordvpn whitelist add port --help
const WhitelistAddPortArgsUsageText = `<port> [protocol <protocol>]

Use this command to whitelist the UDP and TCP port.

Example: 'nordvpn whitelist add port 22'

Optionally, protocol can be provided to specify which protocol should be whitelisted.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn whitelist add port 22 protocol TCP'`

func (c *cmd) WhitelistAddPort(ctx *cli.Context) error {
	args := ctx.Args()
	if !(args.Len() == 1 || (args.Len() == 3 && args.Get(1) == WhitelistProtocol)) {
		return formatError(argsCountError(ctx))
	}

	portString := args.First()
	portJSONNumber := json.Number(portString)
	port, err := strconv.Atoi(args.First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	if !(WhitelistMinPort <= port && port <= WhitelistMaxPort) {
		return formatError(fmt.Errorf(WhitelistPortRangeError, portString, strconv.Itoa(WhitelistMinPort), strconv.Itoa(WhitelistMaxPort)))
	}

	var (
		data    = []interface{}{portString}
		success bool
		UDPSet  = mapset.NewSet()
		TCPSet  = mapset.NewSet()
	)
	if args.Len() == 1 {
		success = UDPSet.Add(portJSONNumber) || success
		success = TCPSet.Add(portJSONNumber) || success
		data = append(data, fmt.Sprintf("%s|%s", config.Protocol_UDP, config.Protocol_TCP))
	} else {
		switch args.Get(2) {
		case config.Protocol_UDP.String():
			success = UDPSet.Add(portJSONNumber) || success
			data = append(data, config.Protocol_UDP)
		case config.Protocol_TCP.String():
			success = TCPSet.Add(portJSONNumber) || success
			data = append(data, config.Protocol_TCP)
		default:
			return formatError(argsParseError(ctx))
		}
	}

	if !success {
		return formatError(fmt.Errorf(WhitelistAddPortExistsError, data...))
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
		return formatError(fmt.Errorf(WhitelistAddPortExistsError, data...))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Whitelist.Ports.UDP = UDPSet
		c.config.Whitelist.Ports.TCP = TCPSet
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(WhitelistAddPortSuccess, data...))
	}
	return nil
}

func (c *cmd) WhitelistAddPortAutoComplete(ctx *cli.Context) {
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
