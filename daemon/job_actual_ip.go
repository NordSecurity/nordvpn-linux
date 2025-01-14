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

// tryInsightsIP is an attempt to get insights IP
// the IP it returns is always valid
func tryInsightsIP(api core.InsightsAPI) (netip.Addr, error) {
	insights, err := api.InsightsViaTunnel()
	if err == nil && insights != nil {
		ip, err := netip.ParseAddr(insights.IP)
		if err == nil {
			return ip, nil
		} else {
			return netip.Addr{}, err
		}
	} else {
		return netip.Addr{}, err
	}
}

func insightsIPUntilSuccess(ctx context.Context, api core.InsightsAPI, backoff func(int) time.Duration) (netip.Addr, error) {
	type Result struct {
		netip.Addr
		error
	}
	result := make(chan Result)
	for i := 0; ; i++ {
		if ctx.Err() != nil {
			return netip.Addr{}, ctx.Err()
		}

		// this goroutine is used so that this function is immediately stopped once the context is cancelled
		go func() {
			ip, err := tryInsightsIP(api)
			result <- Result{ip, err}
		}()

		select {
		case r := <-result:
			if r.error == nil {
				return r.Addr, nil
			} else {
				log.Println(internal.ErrorPrefix, "insights ip attempt failed: ", r.error)
			}
		case <-ctx.Done():
			return netip.Addr{}, ctx.Err()
		}

		select {
		case <-time.After(backoff(i)): // wait before retrying
		// PS: there's no fallthrough
		case <-ctx.Done():
			return netip.Addr{}, ctx.Err()
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

		insightsIP, err := insightsIPUntilSuccess(ctx, api, network.ExponentialBackoff)
		if err != nil {
			return err
		}
		if insightsIP.IsValid() {
			newIP = insightsIP
		}

		return nil
	}
}
