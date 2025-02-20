package internal

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
	"google.golang.org/grpc/credentials"
)

// getUnixCreds returns info from unix socket connection about the process on the other end.
func getUnixCreds(conn net.Conn, authenticator SocketAuthenticator) (*unix.Ucred, error) {
	conn2 := extractConnection(conn)
	if conn2 == nil {
		return nil, fmt.Errorf("cannot extract connection of proper type, is netutil.LimitListener used?")
	}

	unixConn, ok := conn2.(*net.UnixConn)
	if !ok {
		return nil, fmt.Errorf("socket is not a unix socket")
	}

	rawConn, err := unixConn.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("getting raw connection: %w", err)
	}

	var ucred *unix.Ucred
	var internalErr error
	err = rawConn.Control(func(fd uintptr) {
		ucred, internalErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	if internalErr != nil {
		return nil, fmt.Errorf("doing GetsockoptUcred: %w", internalErr)
	}
	if err != nil {
		return nil, fmt.Errorf("doing rawConn Control: %w", err)
	}

	if err := authenticator.Authenticate(ucred); err != nil {
		return nil, err
	}

	return ucred, nil
}

// SocketAuthenticator provides abstraction over various authentication types.
type SocketAuthenticator interface {
	Authenticate(ucred *unix.Ucred) error
}

type DaemonAuthenticator struct{}

func NewDaemonAuthenticator() DaemonAuthenticator {
	return DaemonAuthenticator{}
}

func (DaemonAuthenticator) Authenticate(ucred *unix.Ucred) error {
	// root?
	if ucred.Uid == 0 {
		return nil
	}

	isGroup, err := IsInAllowedGroup(ucred.Uid)
	if err != nil {
		return err
	}

	if !isGroup {
		return ErrNoPermission
	}

	return nil
}

type FileshareAuthenticator struct {
	DaemonAuthenticator
	controllingUserUUID uint32
}

func NewFileshareAuthenticator(controlingUserUUID uint32) FileshareAuthenticator {
	return FileshareAuthenticator{
		controllingUserUUID: controlingUserUUID,
	}
}

func (f FileshareAuthenticator) Authenticate(ucred *unix.Ucred) error {
	if err := f.DaemonAuthenticator.Authenticate(ucred); err != nil {
		return err
	}

	if ucred.Uid != f.controllingUserUUID && ucred.Uid != 0 {
		return ErrNoPermission
	}

	return nil
}

// UnixSocketCredentials is used to retrieve linux user ID from unix socket connection between client and daemon
// Implements credentials.TransportCredentials to be passed to gRPC server initialization
type UnixSocketCredentials struct {
	mu            sync.Mutex
	authenticator SocketAuthenticator
}

func NewUnixSocketCredentials(authenticator SocketAuthenticator) *UnixSocketCredentials {
	return &UnixSocketCredentials{
		authenticator: authenticator,
	}
}

// ServerHandshake is called when client connects to daemon.
// We retrieve user ID which opened the client here.
func (cr *UnixSocketCredentials) ServerHandshake(c net.Conn) (net.Conn, credentials.AuthInfo, error) {
	// under snap and under stress load, cgo calls cause crash or deadlock, need serialize
	cr.mu.Lock()
	defer cr.mu.Unlock()

	creds, err := getUnixCreds(c, cr.authenticator)
	if err != nil || creds == nil {
		return c, UcredAuth{}, err
	}

	return c, UcredAuth(*creds), nil
}

// ClientHandshake is a stub to implement credentials.TransportCredentials
func (cr *UnixSocketCredentials) ClientHandshake(_ context.Context, _ string, c net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return c, nil, nil
}

// Info is a stub to implement credentials.TransportCredentials
func (cr *UnixSocketCredentials) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{}
}

// Clone is a stub to implement credentials.TransportCredentials
func (cr *UnixSocketCredentials) Clone() credentials.TransportCredentials {
	return cr
}

// OverrideServerName is a stub to implement credentials.TransportCredentials
func (cr *UnixSocketCredentials) OverrideServerName(string) error {
	return nil
}

// UcredAuth is a wrapper to use unix.Ucred as gRPC credentials.AuthInfo
type UcredAuth unix.Ucred

// AuthType returns "pid:uid:gid", for example "5555:1000:1000"
// Use StringToUcred to convert string back to unix.Ucred
func (u UcredAuth) AuthType() string {
	return strconv.Itoa(int(u.Pid)) + ":" + strconv.Itoa(int(u.Uid)) + ":" + strconv.Itoa(int(u.Gid))
}

// StringToUcred to convert string received from AuthType back to unix.Ucred
func StringToUcred(ucredStr string) (unix.Ucred, error) {
	idsStr := strings.Split(ucredStr, ":")
	if len(idsStr) != 3 {
		return unix.Ucred{}, fmt.Errorf("invalid ucred string: %s", ucredStr)
	}
	pidStr, uidStr, gidStr := idsStr[0], idsStr[1], idsStr[2]

	pid, err := strconv.ParseInt(pidStr, 10, 32)
	if err != nil {
		return unix.Ucred{}, fmt.Errorf("invalid ucred string: %s", ucredStr)
	}

	uid, err := strconv.ParseInt(uidStr, 10, 32)
	if err != nil {
		return unix.Ucred{}, fmt.Errorf("invalid ucred string: %s", ucredStr)
	}

	gid, err := strconv.ParseInt(gidStr, 10, 32)
	if err != nil {
		return unix.Ucred{}, fmt.Errorf("invalid ucred string: %s", ucredStr)
	}

	return unix.Ucred{Pid: int32(pid), Uid: uint32(uid), Gid: uint32(gid)}, nil
}

// using `netutil.LimitListener` but it has internal unexported type `limitListenerConn`
// which wraps original `net.UnixConn` value inside by embeding abstract interface `net.Conn`
// this way we cannot access `net.UnixConn` value, because of that we use `go reflection`
// to extract original wrapped value.
func extractConnection(c interface{}) net.Conn {
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	if v.Kind() == reflect.Struct {
		y := v.FieldByName("Conn")
		cc := y.Interface().(net.Conn)
		return cc
	}
	return nil
}
