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
	h3ReviveTime       time.Duration
}

func NewRotatingRoundTripper(
	roundTripperH1 http.RoundTripper,
	roundTripperH3 http.RoundTripper,
	h3ReviveTime time.Duration,
) *RotatingRoundTripper {
	return &RotatingRoundTripper{
		roundTripperH1:     roundTripperH1,
		roundTripperH3:     roundTripperH3,
		h3ReviveTime:       h3ReviveTime,
		lastH3AttemptMilli: &atomic.Int64{},
		isCurrentH3:        &atomic.Bool{},
	}
}

func (rt *RotatingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !rt.isCurrentH3.Load() && time.Now().After(time.UnixMilli(rt.lastH3AttemptMilli.Load()).Add(rt.h3ReviveTime)) {
		now := time.Now()
		rt.lastH3AttemptMilli.Store(now.UnixMilli())
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
