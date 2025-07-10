package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/urfave/cli/v2"
)

func (c *cmd) Version(ctx *cli.Context) error {
	outdated := ""
	resp, err := c.client.Ping(context.Background(), &pb.Empty{})
	if err == nil && resp.Type == internal.CodeOutdated {
		outdated = " (outdated)"
	}

	fmt.Printf("NordVPN Version %s%s\n", ctx.App.Version, outdated)

	return nil
}
