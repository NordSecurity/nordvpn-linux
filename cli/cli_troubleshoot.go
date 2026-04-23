package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

const (
	TroubleshootUsageText = "Collects diagnostic logs and system information for troubleshooting"
)

func (c *cmd) Troubleshoot(ctx *cli.Context) error {
	stream, err := c.client.CollectDiagnostics(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}

	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return formatError(err)
		}

		// Check for error in response
		if resp.Error != "" {
			return formatError(fmt.Errorf(resp.Error))
		}

		// Show progress if TTY
		if isTTY && !resp.Done {
			fmt.Printf("%s\n", resp.Step)
		}

		// Final response
		if resp.Done {
			color.Green(MsgTroubleshootSuccess, resp.FilePath)
			return nil
		}
	}

	return nil
}
