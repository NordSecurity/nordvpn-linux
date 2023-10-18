package request

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

var (
	err1   = fmt.Errorf("error1")
	respH1 = &http.Response{ProtoMajor: 1}
	respH3 = &http.Response{ProtoMajor: 3}
)

type mockRoundTripper struct {
	duration time.Duration
	resp     *http.Response
	err      error
}

func (m mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	time.Sleep(m.duration)
	return m.resp, m.err
}

func newAtomicBool(val bool) *atomic.Bool {
	b := &atomic.Bool{}
	b.Store(val)
	return b
}

func newAtomicInt64(val int64) *atomic.Int64 {
	i := &atomic.Int64{}
	i.Store(val)
	return i
}

func TestRotatingRoundTripper_RoundTrip(t *testing.T) {
	category.Set(t, category.Unit)
	for _, tt := range []struct {
		name         string
		roundTripper *RotatingRoundTripper
		resp         *http.Response
		err          error
	}{
		{
			name: "h3",
			roundTripper: NewRotatingRoundTripper(
				mockRoundTripper{resp: respH1},
				mockRoundTripper{resp: respH3},
				time.Duration(0),
			),
			resp: respH3,
		},
		{
			name: "h3 without switch",
			roundTripper: &RotatingRoundTripper{
				roundTripperH1:     mockRoundTripper{resp: respH1},
				roundTripperH3:     mockRoundTripper{resp: respH3},
				isCurrentH3:        newAtomicBool(true),
				lastH3AttemptMilli: &atomic.Int64{},
				h3ReviveTime:       time.Duration(0),
			},
			resp: respH3,
		},
		{
			name: "h1",
			roundTripper: &RotatingRoundTripper{
				roundTripperH1:     mockRoundTripper{resp: respH1},
				roundTripperH3:     mockRoundTripper{resp: respH3},
				isCurrentH3:        newAtomicBool(false),
				lastH3AttemptMilli: newAtomicInt64(time.Now().Add(-time.Second).UnixMilli()),
				h3ReviveTime:       time.Second * 2,
			},
			resp: respH1,
		},
		{
			name: "h1 fails while on h3",
			roundTripper: &RotatingRoundTripper{
				roundTripperH1:     mockRoundTripper{err: err1},
				roundTripperH3:     mockRoundTripper{resp: respH3},
				isCurrentH3:        newAtomicBool(false),
				lastH3AttemptMilli: &atomic.Int64{},
				h3ReviveTime:       time.Duration(0),
			},
			resp: respH3,
		},
		{
			name: "h3 fails",
			roundTripper: &RotatingRoundTripper{
				roundTripperH1:     mockRoundTripper{resp: respH1},
				roundTripperH3:     mockRoundTripper{err: err1},
				isCurrentH3:        newAtomicBool(false),
				lastH3AttemptMilli: &atomic.Int64{},
				h3ReviveTime:       time.Duration(0),
			},
			resp: respH1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.roundTripper.RoundTrip(&http.Request{})
			assert.ErrorIs(t, tt.err, err)
			assert.Equal(t, tt.resp, resp)
		})
	}
}

func TestRotatingRoundTripper_RoundTripThreadSafety(t *testing.T) {
	category.Set(t, category.Unit)
	for _, tt := range []struct {
		name         string
		roundTripper *RotatingRoundTripper
		n            int
		duration     time.Duration
	}{
		{
			name: "requests happen simultaniously in after first iteration",
			n:    5,
			roundTripper: NewRotatingRoundTripper(
				mockRoundTripper{err: err1},
				mockRoundTripper{resp: respH3, duration: time.Millisecond * 100},
				time.Duration(0),
			),
			duration: time.Millisecond * 200,
		},
		{
			name: "http3 fails and subsequent calls wait for every h1 rt",
			n:    5,
			roundTripper: NewRotatingRoundTripper(
				mockRoundTripper{resp: respH1, duration: time.Millisecond * 100},
				mockRoundTripper{err: err1, duration: time.Millisecond * 200},
				time.Duration(0),
			),
			duration: time.Millisecond * 600,
		},
		{
			name: "http3 fails and subsequent calls wait for h3 once",
			n:    5,
			roundTripper: NewRotatingRoundTripper(
				mockRoundTripper{resp: respH1, duration: time.Millisecond * 100},
				mockRoundTripper{err: err1, duration: time.Millisecond * 200},
				time.Minute,
			),
			duration: time.Millisecond * 400,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()
			wg := sync.WaitGroup{}
			for i := 0; i < tt.n; i++ {
				if i == 0 {
					tt.roundTripper.RoundTrip(&http.Request{})
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					tt.roundTripper.RoundTrip(&http.Request{})
				}()
			}
			wg.Wait()
			now := time.Now()
			expEnd := startTime.Add(tt.duration)
			assert.True(
				t,
				now.After(expEnd) && now.Before(expEnd.Add(tt.duration/10)),
				"Expected:\n  From: %s\n    To: %s\nActual: %s",
				expEnd,
				expEnd.Add(tt.duration/10),
				now,
			)
		})
	}
}
