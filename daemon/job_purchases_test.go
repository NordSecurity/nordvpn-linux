package daemon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	core_test "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/stretchr/testify/assert"
)

// spyPublisher records everything the job publishes in order
type spyPublisher struct {
	published [][]core.Purchase
}

func (s *spyPublisher) Publish(data []core.Purchase) {
	s.published = append(s.published, data)
}

// spySleep records each sleep duration and controls the sleep return value
type spySleep struct {
	durations []time.Duration
	completed bool         // default for every call
	returns   map[int]bool // optional per-call index override
}

func (s *spySleep) fn(_ context.Context, d time.Duration) bool {
	idx := len(s.durations)
	s.durations = append(s.durations, d)
	if s.returns != nil {
		if v, ok := s.returns[idx]; ok {
			return v
		}
	}
	return s.completed
}

type resp = struct {
	result core.PurchasesResult
	err    error
}

var errRandom = errors.New("random error")

func fixedBackoff(i int) time.Duration {
	return time.Duration(i+1) * 10 * time.Millisecond
}

func dummyPurchases() []core.Purchase {
	return []core.Purchase{
		{
			Order: core.PurchaseOrder{
				ID: 269195093, RemoteID: 200, Status: "active", Source: "nordvpn",
			},
			Plans: []core.PurchasePlan{
				{
					ID:         269195093,
					Identifier: "nordsec_ultra_bundle_android_1_year_299",
					Type:       "android",
					Services:   []core.PurchaseService{{Merchant: "nordvpn", Identifier: "vpn"}},
				},
			},
			Payments: []core.PurchasePayment{
				{
					ID: 269195093, CreatedAt: "2017-11-09 18:45:02", Status: "done",
					Amount: "100.00", Currency: "USD", Provider: "stripe", SubscriptionID: 269195093,
				},
			},
			Subscriptions: []core.PurchaseSubscription{
				{
					ID: 269195093, FrequencyInterval: 1, FrequencyUnit: "year",
					MerchantID: "1", Status: "active",
				},
			},
		},
	}
}

func TestJobPurchases(t *testing.T) {
	data := dummyPurchases()

	ready := func() resp {
		return resp{result: core.PurchasesResult{Status: core.PurchaseStatusReady, Data: data}}
	}
	pendingAfter := func(d time.Duration) resp {
		return resp{result: core.PurchasesResult{Status: core.PurchaseStatusPending, RetryAfter: d}}
	}
	unknown := resp{result: core.PurchasesResult{Status: core.PurchaseStatusUnknown}}
	failure := resp{err: errRandom}

	tests := []struct {
		name string

		responses []resp
		sleepRet  bool
		sleepPer  map[int]bool

		wantCalls     int
		wantPublished [][]core.Purchase
		wantSleeps    []time.Duration
	}{
		{
			name:          "ready on first try publishes and never sleeps",
			responses:     []resp{ready()},
			sleepRet:      true,
			wantCalls:     1,
			wantPublished: [][]core.Purchase{data},
			wantSleeps:    nil,
		},
		{
			name:          "pending then ready honors RetryAfter",
			responses:     []resp{pendingAfter(30 * time.Second), ready()},
			sleepRet:      true,
			wantCalls:     2,
			wantPublished: [][]core.Purchase{data},
			wantSleeps:    []time.Duration{30 * time.Second},
		},
		{
			name: "multiple pending values each honored in order",
			responses: []resp{
				pendingAfter(10 * time.Second),
				pendingAfter(45 * time.Second),
				ready(),
			},
			sleepRet:      true,
			wantCalls:     3,
			wantPublished: [][]core.Purchase{data},
			wantSleeps:    []time.Duration{10 * time.Second, 45 * time.Second},
		},
		{
			name:          "error then ready uses backoff then publishes",
			responses:     []resp{failure, failure, ready()},
			sleepRet:      true,
			wantCalls:     3,
			wantPublished: [][]core.Purchase{data},
			wantSleeps:    []time.Duration{fixedBackoff(0), fixedBackoff(1)},
		},
		{
			name:          "always pending gives up after 5 tries",
			responses:     []resp{pendingAfter(5 * time.Second)},
			sleepRet:      true,
			wantCalls:     5,
			wantPublished: nil,
			wantSleeps: []time.Duration{
				5 * time.Second, 5 * time.Second, 5 * time.Second, 5 * time.Second, 5 * time.Second,
			},
		},
		{
			name:          "always error gives up after 5 tries using backoff",
			responses:     []resp{failure},
			sleepRet:      true,
			wantCalls:     5,
			wantPublished: nil,
			wantSleeps: []time.Duration{
				fixedBackoff(0), fixedBackoff(1), fixedBackoff(2), fixedBackoff(3), fixedBackoff(4),
			},
		},
		{
			name:          "unknown status bails immediately",
			responses:     []resp{unknown},
			sleepRet:      true,
			wantCalls:     1,
			wantPublished: nil,
			wantSleeps:    nil,
		},
		{
			name:          "cancelled during error backoff stops after first attempt",
			responses:     []resp{failure},
			sleepRet:      false,
			wantCalls:     1,
			wantPublished: nil,
			wantSleeps:    []time.Duration{fixedBackoff(0)},
		},
		{
			name:          "cancelled during pending wait stops after first attempt",
			responses:     []resp{pendingAfter(60 * time.Second)},
			sleepRet:      false,
			wantCalls:     1,
			wantPublished: nil,
			wantSleeps:    []time.Duration{60 * time.Second},
		},
		{
			name:          "cancelled on second wait stops mid-retry",
			responses:     []resp{pendingAfter(1 * time.Second)},
			sleepPer:      map[int]bool{0: true, 1: false},
			wantCalls:     2,
			wantPublished: nil,
			wantSleeps:    []time.Duration{1 * time.Second, 1 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			api := core_test.NewMockSubscriptionAPI()
			api.PurchasesFunc = func() (core.PurchasesResult, error) {
				i := calls
				calls++
				if i >= len(tt.responses) {
					i = len(tt.responses) - 1
				}
				return tt.responses[i].result, tt.responses[i].err
			}

			pub := &spyPublisher{}
			slp := &spySleep{completed: tt.sleepRet, returns: tt.sleepPer}

			JobPurchases(api, pub, fixedBackoff, slp.fn)(context.Background())

			assert.Equal(t, tt.wantCalls, calls, "api call count")
			assert.Equal(t, tt.wantPublished, pub.published, "published data")
			assert.Equal(t, tt.wantSleeps, slp.durations, "sleep durations")
		})
	}
}
