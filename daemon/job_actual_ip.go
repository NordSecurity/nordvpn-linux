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

func insightsIPUntilSuccess(ctx context.Context, api core.InsightsAPI) (netip.Addr, error) {
	for i := 0; ; i++ {
		if ctx.Err() != nil {
			return netip.Addr{}, ctx.Err()
		}

		insights, err := api.InsightsViaTunnel()
		if err == nil && insights != nil {
			ip, err := netip.ParseAddr(insights.IP)
			if err != nil {
				return netip.Addr{}, err
			}

			return ip, nil
		} else {
			log.Println(internal.ErrorPrefix, err)
		}

		backoff := network.ExponentialBackoff(i)

		// Wait before retrying
		select {
		case <-time.After(backoff):
			// Continue to the next retry
		case <-ctx.Done():
			return netip.Addr{}, ctx.Err() // Exit if context is canceled during sleep
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

		insightsIP, err := insightsIPUntilSuccess(ctx, api)
		if err != nil {
			return err
		}
		if insightsIP.IsValid() {
			newIP = insightsIP
		}

		return nil
	}
}
