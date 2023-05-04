package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	mapset "github.com/deckarep/golang-set"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// WhitelistRemovePortsUsageText is shown next to ports command by nordvpn whitelist remove --help
const WhitelistRemovePortsUsageText = "Removes port range from a whitelist"

// WhitelistRemovePortsArgsUsageText is shown by nordvpn whitelist remove ports --help
const WhitelistRemovePortsArgsUsageText = `<port_from> <port_to> [protocol <protocol>]

Use this command to remove ports from whitelist.

Example: 'nordvpn whitelist remove ports 3000 8000'

Optionally, protocol can be provided to specify which protocol should be removed from whitelist.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn whitelist remove ports 3000 8000 protocol TCP'`

func (c *cmd) WhitelistRemovePorts(ctx *cli.Context) error {
	args := ctx.Args()

	if !(args.Len() == 2 || (args.Len() == 4 && args.Get(2) == WhitelistProtocol)) {
		return formatError(argsCountError(ctx))
	}

	startPort, err := strconv.Atoi(args.First())
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return formatError(argsParseError(ctx))
	}

	endPort, err := strconv.Atoi(args.Get(1))
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
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
		UDPSet  = mapset.NewSetFromSlice(c.config.Whitelist.Ports.UDP.ToSlice())
		TCPSet  = mapset.NewSetFromSlice(c.config.Whitelist.Ports.TCP.ToSlice())
	)
	if args.Len() == 2 {
		for port := startPort; port <= endPort; port++ {
			portJSONNumber := json.Number(strconv.Itoa(port))
			success = UDPSet.Contains(portJSONNumber) || success
			success = TCPSet.Contains(portJSONNumber) || success
			UDPSet.Remove(portJSONNumber)
			TCPSet.Remove(portJSONNumber)
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
			switch args.Get(3) {
			case config.Protocol_UDP.String():
				success = set.Contains(portJSONNumber) || success
				set.Remove(portJSONNumber)
			case config.Protocol_TCP.String():
				success = set.Contains(portJSONNumber) || success
				set.Remove(portJSONNumber)
			}
		}
	}

	if !success {
		return formatError(fmt.Errorf(WhitelistRemovePortsExistsError, data...))
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
		return formatError(fmt.Errorf(WhitelistRemovePortsExistsError, data...))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Whitelist.Ports.UDP = UDPSet
		c.config.Whitelist.Ports.TCP = TCPSet
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(WhitelistRemovePortsSuccess, data...))
	}
	return nil
}

func (c *cmd) WhitelistRemovePortsAutoComplete(ctx *cli.Context) {
	switch ctx.NArg() {
	case 0:
		ports := client.InterfacesToInt64s(c.config.Whitelist.Ports.UDP.Union(c.config.Whitelist.Ports.TCP).ToSlice())
		sort.Slice(ports, func(i, j int) bool {
			return client.InterfaceToInt64(ports[i]) < client.InterfaceToInt64(ports[j])
		})
		for _, port := range ports {
			fmt.Println(port)
		}
	case 1:
		startPort, err := strconv.ParseInt(ctx.Args().First(), 10, 64)
		if err != nil {
			return
		}
		ports := client.InterfacesToInt64s(c.config.Whitelist.Ports.UDP.Union(c.config.Whitelist.Ports.TCP).ToSlice())
		sort.Slice(ports, func(i, j int) bool {
			return client.InterfaceToInt64(ports[i]) < client.InterfaceToInt64(ports[j])
		})
		for _, port := range ports {
			if startPort <= port {
				fmt.Println(port)
			}
		}
	case 2:
		fmt.Println(stringProtocol)
	case 3:
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
