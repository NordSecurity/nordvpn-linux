package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

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
  Address should be in IPv4 CIDR notation`
)

func (c *cmd) AllowlistAddSubnet(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsCountError(ctx))
	}

	subnet := args.First()

	// first call to rpc.SetAllowList() has Force=false which means do not remove narrower subnet, if found
	resp, err := c.client.SetAllowlist(context.Background(), &pb.SetAllowlistRequest{
		Request: &pb.SetAllowlistRequest_SetAllowlistSubnetRequest{
			SetAllowlistSubnetRequest: &pb.SetAllowlistSubnetRequest{Subnet: subnet, Force: false},
		},
	})

	if err != nil {
		return formatError(err)
	}

	if resp != nil && resp.Type == internal.CodeAllowlistSubnetWider {
		// ask user to confirm removal of narrower subnet when wider subnet is added to allowlist
		if !readForConfirmationDefaultValue(os.Stdin, MsgRemoveNarrowConfirmPrompt, true) {
			return nil
		}
		// second call to rpc.SetAllowList() has Force=true which means force remove narrower subnet from allowlist
		resp, err = c.client.SetAllowlist(context.Background(), &pb.SetAllowlistRequest{
			Request: &pb.SetAllowlistRequest_SetAllowlistSubnetRequest{
				SetAllowlistSubnetRequest: &pb.SetAllowlistSubnetRequest{Subnet: subnet, Force: true},
			},
		})
		if err != nil {
			return formatError(err)
		}
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(AllowlistAddSubnetExistsError, subnet))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodePrivateSubnetLANDiscovery:
		return formatError(errors.New(AllowlistAddSubnetLANDiscovery))
	case internal.CodeAllowlistInvalidSubnet:
		return formatError(argsParseError(ctx))
	case internal.CodeAllowlistSubnetNoop:
		return formatError(fmt.Errorf(AllowlistAddSubnetExistsError, subnet))
	case internal.CodeAllowlistSubnetSmallerNoop:
		return formatError(fmt.Errorf(AllowlistAddSubnetExistsError, subnet))
	case internal.CodeAllowlistSubnetTooWideWarn:
		color.Yellow(AllowlistAddSubnetTooWideWarning) // show warning in yellow, and then show following success msg in green
		fallthrough
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(AllowlistAddSubnetSuccess, subnet))
	}
	return nil
}

func (c *cmd) AllowlistAddSubnetAutoComplete(ctx *cli.Context) {}
