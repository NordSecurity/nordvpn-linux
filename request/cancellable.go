package request

import (
	"context"
	"net/http"
)

type CancellableRoundTripper struct {
	inner http.RoundTripper
	ctx   context.Context
}

func NewCancellableRoundTripper(ctx context.Context) *CancellableRoundTripper {
	return &CancellableRoundTripper{
		inner: http.DefaultTransport,
		ctx:   ctx,
	}
}

func (ct *CancellableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqWithCtx := req.WithContext(ct.ctx)
	return ct.inner.RoundTrip(reqWithCtx)
}
