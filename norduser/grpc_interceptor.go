package norduser

import (
	"context"
	"log"
	"os/user"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

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
