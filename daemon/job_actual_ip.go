package daemon

import (
	"context"
	"log"
	"net/netip"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/network"
)

func insightsUntilSuccess(ctx context.Context, api core.InsightsAPI) (core.Insights, error) {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return core.Insights{}, ctx.Err()
		default:
			insights, err := api.Insights()
			if err == nil || insights != nil {
				return *insights, nil
			} else {
				log.Println(internal.ErrorPrefix, err)
			}

			backoff := network.ExponentialBackoff(i)

			// Wait before retrying
			select {
			case <-time.After(backoff):
				// Continue to the next retry
			case <-ctx.Done():
				return core.Insights{}, ctx.Err() // Exit if context is canceled during sleep
			}
		}
	}

}

func JobActualIP(dm *DataManager, api core.InsightsAPI) func(context.Context, bool) error {
	return func(ctx context.Context, isConnected bool) error {
		var newIP netip.Addr
		defer func() {
			dm.SetActualIP(newIP)
		}()

		if !isConnected {
			return nil
		}

		insights, err := insightsUntilSuccess(ctx, api)
		if err != nil {
			return err
		}

		insightsIP, err := netip.ParseAddr(insights.IP)
		if err != nil {
			return err
		}
		if insightsIP.IsValid() {
			newIP = insightsIP
		}

		return nil
	}
}
