package request

import (
	"context"
	"net/http"
)

// ContextRoundTripper adds a common context to requests that pass through it.
type ContextRoundTripper struct {
	inner http.RoundTripper
	ctx   context.Context
}

// NewCancellableRoundTripper creates a new ContextRoundTripper.
// ctx is added to request passing through the RoundTripper.
func NewCancellableRoundTripper(ctx context.Context) *ContextRoundTripper {
	return &ContextRoundTripper{
		inner: http.DefaultTransport,
		ctx:   ctx,
	}
}

func (ct *ContextRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqWithCtx := req.WithContext(ct.ctx)
	return ct.inner.RoundTrip(reqWithCtx)
}
