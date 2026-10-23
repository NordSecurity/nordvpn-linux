package daemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	core_test "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/stretchr/testify/assert"
)

var errRandom = errors.New("random error")

const (
	testTimeout = time.Second
	quietPeriod = 50 * time.Millisecond
)

func fixedBackoff(i int) time.Duration {
	return time.Duration(i+1) * 10 * time.Millisecond
}

func backoffs(to int) []time.Duration {
	var out []time.Duration
	for i := 0; i < to; i++ {
		out = append(out, fixedBackoff(i))
	}
	return out
}

func repeatDur(d time.Duration, n int) []time.Duration {
	out := make([]time.Duration, n)
	for i := range out {
		out[i] = d
	}
	return out
}

func trackMax(max *atomic.Int32, n int32) {
	for {
		m := max.Load()
		if n <= m || max.CompareAndSwap(m, n) {
			return
		}
	}
}

type spyPublisher struct {
	published [][]core.Purchase
}

func (s *spyPublisher) Publish(data []core.Purchase) {
	s.published = append(s.published, data)
}

// spySleep records each sleep call and controls the sleep return value
type spySleep struct {
	durations []time.Duration
	ctxs      []context.Context
	completed bool
	returns   map[int]bool
}

func (s *spySleep) fn(ctx context.Context, d time.Duration) bool {
	idx := len(s.durations)
	s.durations = append(s.durations, d)
	s.ctxs = append(s.ctxs, ctx)
	if v, ok := s.returns[idx]; ok {
		return v
	}
	return s.completed
}

type resp = struct {
	result core.PurchasesResult
	err    error
}

func repeatResp(r resp, n int) []resp {
	out := make([]resp, n)
	for i := range out {
		out[i] = r
	}
	return out
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

func newMockAPI(responses []resp, calls *int) *core_test.MockSubscriptionAPI {
	api := core_test.NewMockSubscriptionAPI()
	api.PurchasesFunc = func() (core.PurchasesResult, error) {
		i := *calls
		*calls++
		if i >= len(responses) {
			i = len(responses) - 1
		}
		return responses[i].result, responses[i].err
	}
	return api
}

func TestJobPurchases(t *testing.T) {
	category.Set(t, category.Unit)

	data := dummyPurchases()
	last := maxRetryAttempts - 1 // number of sleeps when every attempt is used

	ready := resp{result: core.PurchasesResult{Status: core.PurchaseStatusReady, Data: data}}
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
			responses:     []resp{ready},
			sleepRet:      true,
			wantCalls:     1,
			wantPublished: [][]core.Purchase{data},
		},
		{
			name:          "pending then ready honors RetryAfter",
			responses:     []resp{pendingAfter(30 * time.Second), ready},
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
				ready,
			},
			sleepRet:      true,
			wantCalls:     3,
			wantPublished: [][]core.Purchase{data},
			wantSleeps:    []time.Duration{10 * time.Second, 45 * time.Second},
		},
		{
			name:          "error then ready uses backoff then publishes",
			responses:     []resp{failure, failure, ready},
			sleepRet:      true,
			wantCalls:     3,
			wantPublished: [][]core.Purchase{data},
			wantSleeps:    backoffs(2),
		},
		{
			name:          "mixed errors and pending use attempt index for backoff",
			responses:     []resp{failure, pendingAfter(7 * time.Second), failure, ready},
			sleepRet:      true,
			wantCalls:     4,
			wantPublished: [][]core.Purchase{data},
			wantSleeps:    []time.Duration{fixedBackoff(0), 7 * time.Second, fixedBackoff(2)},
		},
		{
			name:          "ready on the last allowed attempt still publishes",
			responses:     append(repeatResp(failure, last), ready),
			sleepRet:      true,
			wantCalls:     maxRetryAttempts,
			wantPublished: [][]core.Purchase{data},
			wantSleeps:    backoffs(last),
		},
		{
			name:       "always pending gives up without sleeping after the last try",
			responses:  []resp{pendingAfter(5 * time.Second)},
			sleepRet:   true,
			wantCalls:  maxRetryAttempts,
			wantSleeps: repeatDur(5*time.Second, last),
		},
		{
			name:       "always error gives up without sleeping after the last try",
			responses:  []resp{failure},
			sleepRet:   true,
			wantCalls:  maxRetryAttempts,
			wantSleeps: backoffs(last),
		},
		{
			name:      "unknown status bails immediately",
			responses: []resp{unknown},
			sleepRet:  true,
			wantCalls: 1,
		},
		{
			name:       "unknown status after pending bails",
			responses:  []resp{pendingAfter(1 * time.Second), unknown},
			sleepRet:   true,
			wantCalls:  2,
			wantSleeps: []time.Duration{1 * time.Second},
		},
		{
			name:       "cancelled during error backoff stops after first attempt",
			responses:  []resp{failure},
			sleepRet:   false,
			wantCalls:  1,
			wantSleeps: backoffs(1),
		},
		{
			name:       "cancelled during pending wait stops after first attempt",
			responses:  []resp{pendingAfter(60 * time.Second)},
			sleepRet:   false,
			wantCalls:  1,
			wantSleeps: []time.Duration{60 * time.Second},
		},
		{
			name:       "cancelled on second wait stops mid-retry",
			responses:  []resp{pendingAfter(1 * time.Second)},
			sleepPer:   map[int]bool{0: true, 1: false},
			wantCalls:  2,
			wantSleeps: []time.Duration{1 * time.Second, 1 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			api := newMockAPI(tt.responses, &calls)
			pub := &spyPublisher{}
			slp := &spySleep{completed: tt.sleepRet, returns: tt.sleepPer}

			JobPurchases(api, pub, fixedBackoff, slp.fn)(context.Background())

			assert.Equal(t, tt.wantCalls, calls, "api call count")
			assert.Equal(t, tt.wantPublished, pub.published, "published data")
			assert.Equal(t, tt.wantSleeps, slp.durations, "sleep durations")
		})
	}
}

func TestJobPurchases_PassesContextToSleep(t *testing.T) {
	category.Set(t, category.Unit)

	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "job-ctx")

	calls := 0
	api := newMockAPI([]resp{{err: errRandom}, {result: core.PurchasesResult{
		Status: core.PurchaseStatusPending, RetryAfter: time.Second,
	}}}, &calls)
	slp := &spySleep{returns: map[int]bool{0: true, 1: false}}

	JobPurchases(api, &spyPublisher{}, fixedBackoff, slp.fn)(ctx)

	assert.Len(t, slp.ctxs, 2)
	for i, got := range slp.ctxs {
		assert.Equal(t, "job-ctx", got.Value(key{}), "sleep call %d got a different context", i)
	}
}

func TestPurchasesJobSleepFunction(t *testing.T) {
	category.Set(t, category.Unit)

	t.Run("returns true when the wait completes", func(t *testing.T) {
		assert.True(t, PurchasesJobSleepFunction(context.Background(), time.Millisecond))
	})

	t.Run("returns false when the context is already cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.False(t, PurchasesJobSleepFunction(ctx, time.Hour))
	})

	t.Run("returns false when cancelled during the wait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(10*time.Millisecond, cancel)

		start := time.Now()
		assert.False(t, PurchasesJobSleepFunction(ctx, time.Hour))
		assert.Less(t, time.Since(start), testTimeout)
	})
}

type fakeChecker struct {
	auth.Checker
	loggedIn atomic.Bool
}

func (c *fakeChecker) IsLoggedIn() (bool, error) { return c.loggedIn.Load(), nil }

type blockingJob struct {
	started   chan context.Context
	release   chan struct{}
	calls     atomic.Int32
	active    atomic.Int32
	maxActive atomic.Int32
}

func (j *blockingJob) run(ctx context.Context) {
	j.calls.Add(1)
	trackMax(&j.maxActive, j.active.Add(1))
	defer j.active.Add(-1)

	j.started <- ctx
	select {
	case <-j.release:
	case <-ctx.Done():
	}
}

var (
	authSuccess = events.DataAuthorization{EventStatus: events.StatusSuccess}
	authFailure = events.DataAuthorization{EventStatus: events.StatusFailure}
)

type runnerHarness struct {
	t    *testing.T
	r    *PurchasesRunner
	job  *blockingJob
	chk  *fakeChecker
	done <-chan struct{}
}

func newHarness(t *testing.T, loggedIn bool) *runnerHarness {
	job := &blockingJob{
		started: make(chan context.Context, 100),
		release: make(chan struct{}),
	}
	chk := &fakeChecker{}
	chk.loggedIn.Store(loggedIn)
	return &runnerHarness{
		t:   t,
		r:   NewPurchasesRunner(context.Background(), job.run, chk),
		job: job,
		chk: chk,
	}
}

func (h *runnerHarness) start(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		h.r.Run(ctx)
		close(done)
	}()
	h.done = done
}

func (h *runnerHarness) started() context.Context {
	h.t.Helper()
	select {
	case ctx := <-h.job.started:
		return ctx
	case <-time.After(testTimeout):
		h.t.Fatal("job did not start")
		return nil
	}
}

func (h *runnerHarness) finish() {
	h.t.Helper()
	select {
	case h.job.release <- struct{}{}:
	case <-time.After(testTimeout):
		h.t.Fatal("no running job to release")
	}
}

func (h *runnerHarness) cancelled(ctx context.Context) {
	h.t.Helper()
	select {
	case <-ctx.Done():
	case <-time.After(testTimeout):
		h.t.Fatal("context was not cancelled")
	}
}

func (h *runnerHarness) isIdle() bool {
	h.r.mu.Lock()
	defer h.r.mu.Unlock()
	return !h.r.running && !h.r.pending && h.r.cancel == nil
}

func (h *runnerHarness) assertSettled(wantCalls int32) {
	h.t.Helper()
	if h.done != nil {
		select {
		case <-h.done:
		case <-time.After(testTimeout):
			h.t.Fatal("Run did not return")
		}
	}
	assert.Eventually(h.t, h.isIdle, testTimeout, time.Millisecond, "runner did not become idle")

	select {
	case <-h.job.started:
		h.t.Error("an unexpected extra run started")
	case <-time.After(quietPeriod):
	}

	assert.Equal(h.t, wantCalls, h.job.calls.Load(), "job runs")
	assert.LessOrEqual(h.t, h.job.maxActive.Load(), int32(1), "runs must never overlap")
}

func TestPurchasesRunner(t *testing.T) {
	category.Set(t, category.Unit)

	bg := context.Background()

	tests := []struct {
		name      string
		loggedIn  bool
		scenario  func(t *testing.T, h *runnerHarness)
		wantCalls int32
	}{
		{
			name:      "not logged in skips the job",
			loggedIn:  false,
			scenario:  func(t *testing.T, h *runnerHarness) { h.r.Run(bg) },
			wantCalls: 0,
		},
		{
			name:     "logged in runs the job once",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				h.start(bg)
				h.started()
				h.finish()
			},
			wantCalls: 1,
		},
		{
			name:     "cancelling the parent context stops the job",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				parent, cancel := context.WithCancel(bg)
				h.start(parent)
				ctx := h.started()
				cancel()
				h.cancelled(ctx)
			},
			wantCalls: 1,
		},
		{
			name:     "triggers during a run coalesce into one rerun",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				h.start(bg)
				first := h.started()

				h.r.Run(bg)
				h.r.Run(bg)
				h.r.Run(bg)
				assert.NoError(t, first.Err(), "extra triggers must not cancel the current run")

				h.finish()
				h.started()
				h.finish()
			},
			wantCalls: 2,
		},
		{
			name:     "pending rerun is skipped if logged out meanwhile",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				h.start(bg)
				h.started()
				h.r.Run(bg)
				h.chk.loggedIn.Store(false)
				h.finish()
			},
			wantCalls: 1,
		},
		{
			name:     "login when idle starts a run",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				assert.NoError(t, h.r.NotifyLogin(authSuccess))
				h.started()
				h.finish()
			},
			wantCalls: 1,
		},
		{
			name:     "login during a run cancels it and restarts",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				h.start(bg)
				first := h.started()

				assert.NoError(t, h.r.NotifyLogin(authSuccess))
				h.cancelled(first)

				second := h.started()
				assert.NoError(t, second.Err(), "the restarted run must get a fresh context")
				h.finish()
			},
			wantCalls: 2,
		},
		{
			name:     "logout during a run cancels it without restart",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				h.start(bg)
				first := h.started()

				h.chk.loggedIn.Store(false)
				assert.NoError(t, h.r.NotifyLogout(authSuccess))
				h.cancelled(first)
			},
			wantCalls: 1,
		},
		{
			name:     "logout when idle does nothing",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				assert.NoError(t, h.r.NotifyLogout(authSuccess))
			},
			wantCalls: 0,
		},
		{
			name:     "failed login during a run is ignored",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				h.start(bg)
				first := h.started()
				assert.NoError(t, h.r.NotifyLogin(authFailure))
				assert.NoError(t, first.Err(), "current run must not be cancelled")
				h.finish()
			},
			wantCalls: 1,
		},
		{
			name:     "failed logout during a run is ignored",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				h.start(bg)
				first := h.started()
				assert.NoError(t, h.r.NotifyLogout(authFailure))
				assert.NoError(t, first.Err(), "current run must not be cancelled")
				h.finish()
			},
			wantCalls: 1,
		},
		{
			name:     "failed login when idle does not start a run",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				assert.NoError(t, h.r.NotifyLogin(authFailure))
			},
			wantCalls: 0,
		},
		{
			name:     "MFA events have no effect",
			loggedIn: true,
			scenario: func(t *testing.T, h *runnerHarness) {
				assert.NoError(t, h.r.NotifyMFA(true))
				assert.NoError(t, h.r.NotifyMFA(false))
			},
			wantCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, tt.loggedIn)
			tt.scenario(t, h)
			h.assertSettled(tt.wantCalls)
		})
	}
}
