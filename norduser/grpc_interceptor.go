package norduser

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// StartNorduserdMiddleware provides a way to start/stop norduserd when handling nordvpnd gRPCs.
type StartNorduserdMiddleware struct {
	nodruserd *service.Combined
}

func NewStartNorduserMiddleware(norduserd *service.Combined) StartNorduserdMiddleware {
	return StartNorduserdMiddleware{
		nodruserd: norduserd,
	}
}

func (n *StartNorduserdMiddleware) middleware(ctx context.Context) error {
	var ucred unix.Ucred
	peer, ok := peer.FromContext(ctx)
	if !ok || peer.AuthInfo == nil {
		log.Println("no peer/auth info found in stream context")
		return nil
	} else {
		var err error
		ucred, err = internal.StringToUcred(peer.AuthInfo.AuthType())
		if err != nil {
			log.Println("failed to convert auth info to user credentials: ", err.Error())
			return nil
		}
	}

	if err := n.nodruserd.Enable(ucred.Uid, ucred.Gid); err != nil {
		log.Println("failed to enable norduserd: ", err)
	}

	return nil
}

func (n *StartNorduserdMiddleware) StreamMiddleware(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo) error {
	n.middleware(ss.Context())

	return nil
}

func (n *StartNorduserdMiddleware) UnaryMiddleware(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo) (interface{}, error) {
	n.middleware(ctx)

	return nil, nil
}
