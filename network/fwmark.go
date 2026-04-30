package network

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// NewFwmarkControlFn returns a Control function for net.Dialer that applies
// the specified fwmark to the socket.
func NewFwmarkControlFn(fwmark uint32) func(network, address string, conn syscall.RawConn) error {
	return func(_, _ string, conn syscall.RawConn) error {
		var operr error
		if err := conn.Control(func(fd uintptr) {
			operr = syscall.SetsockoptInt(
				int(fd),
				unix.SOL_SOCKET,
				unix.SO_MARK,
				int(fwmark),
			)
		}); err != nil {
			return err
		}
		return operr
	}
}
