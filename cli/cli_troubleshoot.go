package cli

import (
	"context"
	"errors"
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
			return formatError(errors.New(resp.Error))
		}

		// Final response: daemon signals completion by sending the file
		// path with no error.
		if resp.FilePath != "" {
			color.Green(MsgTroubleshootSuccess, resp.FilePath)
			return nil
		}

		// Show progress if TTY
		if isTTY {
			fmt.Printf("%s\n", resp.Step)
		}
	}

	return nil
}
