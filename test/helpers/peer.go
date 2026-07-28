package helpers

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc/peer"
)

// PeerCtx returns a context carrying peer credentials for the given uid.
func PeerCtx(uid uint32) context.Context {
	return peer.NewContext(
		context.Background(),
		&peer.Peer{AuthInfo: internal.UcredAuth{Uid: uid}},
	)
}
