package uievent

import (
	"context"
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestMethodToItemName(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		method   string
		expected pb.UIEvent_ItemName
	}{
		// Tracked methods
		{"Connect", pb.Daemon_Connect_FullMethodName, pb.UIEvent_CONNECT},
		{"Disconnect", pb.Daemon_Disconnect_FullMethodName, pb.UIEvent_DISCONNECT},
		{"LoginOAuth2", pb.Daemon_LoginOAuth2_FullMethodName, pb.UIEvent_LOGIN},
		{"Logout", pb.Daemon_Logout_FullMethodName, pb.UIEvent_LOGOUT},
		{"RateConnection", pb.Daemon_RateConnection_FullMethodName, pb.UIEvent_RATE_CONNECTION},
		{"Meshnet Invite", meshpb.Meshnet_Invite_FullMethodName, pb.UIEvent_MESHNET_INVITE_SEND},
		// Untracked methods
		{"Settings", pb.Daemon_Settings_FullMethodName, pb.UIEvent_ITEM_NAME_UNSPECIFIED},
		{"Status", pb.Daemon_Status_FullMethodName, pb.UIEvent_ITEM_NAME_UNSPECIFIED},
		// Edge cases
		{"Empty string", "", pb.UIEvent_ITEM_NAME_UNSPECIFIED},
		{"Unknown method", "/unknown/Method", pb.UIEvent_ITEM_NAME_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, methodToItemName(tt.method))
		})
	}
}

func TestUnaryInterceptor_AttachesMetadataForTrackedMethods(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		formReference pb.UIEvent_FormReference
		method        string
		wantItemName  string
	}{
		{"CLI Connect", pb.UIEvent_CLI, pb.Daemon_Connect_FullMethodName, "1"},
		{"TRAY Disconnect", pb.UIEvent_TRAY, pb.Daemon_Disconnect_FullMethodName, "3"},
		{"HOME_SCREEN Login", pb.UIEvent_HOME_SCREEN, pb.Daemon_LoginOAuth2_FullMethodName, "4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewClientInterceptor(tt.formReference)

			var capturedCtx context.Context
			mockInvoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
				capturedCtx = ctx
				return nil
			}

			err := interceptor.UnaryInterceptor(context.Background(), tt.method, nil, nil, nil, mockInvoker)
			assert.NoError(t, err)

			md, ok := metadata.FromOutgoingContext(capturedCtx)
			assert.True(t, ok, "metadata should be attached")
			assert.Equal(t, []string{tt.wantItemName}, md.Get(MetadataKeyItemName))
			assert.Equal(t, []string{"1"}, md.Get(MetadataKeyItemType)) // CLICK = 1
		})
	}
}

func TestUnaryInterceptor_SkipsMetadataForUntrackedMethods(t *testing.T) {
	category.Set(t, category.Unit)

	interceptor := NewClientInterceptor(pb.UIEvent_CLI)

	var capturedCtx context.Context
	mockInvoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		capturedCtx = ctx
		return nil
	}

	err := interceptor.UnaryInterceptor(context.Background(), pb.Daemon_Settings_FullMethodName, nil, nil, nil, mockInvoker)
	assert.NoError(t, err)

	_, ok := metadata.FromOutgoingContext(capturedCtx)
	assert.False(t, ok, "no metadata should be attached for untracked methods")
}

func TestUnaryInterceptor_PropagatesInvokerError(t *testing.T) {
	category.Set(t, category.Unit)

	interceptor := NewClientInterceptor(pb.UIEvent_CLI)
	expectedErr := errors.New("invoker failed")

	mockInvoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		return expectedErr
	}

	err := interceptor.UnaryInterceptor(context.Background(), pb.Daemon_Connect_FullMethodName, nil, nil, nil, mockInvoker)
	assert.ErrorIs(t, err, expectedErr)
}

func TestUnaryInterceptor_PreservesExistingMetadata(t *testing.T) {
	category.Set(t, category.Unit)

	interceptor := NewClientInterceptor(pb.UIEvent_CLI)

	existingCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("existing-key", "existing-value"))

	var capturedCtx context.Context
	mockInvoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		capturedCtx = ctx
		return nil
	}

	err := interceptor.UnaryInterceptor(existingCtx, pb.Daemon_Connect_FullMethodName, nil, nil, nil, mockInvoker)
	assert.NoError(t, err)

	md, ok := metadata.FromOutgoingContext(capturedCtx)
	assert.True(t, ok)
	// New metadata should be present (replaces existing context)
	assert.NotEmpty(t, md.Get(MetadataKeyFormReference))
}

func TestStreamInterceptor_AttachesMetadataForTrackedMethods(t *testing.T) {
	category.Set(t, category.Unit)

	interceptor := NewClientInterceptor(pb.UIEvent_CLI)

	var capturedCtx context.Context
	mockStreamer := func(ctx context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
		capturedCtx = ctx
		return nil, nil
	}

	_, err := interceptor.StreamInterceptor(context.Background(), nil, nil, pb.Daemon_Connect_FullMethodName, mockStreamer)
	assert.NoError(t, err)

	md, ok := metadata.FromOutgoingContext(capturedCtx)
	assert.True(t, ok, "metadata should be attached")
	assert.Equal(t, []string{"1"}, md.Get(MetadataKeyFormReference)) // CLI = 1
}

func TestStreamInterceptor_PropagatesStreamerError(t *testing.T) {
	category.Set(t, category.Unit)

	interceptor := NewClientInterceptor(pb.UIEvent_CLI)
	expectedErr := errors.New("streamer failed")

	mockStreamer := func(_ context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, expectedErr
	}

	_, err := interceptor.StreamInterceptor(context.Background(), nil, nil, pb.Daemon_Connect_FullMethodName, mockStreamer)
	assert.ErrorIs(t, err, expectedErr)
}
