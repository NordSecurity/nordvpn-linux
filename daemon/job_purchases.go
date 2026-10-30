package daemon

import (
	"context"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const maxRetryAttempts int = 10

func JobPurchases(
	api core.SubscriptionAPI,
	purchasesPublisher events.Publisher[[]core.Purchase],
	backoff func(int) time.Duration,
	sleep func(context.Context, time.Duration) bool,
) func(context.Context) {
	return func(ctx context.Context) {
		for i := 0; i < maxRetryAttempts; i++ {
			result, err := api.Purchases()
			if err != nil {
				log.Error("purchases job: request failed:", err)
				if shouldGiveUp(i) {
					return
				}
				if !sleep(ctx, backoff(i)) {
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
				if shouldGiveUp(i) {
					return
				}
				if !sleep(ctx, result.RetryAfter) {
					return
				}
				continue

			case core.PurchaseStatusUnknown:
				fallthrough
			default:
				log.Error("purchases job: unexpected status:", result.Status)
				return
			}
		}

		log.Error("purchases job: not able to fetch the purchases after", maxRetryAttempts, "tries, giving up")
	}
}

// PurchasesJobSleepFunction returns true if wait completed, false if canceled
func PurchasesJobSleepFunction(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// shouldGiveUp returns true if we are at the final try
func shouldGiveUp(i int) bool {
	return i >= (maxRetryAttempts - 1)
}

type PurchasesRunner struct {
	baseCtx context.Context
	job     func(context.Context)
	ac      auth.Checker

	mu      sync.Mutex
	running bool
	pending bool
	cancel  context.CancelFunc
}

func NewPurchasesRunner(baseCtx context.Context, job func(context.Context), ac auth.Checker) *PurchasesRunner {
	return &PurchasesRunner{baseCtx: baseCtx, job: job, ac: ac}
}

func (r *PurchasesRunner) Run(ctx context.Context) {
	r.mu.Lock()
	if r.running {
		r.pending = true
		r.mu.Unlock()
		return
	}
	r.running = true
	r.mu.Unlock()

	for {
		r.runOnce(ctx)

		r.mu.Lock()
		if !r.pending {
			r.running = false
			r.mu.Unlock()
			return
		}
		r.pending = false
		r.mu.Unlock()
	}
}

func (r *PurchasesRunner) runOnce(parent context.Context) {
	if ok, _ := r.ac.IsLoggedIn(); !ok {
		log.Info("purchases job: not logged in, skipping")
		return
	}

	ctx, cancel := context.WithCancel(parent)
	r.mu.Lock()
	r.cancel = cancel
	r.mu.Unlock()

	defer func() {
		cancel()
		r.mu.Lock()
		r.cancel = nil
		r.mu.Unlock()
	}()

	r.job(ctx)
}

func (r *PurchasesRunner) NotifyLogin(e events.DataAuthorization) error {
	if e.EventStatus != events.StatusSuccess {
		return nil
	}

	r.mu.Lock()
	if r.cancel != nil {
		r.cancel()
	}
	r.mu.Unlock()

	go r.Run(r.baseCtx)
	return nil
}

func (r *PurchasesRunner) NotifyLogout(e events.DataAuthorization) error {
	if e.EventStatus != events.StatusSuccess {
		return nil
	}

	r.mu.Lock()
	if r.cancel != nil {
		r.cancel()
	}
	r.mu.Unlock()
	return nil
}

func (r *PurchasesRunner) NotifyMFA(bool) error { return nil }
