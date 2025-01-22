package daemon

import (
	"context"
	"log"
	"net/netip"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/events"
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
			if ctx.Err() == nil {
				result <- Result{ip, err}
			}
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

func updateActualIP(dm *DataManager, api core.InsightsAPI, ctx context.Context, isConnected bool) error {
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

// ActualIPResolver is a long-running function that will update the actual IP address indefinitely
// it reacts to state updates from the statePublisher
func ActualIPResolver(statePublisher *state.StatePublisher, dm *DataManager, api core.InsightsAPI) {
	call := func(ctx context.Context, isConnected bool) {
		err := updateActualIP(dm, api, ctx, isConnected)
		if err == nil {
			err := statePublisher.NotifyActualIPUpdate()
			if err != nil {
				log.Println(internal.ErrorPrefix, "notify about actual ip update failed: ", err)
			}
		} else {
			if err == context.Canceled {
				return
			}
			log.Println(internal.ErrorPrefix, "actual ip job error: ", err)
		}
	}

	stateChan, _ := statePublisher.AddSubscriber()
	var cancel context.CancelFunc

	for ev := range stateChan {
		_, isConnect := ev.(events.DataConnect)
		_, isDisconnect := ev.(events.DataDisconnect)

		if isConnect || isDisconnect {
			if cancel != nil {
				cancel()
			}

			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())

			if isConnect {
				go call(ctx, true)
			} else {
				call(ctx, false) // should finish immediately, that's why it's not a separate goroutine
			}
		}
	}

	// Ensure the context is canceled when the loop exits
	if cancel != nil {
		cancel()
	}
}
