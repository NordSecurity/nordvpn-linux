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

	portString := args.First()
	portJSONNumber := json.Number(portString)
	port, err := strconv.Atoi(portString)
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	if !(AllowlistMinPort <= port && port <= AllowlistMaxPort) {
		return formatError(fmt.Errorf(AllowlistPortRangeError, portString, strconv.Itoa(AllowlistMinPort), strconv.Itoa(AllowlistMaxPort)))
	}

	var (
		data    = []interface{}{portString}
		success bool
		UDPSet  = mapset.NewSetFromSlice(c.config.Allowlist.Ports.UDP.ToSlice())
		TCPSet  = mapset.NewSetFromSlice(c.config.Allowlist.Ports.TCP.ToSlice())
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
		return formatError(fmt.Errorf(AllowlistRemovePortExistsError, data...))
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
		return formatError(fmt.Errorf(AllowlistRemovePortExistsError, data...))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Allowlist.Ports.UDP = UDPSet
		c.config.Allowlist.Ports.TCP = TCPSet
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(AllowlistRemovePortSuccess, data...))
	}
	return nil
}

func (c *cmd) AllowlistRemovePortAutoComplete(ctx *cli.Context) {
	switch ctx.NArg() {
	case 0:
		// create config after auth
		for port := range c.config.Allowlist.Ports.UDP.Union(c.config.Allowlist.Ports.TCP).Iter() {
			fmt.Println(port)
		}
	case 1:
		fmt.Println(stringProtocol)
	case 2:
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
