package uievent

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"google.golang.org/grpc"
)

// ClientInterceptor is a gRPC client interceptor that attaches UI event metadata
// to outgoing requests for tracked methods.
type ClientInterceptor struct {
	formReference pb.UIEvent_FormReference
}

// NewClientInterceptor creates a new client interceptor with the specified form reference.
func NewClientInterceptor(formRef pb.UIEvent_FormReference) *ClientInterceptor {
	return &ClientInterceptor{formReference: formRef}
}

// UnaryInterceptor is a gRPC unary client interceptor that attaches UI event metadata.
func (i *ClientInterceptor) UnaryInterceptor(
	ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	itemName := methodToItemName(method)
	if itemName == pb.UIEvent_ITEM_NAME_UNSPECIFIED {
		// Skip metadata attachment
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	uiCtx := &UIEventContext{
		FormReference: i.formReference,
		ItemName:      itemName,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
	}

	enrichedCtx := AttachToOutgoingContext(ctx, uiCtx)
	return invoker(enrichedCtx, method, req, reply, cc, opts...)
}

// StreamInterceptor is a gRPC stream client interceptor that attaches UI event metadata.
func (i *ClientInterceptor) StreamInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	itemName := methodToItemName(method)
	if itemName == pb.UIEvent_ITEM_NAME_UNSPECIFIED {
		// Not a tracked method, skip metadata attachment
		return streamer(ctx, desc, cc, method, opts...)
	}

	uiCtx := &UIEventContext{
		FormReference: i.formReference,
		ItemName:      itemName,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
	}

	enrichedCtx := AttachToOutgoingContext(ctx, uiCtx)
	return streamer(enrichedCtx, desc, cc, method, opts...)
}

// methodToItemName maps a gRPC method name to a UIEvent ItemName.
// Returns ITEM_NAME_UNSPECIFIED for methods that are not tracked.
func methodToItemName(method string) pb.UIEvent_ItemName {
	switch method {
	case pb.Daemon_Connect_FullMethodName:
		return pb.UIEvent_CONNECT
	case pb.Daemon_Disconnect_FullMethodName:
		return pb.UIEvent_DISCONNECT
	case pb.Daemon_LoginOAuth2_FullMethodName:
		return pb.UIEvent_LOGIN
	case pb.Daemon_Logout_FullMethodName:
		return pb.UIEvent_LOGOUT
	case pb.Daemon_RateConnection_FullMethodName:
		return pb.UIEvent_RATE_CONNECTION
	case meshpb.Meshnet_Invite_FullMethodName:
		return pb.UIEvent_MESHNET_INVITE_SEND
	default:
		return pb.UIEvent_ITEM_NAME_UNSPECIFIED
	}
}
