package daemon

import (
	"context"
	"errors"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc/metadata"
)

// connectServer can be used for connections triggered internally(such as from autoconnect or pause).
type connectServer struct {
	err error
}

var errServersUnavailable = errors.New("servers unavailable")

func (connectServer) SetHeader(metadata.MD) error  { return nil }
func (connectServer) SendHeader(metadata.MD) error { return nil }
func (connectServer) SetTrailer(metadata.MD)       {}
func (connectServer) Context() context.Context     { return nil }
func (connectServer) SendMsg(m interface{}) error  { return nil }
func (connectServer) RecvMsg(m interface{}) error  { return nil }
func (a *connectServer) Send(data *pb.Payload) error {
	switch data.GetType() {
	case internal.CodeFailure:
		a.err = errors.New("connect failure")
	case internal.CodeServerUnavailable:
		a.err = errServersUnavailable
	}
	return nil
}
