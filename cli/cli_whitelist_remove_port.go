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

// WhitelistRemovePortUsageText is shown next to port command by nordvpn whitelist remove --help
const WhitelistRemovePortUsageText = "Removes port from a whitelist"

// WhitelistDeletePort is shown by nordvpn whitelist remove port --help
const WhitelistRemovePortArgsUsageText = `<port> [protocol <protocol>]

Use this command to remove port from whitelist.

Example: 'nordvpn whitelist remove port 22'

Optionally, protocol can be provided to specify which protocol should be removed from whitelist.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn whitelist remove port 22 protocol TCP'`

func (c *cmd) WhitelistRemovePort(ctx *cli.Context) error {
	args := ctx.Args()

	if !(args.Len() == 1 || (args.Len() == 3 && args.Get(1) == WhitelistProtocol)) {
		return formatError(argsCountError(ctx))
	}

	portString := args.First()
	portJSONNumber := json.Number(portString)
	port, err := strconv.Atoi(portString)
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	if !(WhitelistMinPort <= port && port <= WhitelistMaxPort) {
		return formatError(fmt.Errorf(WhitelistPortRangeError, portString, strconv.Itoa(WhitelistMinPort), strconv.Itoa(WhitelistMaxPort)))
	}

	var (
		data    = []interface{}{portString}
		success bool
		UDPSet  = mapset.NewSetFromSlice(c.config.Whitelist.Ports.UDP.ToSlice())
		TCPSet  = mapset.NewSetFromSlice(c.config.Whitelist.Ports.TCP.ToSlice())
	)
	if args.Len() == 1 {
		success = UDPSet.Contains(portJSONNumber) || success
		success = TCPSet.Contains(portJSONNumber) || success
		UDPSet.Remove(portJSONNumber)
		TCPSet.Remove(portJSONNumber)
		data = append(data, fmt.Sprintf("%s|%s", config.Protocol_UDP.String(), config.Protocol_TCP.String()))
	} else {
		switch args.Get(2) {
		case config.Protocol_UDP.String():
			success = UDPSet.Contains(portJSONNumber) || success
			UDPSet.Remove(portJSONNumber)
			data = append(data, config.Protocol_UDP.String())
		case config.Protocol_TCP.String():
			success = TCPSet.Contains(portJSONNumber) || success
			TCPSet.Remove(portJSONNumber)
			data = append(data, config.Protocol_TCP.String())
		default:
			return formatError(argsParseError(ctx))
		}
	}

	if !success {
		return formatError(fmt.Errorf(WhitelistRemovePortExistsError, data...))
	}

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
		return formatError(fmt.Errorf(WhitelistRemovePortExistsError, data...))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Whitelist.Ports.UDP = UDPSet
		c.config.Whitelist.Ports.TCP = TCPSet
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(WhitelistRemovePortSuccess, data...))
	}
	return nil
}

func (c *cmd) WhitelistRemovePortAutoComplete(ctx *cli.Context) {
	switch ctx.NArg() {
	case 0:
		// create config after auth
		for port := range c.config.Whitelist.Ports.UDP.Union(c.config.Whitelist.Ports.TCP).Iter() {
			fmt.Println(port)
		}
	case 1:
		fmt.Println(stringProtocol)
	case 2:
		if c.config.Whitelist.Ports.UDP.Contains(json.Number(ctx.Args().First())) {
			fmt.Println(config.Protocol_UDP.String())
		}
		if c.config.Whitelist.Ports.TCP.Contains(json.Number(ctx.Args().First())) {
			fmt.Println(config.Protocol_TCP.String())
		}
	default:
		return
	}
}
