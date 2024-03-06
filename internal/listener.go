package internal

import (
	"net"

	"golang.org/x/net/netutil"
)

// LimitListener customized limiting listener
type LimitListener struct {
	net.Listener
	originalListener net.Listener
}

func NewLimitListener(ol net.Listener) *LimitListener {
	return &LimitListener{
		// enforce how many connections at the same time can be
		// if some bad actor will try to create big load then
		// all subsequent requests will be held waiting until
		// some previous requests are finished to handle
		Listener:         netutil.LimitListener(ol, 100),
		originalListener: ol,
	}
}

// Accept intercept original connection to extract user credetials
func (l *LimitListener) Accept() (net.Conn, error) {
	conn, err := l.originalListener.Accept()
	if err != nil {
		return nil, err
	}
	// verify a requesting user
	_, err = getUnixCreds(conn, DaemonAuthenticator{})
	if err != nil {
		return nil, err
	}
	// if all is good, pass execution down the chain
	return conn, nil
}
