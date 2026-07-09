package norduser

import (
	"context"
	"os/user"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"google.golang.org/grpc"
)

// StartNorduserdMiddleware provides a way to start/stop norduserd when handling nordvpnd gRPCs.
type StartNorduserdMiddleware struct {
	norduserd service.Service
}

func NewStartNorduserMiddleware(norduserdService service.Service) StartNorduserdMiddleware {
	return StartNorduserdMiddleware{
		norduserd: norduserdService,
	}
}

func (n *StartNorduserdMiddleware) middleware(ctx context.Context) {
	ucred, err := internal.UcredFromContext(ctx)
	if err != nil {
		log.Info("failed to get peer credentials in the middleware:", err)
	}

	u, err := user.LookupId(strconv.FormatInt(int64(ucred.Uid), 10))
	if err != nil {
		log.Info("failed to find user by UID:", err)
	}
	if err := n.norduserd.Enable(ucred.Uid, ucred.Gid, u.HomeDir); err != nil {
		log.Info("failed to enable norduserd:", err)
	}
}

func (n *StartNorduserdMiddleware) StreamMiddleware(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
) error {
	n.middleware(ss.Context())

	return nil
}

func (n *StartNorduserdMiddleware) UnaryMiddleware(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
) (any, error) {
	n.middleware(ctx)

	return nil, nil
}
