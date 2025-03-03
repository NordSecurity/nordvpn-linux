package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

const (
	h1RespBody = "h1 body"
	h3RespBody = "h3 body"
)

type responseTemplate struct {
	protoMajor int
	body       string
}

var (
	err1           = fmt.Errorf("error1")
	respH1Template = responseTemplate{
		protoMajor: 1,
		body:       h1RespBody,
	}
	respH3Template = responseTemplate{
		protoMajor: 3,
		body:       h3RespBody,
	}
)

type mockRoundTripper struct {
	duration         time.Duration
	responseTemplate responseTemplate
	err              error
}

func (m mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	time.Sleep(m.duration)

	resp := http.Response{
		ProtoMajor: m.responseTemplate.protoMajor,
		Body:       io.NopCloser(strings.NewReader(m.responseTemplate.body)),
	}
	return &resp, m.err
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
		resp         responseTemplate
		err          error
	}{
		{
			name: "h3",
			roundTripper: NewRotatingRoundTripper(
				mockRoundTripper{responseTemplate: respH1Template},
				mockRoundTripper{responseTemplate: respH3Template},
				time.Duration(0),
			),
			resp: respH3Template,
		},
		{
			name: "h3 without switch",
			roundTripper: &RotatingRoundTripper{
				roundTripperH1:     mockRoundTripper{responseTemplate: respH1Template},
				roundTripperH3:     mockRoundTripper{responseTemplate: respH3Template},
				isCurrentH3:        newAtomicBool(true),
				lastH3AttemptMilli: &atomic.Int64{},
				h3ReviveTime:       time.Duration(0),
			},
			resp: respH3Template,
		},
		{
			name: "h1",
			roundTripper: &RotatingRoundTripper{
				roundTripperH1:     mockRoundTripper{responseTemplate: respH1Template},
				roundTripperH3:     mockRoundTripper{responseTemplate: respH3Template},
				isCurrentH3:        newAtomicBool(false),
				lastH3AttemptMilli: newAtomicInt64(time.Now().Add(-time.Second).UnixMilli()),
				h3ReviveTime:       time.Second * 2,
			},
			resp: respH1Template,
		},
		{
			name: "h1 fails while on h3",
			roundTripper: &RotatingRoundTripper{
				roundTripperH1:     mockRoundTripper{err: err1},
				roundTripperH3:     mockRoundTripper{responseTemplate: respH3Template},
				isCurrentH3:        newAtomicBool(false),
				lastH3AttemptMilli: &atomic.Int64{},
				h3ReviveTime:       time.Duration(0),
			},
			resp: respH3Template,
		},
		{
			name: "h3 fails",
			roundTripper: &RotatingRoundTripper{
				roundTripperH1:     mockRoundTripper{responseTemplate: respH1Template},
				roundTripperH3:     mockRoundTripper{err: err1},
				isCurrentH3:        newAtomicBool(false),
				lastH3AttemptMilli: &atomic.Int64{},
				h3ReviveTime:       time.Duration(0),
			},
			resp: respH1Template,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.roundTripper.RoundTrip(&http.Request{})
			if err != nil {
				defer resp.Body.Close()
			}
			assert.ErrorIs(t, tt.err, err)
			assert.Equal(t, tt.resp.protoMajor, resp.ProtoMajor)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.resp.body, string(body))
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
			name: "requests happen simultaneously in after first iteration",
			n:    5,
			roundTripper: NewRotatingRoundTripper(
				mockRoundTripper{err: err1},
				mockRoundTripper{responseTemplate: respH3Template, duration: time.Millisecond * 100},
				time.Duration(0),
			),
			duration: time.Millisecond * 200,
		},
		{
			name: "http3 fails and subsequent calls wait for every h1 rt",
			n:    5,
			roundTripper: NewRotatingRoundTripper(
				mockRoundTripper{responseTemplate: respH1Template, duration: time.Millisecond * 100},
				mockRoundTripper{err: err1, duration: time.Millisecond * 200},
				time.Duration(0),
			),
			duration: time.Millisecond * 600,
		},
		{
			name: "http3 fails and subsequent calls wait for h3 once",
			n:    5,
			roundTripper: NewRotatingRoundTripper(
				mockRoundTripper{responseTemplate: respH1Template, duration: time.Millisecond * 100},
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
					resp, err := tt.roundTripper.RoundTrip(&http.Request{})
					if err != nil {
						defer resp.Body.Close()
					}
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					resp, err := tt.roundTripper.RoundTrip(&http.Request{})
					if err != nil {
						defer resp.Body.Close()
					}
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
