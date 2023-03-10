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

// WhitelistRemoveSubnetUsageText text is shown next to subnet command by nordvpn whitelist remove --help
const WhitelistRemoveSubnetUsageText = "Removes subnet from a whitelist"

// WhitelistRemoveSubnetArgsUsageText is shown by nordvpn whitelist remove subnet --help
const WhitelistRemoveSubnetArgsUsageText = `[address]

Use this command to remove subnet from whitelist.

Example: 'nordvpn whitelist remove subnet 192.168.1.1/24'`

func (c *cmd) WhitelistRemoveSubnet(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsCountError(ctx))
	}

	_, subnet, err := net.ParseCIDR(args.First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	if !c.config.Whitelist.Subnets.Contains(subnet.String()) {
		return formatError(fmt.Errorf(WhitelistRemoveSubnetExistsError, subnet.String()))
	}

	subnets := mapset.NewSetFromSlice(c.config.Whitelist.Subnets.ToSlice())
	subnets.Remove(subnet.String())

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
		return formatError(fmt.Errorf(WhitelistRemoveSubnetExistsError, subnet))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Whitelist.Subnets = subnets
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(WhitelistRemoveSubnetSuccess, subnet))
	}
	return nil
}

func (c *cmd) WhitelistRemoveSubnetAutoComplete(ctx *cli.Context) {
	subnets := internal.SetToStrings(c.config.Whitelist.Subnets)
	for _, subnet := range subnets {
		if !internal.StringsContains(ctx.Args().Slice(), subnet) {
			fmt.Println(subnet)
		}
	}
}
