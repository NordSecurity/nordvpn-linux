package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc/peer"
)

const trayTestUID uint32 = 1000

// peerCtx returns a context carrying kernel-verified peer credentials for the given uid.
// Use this in unit tests that call RPC handlers which require getCallerCred to succeed.
func peerCtx(uid uint32) context.Context {
	return peer.NewContext(
		context.Background(),
		&peer.Peer{AuthInfo: internal.UcredAuth{Uid: uid}},
	)
}
