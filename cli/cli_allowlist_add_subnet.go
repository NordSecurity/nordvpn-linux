package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

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

	subnet := args.First()

	resp, err := c.client.SetAllowlist(context.Background(), &pb.SetAllowlistRequest{
		Request: &pb.SetAllowlistRequest_SetAllowlistSubnetRequest{
			SetAllowlistSubnetRequest: &pb.SetAllowlistSubnetRequest{Subnet: subnet},
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
	case internal.CodeAllowlistInvalidSubnet:
		return formatError(argsParseError(ctx))
	case internal.CodeAllowlistSubnetNoop:
		return formatError(fmt.Errorf(AllowlistAddSubnetExistsError, subnet))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(AllowlistAddSubnetSuccess, subnet))
	}
	return nil
}

func (c *cmd) AllowlistAddSubnetAutoComplete(ctx *cli.Context) {}
