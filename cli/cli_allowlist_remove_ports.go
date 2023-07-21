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

// AllowlistRemovePortsUsageText is shown next to ports command by nordvpn allowlist remove --help
const AllowlistRemovePortsUsageText = "Removes port range from the allowlist"

// AllowlistRemovePortsArgsUsageText is shown by nordvpn allowlist remove ports --help
const AllowlistRemovePortsArgsUsageText = `<port_from> <port_to> [protocol <protocol>]

Use this command to remove ports from the allowlist.

Example: 'nordvpn allowlist remove ports 3000 8000'

Optionally, protocol can be provided to specify which protocol should be removed from the allowlist.
Supported values for <protocol>: TCP, UDP

Example: 'nordvpn allowlist remove ports 3000 8000 protocol TCP'`

func (c *cmd) AllowlistRemovePorts(ctx *cli.Context) error {
	args := ctx.Args()

	if !(args.Len() == 2 || (args.Len() == 4 && args.Get(2) == AllowlistProtocol)) {
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

	if !(AllowlistMinPort <= startPort && startPort <= AllowlistMaxPort && AllowlistMinPort <= endPort && endPort <= AllowlistMaxPort) {
		return formatError(fmt.Errorf(AllowlistPortsRangeError, args.First(), args.Get(1), strconv.Itoa(AllowlistMinPort), strconv.Itoa(AllowlistMaxPort)))
	}

	var (
		data    = []interface{}{args.First(), args.Get(1)}
		success bool
		UDPSet  = mapset.NewSetFromSlice(c.config.Allowlist.Ports.UDP.ToSlice())
		TCPSet  = mapset.NewSetFromSlice(c.config.Allowlist.Ports.TCP.ToSlice())
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
		return formatError(fmt.Errorf(AllowlistRemovePortsExistsError, data...))
	}

	resp, err := c.client.SetAllowlist(context.Background(), &pb.SetAllowlistRequest{
		Allowlist: &pb.Allowlist{
			Ports: &pb.Ports{
				Udp: client.SetToInt64s(UDPSet),
				Tcp: client.SetToInt64s(TCPSet),
			},
			Subnets: internal.SetToStrings(c.config.Allowlist.Subnets),
		},
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(AllowlistRemovePortsExistsError, data...))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Allowlist.Ports.UDP = UDPSet
		c.config.Allowlist.Ports.TCP = TCPSet
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(AllowlistRemovePortsSuccess, data...))
	}
	return nil
}

func (c *cmd) AllowlistRemovePortsAutoComplete(ctx *cli.Context) {
	switch ctx.NArg() {
	case 0:
		ports := client.InterfacesToInt64s(c.config.Allowlist.Ports.UDP.Union(c.config.Allowlist.Ports.TCP).ToSlice())
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
		ports := client.InterfacesToInt64s(c.config.Allowlist.Ports.UDP.Union(c.config.Allowlist.Ports.TCP).ToSlice())
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
		if c.config.Allowlist.Ports.UDP.Contains(json.Number(ctx.Args().First())) {
			fmt.Println(config.Protocol_UDP.String())
		}
		if c.config.Allowlist.Ports.TCP.Contains(json.Number(ctx.Args().First())) {
			fmt.Println(config.Protocol_TCP.String())
		}
	default:
		return
	}
}
