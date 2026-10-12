package daemon

import (
	"context"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/network"
)

func JobPurchases(
	api core.SubscriptionAPI,
	purchasesPublisher events.Publisher[[]core.Purchase],
) func(context.Context) {
	return func(ctx context.Context) {
		for i := 0; i < 5; i++ {
			result, err := api.Purchases()
			if err != nil {
				log.Error("purchases job: request failed:", err)
				if !sleep(ctx, network.ExponentialBackoff(i)) {
					return
				}
				continue
			}

			switch result.Status {
			case core.PurchaseStatusReady:
				log.Info("purchases job: data fetched successfully")
				purchasesPublisher.Publish(result.Data)
				return

			case core.PurchaseStatusPending:
				log.Info("purchases job: pending, retrying in", result.RetryAfter)
				if !sleep(ctx, result.RetryAfter) {
					return
				}
				continue

			case core.PurchaseStatusUnknown:
			default:
				log.Error("purchases job: unexpected status:", result.Status)
				return
			}
		}

		log.Error("purchases job: not able to fetch the purchases after 5 tries, giving up")
	}
}

// Returns true if wait completed, false if canceled
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}
