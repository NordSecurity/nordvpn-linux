package uievent

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/events"
	"google.golang.org/grpc"
)

// Middleware extracts UI event metadata from incoming gRPC requests
// and publishes them to the Moose analytics system.
type Middleware struct {
	publisher events.Publisher[events.UiItemsAction]
}

// NewMiddleware creates a new UI event middleware.
func NewMiddleware(publisher events.Publisher[events.UiItemsAction]) *Middleware {
	return &Middleware{
		publisher: publisher,
	}
}

// UnaryMiddleware is a gRPC unary middleware that extracts UI event
// metadata and publishes it to Moose analytics.
func (m *Middleware) UnaryMiddleware(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
) (interface{}, error) {
	m.publishUIEvent(ctx)
	return nil, nil
}

// StreamMiddleware is a gRPC stream middleware that extracts UI event
// metadata and publishes it to Moose analytics.
func (m *Middleware) StreamMiddleware(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
) error {
	m.publishUIEvent(ss.Context())
	return nil
}

// publishUIEvent extracts UI event context from the gRPC context and publishes
// it to Moose analytics. If the context is missing or invalid, no event is published.
func (m *Middleware) publishUIEvent(ctx context.Context) {
	if m.publisher == nil {
		return
	}

	uiCtx := FromIncomingContext(ctx)
	if uiCtx == nil || !IsValid(uiCtx) {
		return
	}

	action := ToMooseStrings(uiCtx)
	m.publisher.Publish(action)
}
