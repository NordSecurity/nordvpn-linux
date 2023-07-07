package request

import (
	"net/http"
	"sync/atomic"
	"time"
)

type RotatingRoundTripper struct {
	roundTripperH1     http.RoundTripper
	roundTripperH3     http.RoundTripper
	isCurrentH3        *atomic.Bool
	lastH3AttemptMilli *atomic.Int64
	h3ReviveTime       *atomic.Int64
}

func NewRotatingRoundTripper(
	roundTripperH1 http.RoundTripper,
	roundTripperH3 http.RoundTripper,
	h3ReviveTime time.Duration,
) *RotatingRoundTripper {
	t := atomic.Int64{}
	t.Store(int64(h3ReviveTime))
	return &RotatingRoundTripper{
		roundTripperH1:     roundTripperH1,
		roundTripperH3:     roundTripperH3,
		h3ReviveTime:       &t,
		lastH3AttemptMilli: &atomic.Int64{},
		isCurrentH3:        &atomic.Bool{},
	}
}

func (rt *RotatingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !rt.isCurrentH3.Load() && time.Now().After(time.UnixMilli(rt.lastH3AttemptMilli.Load())) {
		now := time.Now()
		rt.lastH3AttemptMilli.Store(int64(now.UnixMilli()))
		rt.isCurrentH3.Store(true)
	}
	if rt.isCurrentH3.Load() {
		resp, err := rt.roundTripperH3.RoundTrip(req)
		if err == nil {
			return resp, err
		}
		rt.isCurrentH3.Store(false)
	}
	return rt.roundTripperH1.RoundTrip(req)
}
