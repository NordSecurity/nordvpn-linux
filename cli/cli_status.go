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

func formatDuration(secs uint32) string {
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%02d:%02d", m, s)
	}
	return fmt.Sprintf("%02d", s)
}

// Status returns ready to print status string.
func Status(resp *pb.StatusResponse) string {
	state := "Disconnected"
	//exhaustive:ignore
	switch resp.State {
	case pb.ConnectionState_CONNECTED:
		state = "Connected"
	case pb.ConnectionState_CONNECTING:
		state = "Connecting"
	case pb.ConnectionState_PAUSED:
		state = "Paused"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Status: %s\n", state))

	if resp.PauseRemainingDurationSec != 0 {
		duration := formatDuration(resp.PauseRemainingDurationSec)
		b.WriteString(fmt.Sprintf("Pause time left: %s\n", duration))
	}

	if resp.Name != "" {
		serverName := resp.Name
		if resp.VirtualLocation {
			serverName += " - Virtual"
		}
		b.WriteString(fmt.Sprintf("Server: %s\n", serverName))
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
		pqLabel := "Disabled"
		if resp.PostQuantum {
			pqLabel = "Enabled"
		}
		b.WriteString(
			fmt.Sprintf("Post-quantum VPN: %s\n", pqLabel),
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
