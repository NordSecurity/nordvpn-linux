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

// Allowlist add subnet help text
const (
	AllowlistAddSubnetUsageText     = "Adds subnet to the allowlist"
	AllowlistAddSubnetArgsUsageText = `<address>`
	AllowlistAddSubnetDescription   = `Use this command to allowlist subnet.

Example: 'nordvpn allowlist add subnet 192.168.1.1/24'

Notes:
  Address should be in CIDR notation`
)

func (c *cmd) AllowlistAddSubnet(ctx *cli.Context) error {
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
		return formatError(fmt.Errorf(AllowlistAddSubnetExistsError, subnet.String()))
	}

	subnets = c.config.Allowlist.Subnets.Union(subnets)
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
		return formatError(fmt.Errorf(AllowlistAddSubnetExistsError, subnet))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodePrivateSubnetLANDiscovery:
		return formatError(fmt.Errorf(AllowlistAddSubnetLANDiscovery))
	case internal.CodeSuccess:
		c.config.Allowlist.Subnets = subnets
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(AllowlistAddSubnetSuccess, subnet))
	}
	return nil
}

func (c *cmd) AllowlistAddSubnetAutoComplete(ctx *cli.Context) {}
