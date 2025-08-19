package norduser

import (
	"context"
	"log"
	"os/user"
	"strconv"
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/stats"
)

type connIDKey struct{}

type ConnTagger struct{}

var nextID int64

// Called once per transport connection.
func (c *ConnTagger) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	id := atomic.AddInt64(&nextID, 1)
	return context.WithValue(ctx, connIDKey{}, id)
}

// Optional: observe begin/end of the connection.
func (c *ConnTagger) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnBegin:
		log.Printf("conn_begin id=%d", ctx.Value(connIDKey{}))
	case *stats.ConnEnd:
		log.Printf("conn_end   id=%d", ctx.Value(connIDKey{}))
	}
}

func (c *ConnTagger) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context { return ctx }
func (c *ConnTagger) HandleRPC(context.Context, stats.RPCStats)                       {}

// StartNorduserdMiddleware provides a way to start/stop norduserd when handling nordvpnd gRPCs.
type StartNorduserdMiddleware struct {
	norduserd service.Service
}

func NewStartNorduserMiddleware(norduserd_service service.Service) StartNorduserdMiddleware {
	return StartNorduserdMiddleware{
		norduserd: norduserd_service,
	}
}

func (n *StartNorduserdMiddleware) middleware(ctx context.Context) {
	var ucred unix.Ucred
	peer, ok := peer.FromContext(ctx)
	if !ok || peer.AuthInfo == nil {
		log.Println("no peer/auth info found in stream context")
	} else {
		var err error
		ucred, err = internal.StringToUcred(peer.AuthInfo.AuthType())
		if err != nil {
			log.Println("failed to convert auth info to user credentials:", err.Error())
		}
	}

	u, err := user.LookupId(strconv.FormatInt(int64(ucred.Uid), 10))
	if err != nil {
		log.Println("failed to find user by UID:", err)
	}
	if err := n.norduserd.Enable(ucred.Uid, ucred.Gid, u.HomeDir); err != nil {
		log.Println("failed to enable norduserd:", err)
	}
}

func (n *StartNorduserdMiddleware) StreamMiddleware(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
) error {
	id, _ := ss.Context().Value(connIDKey{}).(int64)
	log.Printf("-> start ss: %s - %p - %d\n", info.FullMethod, ss.Context(), id)
	n.middleware(ss.Context())
	log.Printf("-> end ss:   %s - %p - %d\n", info.FullMethod, ss.Context(), id)

	return nil
}

func (n *StartNorduserdMiddleware) UnaryMiddleware(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
) (interface{}, error) {

	id, _ := ctx.Value(connIDKey{}).(int64)

	log.Printf("-> start: %s - %p - %d\n", info.FullMethod, ctx, id)
	n.middleware(ctx)
	log.Printf("<- end:   %s - %p - %d\n", info.FullMethod, ctx, id)

	return nil, nil
}
