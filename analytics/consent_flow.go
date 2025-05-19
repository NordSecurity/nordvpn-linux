package analytics

import (
	"context"
	"os"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/fatih/color"
)

func StartConsentFlow(client pb.DaemonClient) {
	// XXX: Implement proper consent flow
	color.Green("Starting consent flow")
	if _, err := client.SetAnalytics(
		context.Background(),
		&pb.SetGenericRequest{Enabled: true},
	); err != nil {
		// XXX: Improve this
		color.Red("Error when setting analytics:", err)
	}
	time.Sleep(time.Second * 10)
	os.Exit(0)
}
