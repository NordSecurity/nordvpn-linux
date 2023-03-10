package cli

import (
	"context"
	"fmt"
	"net"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	mapset "github.com/deckarep/golang-set"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// WhitelistAddSubnetUsageText is shown next to subnet command by nordvpn whitelist add --help
const WhitelistAddSubnetUsageText = "Adds subnet to a whitelist"

// WhitelistAddSubnetArgsUsageText is shown by nordvpn whitelist add subnet --help
const WhitelistAddSubnetArgsUsageText = `[address]

Use this command to whitelist subnet.

Example: 'nordvpn whitelist add subnet 192.168.1.1/24'

Notes:
  Address should be in CIDR notation`

func (c *cmd) WhitelistAddSubnet(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsCountError(ctx))
	}

	_, subnet, err := net.ParseCIDR(args.First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	var subnets = mapset.NewSet()
	if !subnets.Add(subnet.String()) {
		return formatError(fmt.Errorf(WhitelistAddSubnetExistsError, subnet.String()))
	}

	subnets = c.config.Whitelist.Subnets.Union(subnets)
	resp, err := c.client.SetWhitelist(context.Background(), &pb.SetWhitelistRequest{
		Whitelist: &pb.Whitelist{
			Ports: &pb.Ports{
				Udp: client.SetToInt64s(c.config.Whitelist.Ports.UDP),
				Tcp: client.SetToInt64s(c.config.Whitelist.Ports.TCP),
			},
			Subnets: internal.SetToStrings(subnets),
		},
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(WhitelistAddSubnetExistsError, subnet))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Whitelist.Subnets = subnets
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(WhitelistAddSubnetSuccess, subnet))
	}
	return nil
}

func (c *cmd) WhitelistAddSubnetAutoComplete(ctx *cli.Context) {}
