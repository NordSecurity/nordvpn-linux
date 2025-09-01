package clientid

import (
	"context"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// SetClientIDInterceptor provides methods that insert the client ID into the gRPC metadata.
type SetClientIDInterceptor struct {
	clientID pb.ClientID
}

func NewInsertClientIDInterceptor(clientID pb.ClientID) SetClientIDInterceptor {
	return SetClientIDInterceptor{
		clientID: clientID,
	}
}

// SetMetadataUnaryInterceptor inserts metadata into unary gRPC
func (i *SetClientIDInterceptor) SetMetadataUnaryInterceptor(ctx context.Context,
	method string,
	req, reply any, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	metadataCtx := metadata.AppendToOutgoingContext(ctx, clientIDMetadataKey, strconv.Itoa(int(i.clientID.Number())))
	return invoker(metadataCtx, method, req, reply, cc, opts...)
}

// SetMetadataStreamInterceptor inserts metadata into a stream gRPC
func (i *SetClientIDInterceptor) SetMetadataStreamInterceptor(ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {
	metadataCtx := metadata.AppendToOutgoingContext(ctx, clientIDMetadataKey, strconv.Itoa(int(i.clientID.Number())))
	return streamer(metadataCtx, desc, cc, method, opts...)
}
