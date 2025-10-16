package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// AllowlistRemoveAllUsageText is shown next to all command by nordvpn allowlist remove --help
const AllowlistRemoveAllUsageText = "Removes all ports and subnets from the allowlist"

func (c *cmd) AllowlistRemoveAll(ctx *cli.Context) error {
	resp, err := c.client.UnsetAllAllowlist(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure:
		return formatError(fmt.Errorf(AllowlistRemoveAllError))
	case internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(AllowlistRemoveAllSuccess))
	}

	return nil
}
