package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/exp/slices"

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

	subnet := args.First()

	resp, err := c.client.UnsetAllowlist(context.Background(), &pb.SetAllowlistRequest{
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
		return formatError(fmt.Errorf(AllowlistRemoveSubnetExistsError, subnet))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeAllowlistInvalidSubnet:
		return formatError(argsParseError(ctx))
	case internal.CodeAllowlistSubnetNoop:
		return formatError(fmt.Errorf(AllowlistRemoveSubnetExistsError, subnet))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(AllowlistRemoveSubnetSuccess, subnet))
	}
	return nil
}

func (c *cmd) AllowlistRemoveSubnetAutoComplete(ctx *cli.Context) {
	settings, err := c.client.Settings(context.Background(), &pb.SettingsRequest{})
	if err != nil {
		return
	}
	allowlist := settings.GetData().Settings.GetAllowlist()
	for _, subnet := range allowlist.Subnets {
		if !slices.Contains(ctx.Args().Slice(), subnet) {
			fmt.Println(subnet)
		}
	}
}
