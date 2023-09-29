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

// Allowlist remove subnet help text
const (
	AllowlistRemoveSubnetUsageText       = "Removes subnet from the allowlist"
	AllowlistRemoveSubnetArgsUsageText   = `<address>`
	AllowlistRemoveSubnetArgsDescription = `Use this command to remove subnet from the allowlist.

Example: 'nordvpn allowlist remove subnet 192.168.1.1/24'`
)

func (c *cmd) AllowlistRemoveSubnet(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsCountError(ctx))
	}

	_, subnet, err := net.ParseCIDR(args.First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	if !c.config.Allowlist.Subnets.Contains(subnet.String()) {
		return formatError(fmt.Errorf(AllowlistRemoveSubnetExistsError, subnet.String()))
	}

	subnets := mapset.NewSetFromSlice(c.config.Allowlist.Subnets.ToSlice())
	subnets.Remove(subnet.String())

	resp, err := c.client.SetAllowlist(context.Background(), &pb.SetAllowlistRequest{
		Allowlist: &pb.Allowlist{
			Ports: &pb.Ports{
				Udp: client.SetToInt64s(c.config.Allowlist.Ports.UDP),
				Tcp: client.SetToInt64s(c.config.Allowlist.Ports.TCP),
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
		return formatError(fmt.Errorf(AllowlistRemoveSubnetExistsError, subnet))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.Allowlist.Subnets = subnets
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(AllowlistRemoveSubnetSuccess, subnet))
	}
	return nil
}

func (c *cmd) AllowlistRemoveSubnetAutoComplete(ctx *cli.Context) {
	subnets := internal.SetToStrings(c.config.Allowlist.Subnets)
	for _, subnet := range subnets {
		if !internal.StringsContains(ctx.Args().Slice(), subnet) {
			fmt.Println(subnet)
		}
	}
}
