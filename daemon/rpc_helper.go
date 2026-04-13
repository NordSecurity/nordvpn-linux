package daemon

import (
	"context"
	"errors"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc/peer"
)

var (
	errNoPeerInfo  = errors.New("failed to retrieve gRPC peer information from the context")
	errNoUcredAuth = errors.New("failed to extract ucred out of gRPC peer info")
)

// getCallerCred extracts kernel-verified caller credentials from a gRPC context.
func getCallerCred(ctx context.Context) (internal.UcredAuth, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return internal.UcredAuth{}, errNoPeerInfo
	}
	cred, ok := p.AuthInfo.(internal.UcredAuth)
	if !ok {
		return internal.UcredAuth{}, errNoUcredAuth
	}
	return cred, nil
}
