package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (c *cmd) Version(ctx *cli.Context) error {
	resp, err := c.client.Ping(context.Background(), &pb.Empty{})
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return internal.ErrSocketNotFound
		}
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "connection reset by peer") {
			return internal.ErrSocketAccessDenied
		}
		if snapErr := RetrieveSnapConnsError(err); snapErr != nil {
			return err
		}
		return internal.ErrDaemonConnectionRefused
	}

	switch resp.Type {
	case internal.CodeOffline:
		return ErrInternetConnection
	case internal.CodeDaemonOffline:
		return internal.ErrDaemonConnectionRefused
	case internal.CodeOutdated:
	case internal.CodeSuccess:
	}

	fmt.Printf("Daemon: NordVPN Version %d.%d.%d", resp.Major, resp.Minor, resp.Patch)
	if resp.Metadata != "" {
		fmt.Printf("+%s", resp.Metadata)
	}
	if resp.Type == internal.CodeOutdated {
		fmt.Print(" (outdated)")
	}
	fmt.Print("\n")

	return nil
}
