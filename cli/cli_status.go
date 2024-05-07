package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"

	"github.com/hako/durafmt"
	"github.com/urfave/cli/v2"
)

// StatusUsageText is shown next to status command by nordvpn --help
const StatusUsageText = "Shows connection status"

func (c *cmd) Status(ctx *cli.Context) error {
	resp, err := c.client.Status(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}
	fmt.Print(Status(resp))
	return nil
}

// Status returns ready to print status string.
func Status(resp *pb.StatusResponse) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Status: %s\n", resp.State))

	if resp.Name != "" {
		b.WriteString(fmt.Sprintf("Server: %s\n", resp.Name))
	}

	if resp.Hostname != "" {
		b.WriteString(fmt.Sprintf("Hostname: %s\n", resp.Hostname))
	}

	if resp.Ip != "" {
		b.WriteString(fmt.Sprintf("IP: %s\n", resp.Ip))
	}

	if resp.Country != "" {
		b.WriteString(fmt.Sprintf("Country: %s\n", resp.Country))
	}

	if resp.City != "" {
		b.WriteString(fmt.Sprintf("City: %s\n", resp.City))
	}

	if resp.Uptime != -1 {
		b.WriteString(
			fmt.Sprintf("Current technology: %s\n", resp.Technology.String()),
		)
		b.WriteString(
			fmt.Sprintf("Current protocol: %s\n", resp.Protocol.String()),
		)
	}

	// show transfer rates only if running
	if resp.Download != 0 || resp.Upload != 0 {
		b.WriteString(fmt.Sprintf(
			"Transfer: %s received, %s sent\n",
			uint64ToHumanBytes(resp.Download), uint64ToHumanBytes(resp.Upload)),
		)
	}

	if resp.Uptime != -1 {
		// truncate to skip milliseconds from being displayed
		uptime := time.Duration(resp.Uptime).Truncate(1000 * time.Millisecond)
		b.WriteString(fmt.Sprintf("Uptime: %s\n", durafmt.Parse(uptime).String()))
	}
	return b.String()
}
