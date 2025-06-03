package request

import (
	"net/http"
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
)

type PublishingRoundTripper struct {
	roundTripper http.RoundTripper
	publisher    events.Publisher[events.DataRequestAPI]
}

func NewPublishingRoundTripper(
	roundTripper http.RoundTripper,
	publisher events.Publisher[events.DataRequestAPI],
) *PublishingRoundTripper {
	return &PublishingRoundTripper{
		roundTripper: roundTripper,
		publisher:    publisher,
	}
}

func (rt *PublishingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()
	rt.publisher.Publish(events.DataRequestAPI{
		Request:   req,
		Error:     nil,
		Duration:  time.Since(startTime),
		IsAttempt: true,
	})
	resp, err := rt.roundTripper.RoundTrip(req)
	rt.publisher.Publish(events.DataRequestAPI{
		Request:   req,
		Response:  resp,
		Error:     err,
		Duration:  time.Since(startTime),
		IsAttempt: false,
	})
	return resp, err
}
