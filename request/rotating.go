package request

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
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

func (rt *RotatingRoundTripper) roundTripH3(req *http.Request) (*http.Response, error) {
	resp, err := rt.roundTripperH3.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// HTTP/3 responses are streamed, so error can occur while reading the response body. We need to attempt a read here
	// so all potential errors can be detected and we can rotate to HTTP/1
	var buf bytes.Buffer
	reader := io.LimitReader(resp.Body, internal.MaxBytesLimit)
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	resp.Body = io.NopCloser(&buf)
	return resp, nil
}

func (rt *RotatingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !rt.isCurrentH3.Load() && time.Now().After(time.UnixMilli(rt.lastH3AttemptMilli.Load()).Add(rt.h3ReviveTime)) {
		now := time.Now()
		rt.lastH3AttemptMilli.Store(now.UnixMilli())
		rt.isCurrentH3.Store(true)
	}
	if rt.isCurrentH3.Load() {
		resp, err := rt.roundTripH3(req)
		if err == nil {
			return resp, err
		}
		log.Println(internal.ErrorPrefix, "HTTP/3 request failed:", err, "rotating to HTTP/1")
		rt.isCurrentH3.Store(false)
	}
	return rt.roundTripperH1.RoundTrip(req)
}
