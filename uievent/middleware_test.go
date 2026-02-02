package uievent

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// mockPublisher is a test double for events.Publisher
type mockPublisher struct {
	published []events.UiItemsAction
}

func (m *mockPublisher) Publish(action events.UiItemsAction) {
	m.published = append(m.published, action)
}

// mockServerStream is a test double for grpc.ServerStream
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func TestMiddleware_UnaryMiddleware_PublishesEvent(t *testing.T) {
	category.Set(t, category.Unit)

	publisher := &mockPublisher{}
	middleware := NewMiddleware(publisher)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		MetadataKeyFormReference: []string{"1"}, // CLI
		MetadataKeyItemName:      []string{"1"}, // CONNECT
		MetadataKeyItemType:      []string{"1"}, // CLICK
		MetadataKeyItemValue:     []string{"1"}, // COUNTRY
	})

	_, err := middleware.UnaryMiddleware(ctx, nil, &grpc.UnaryServerInfo{})

	assert.NoError(t, err)
	require.Len(t, publisher.published, 1)
	assert.Equal(t, "cli", publisher.published[0].FormReference)
	assert.Equal(t, "connect", publisher.published[0].ItemName)
	assert.Equal(t, "click", publisher.published[0].ItemType)
	assert.Equal(t, "country", publisher.published[0].ItemValue)
}

func TestMiddleware_UnaryMiddleware_NoMetadata(t *testing.T) {
	category.Set(t, category.Unit)

	publisher := &mockPublisher{}
	middleware := NewMiddleware(publisher)

	_, err := middleware.UnaryMiddleware(context.Background(), nil, &grpc.UnaryServerInfo{})

	assert.NoError(t, err)
	assert.Empty(t, publisher.published)
}

func TestMiddleware_UnaryMiddleware_InvalidContext(t *testing.T) {
	category.Set(t, category.Unit)

	publisher := &mockPublisher{}
	middleware := NewMiddleware(publisher)

	// Only FormReference set, ItemName and ItemType are unspecified
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		MetadataKeyFormReference: []string{"1"}, // CLI
		MetadataKeyItemName:      []string{"0"}, // UNSPECIFIED
		MetadataKeyItemType:      []string{"0"}, // UNSPECIFIED
	})

	_, err := middleware.UnaryMiddleware(ctx, nil, &grpc.UnaryServerInfo{})

	assert.NoError(t, err)
	assert.Empty(t, publisher.published)
}

func TestMiddleware_UnaryMiddleware_NilPublisher(t *testing.T) {
	category.Set(t, category.Unit)

	middleware := NewMiddleware(nil)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		MetadataKeyFormReference: []string{"1"},
		MetadataKeyItemName:      []string{"1"},
		MetadataKeyItemType:      []string{"1"},
	})

	// Should not panic
	_, err := middleware.UnaryMiddleware(ctx, nil, &grpc.UnaryServerInfo{})
	assert.NoError(t, err)
}

func TestMiddleware_StreamMiddleware_PublishesEvent(t *testing.T) {
	category.Set(t, category.Unit)

	publisher := &mockPublisher{}
	middleware := NewMiddleware(publisher)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		MetadataKeyFormReference: []string{"2"}, // TRAY
		MetadataKeyItemName:      []string{"2"}, // CONNECT_RECENTS
		MetadataKeyItemType:      []string{"1"}, // CLICK
		MetadataKeyItemValue:     []string{"2"}, // CITY
	})

	err := middleware.StreamMiddleware(nil, &mockServerStream{ctx: ctx}, &grpc.StreamServerInfo{})

	assert.NoError(t, err)
	require.Len(t, publisher.published, 1)
	assert.Equal(t, "tray", publisher.published[0].FormReference)
	assert.Equal(t, "connect_recents", publisher.published[0].ItemName)
	assert.Equal(t, "click", publisher.published[0].ItemType)
	assert.Equal(t, "city", publisher.published[0].ItemValue)
}

func TestMiddleware_StreamMiddleware_NoMetadata(t *testing.T) {
	category.Set(t, category.Unit)

	publisher := &mockPublisher{}
	middleware := NewMiddleware(publisher)

	assert.NoError(t, middleware.StreamMiddleware(nil, &mockServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{}))
	assert.Empty(t, publisher.published)
}

func TestMiddleware_StreamMiddleware_NilPublisher(t *testing.T) {
	category.Set(t, category.Unit)

	middleware := NewMiddleware(nil)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		MetadataKeyFormReference: []string{"1"},
		MetadataKeyItemName:      []string{"1"},
		MetadataKeyItemType:      []string{"1"},
	})

	// Should not panic
	assert.NoError(t, middleware.StreamMiddleware(nil, &mockServerStream{ctx: ctx}, &grpc.StreamServerInfo{}))
}

func TestMiddleware_IntegrationScenario_TrayConnectRecents(t *testing.T) {
	category.Set(t, category.Unit)

	// Simulate a real scenario: Tray client connects via recent connections
	publisher := &mockPublisher{}
	middleware := NewMiddleware(publisher)

	// Client side: attach metadata to outgoing context
	clientCtx := AttachToOutgoingContext(context.Background(), &UIEventContext{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      pb.UIEvent_CONNECT_RECENTS,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     pb.UIEvent_CITY,
	})
	outgoingMD, _ := metadata.FromOutgoingContext(clientCtx)

	// Server side: receive as incoming context and process
	_, _ = middleware.UnaryMiddleware(
		metadata.NewIncomingContext(context.Background(), outgoingMD),
		nil,
		&grpc.UnaryServerInfo{},
	)

	// Verify the event was published correctly
	require.Len(t, publisher.published, 1)
	assert.Equal(t, "tray", publisher.published[0].FormReference)
	assert.Equal(t, "connect_recents", publisher.published[0].ItemName)
	assert.Equal(t, "click", publisher.published[0].ItemType)
	assert.Equal(t, "city", publisher.published[0].ItemValue)
}
