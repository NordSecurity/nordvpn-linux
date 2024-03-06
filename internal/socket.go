package internal

import (
	"context"
	"fmt"
	"net"
	"os/user"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
	"google.golang.org/grpc/credentials"
)

var allowedGroups []string = []string{"nordvpn"}
var ErrNoPermission error = fmt.Errorf("requesting user does not have permissions")

func isInAllowedGroup(ucred *unix.Ucred) (bool, error) {
	userInfo, err := user.LookupId(fmt.Sprintf("%d", ucred.Uid))
	if err != nil {
		return false, fmt.Errorf("authenticate user, lookup user info: %s", err)
	}
	// user belongs to the allowed group?
	groups, err := userInfo.GroupIds()
	if err != nil {
		return false, fmt.Errorf("authenticate user, check user groups: %s", err)
	}

	for _, groupId := range groups {
		groupInfo, err := user.LookupGroupId(groupId)
		if err != nil {
			return false, fmt.Errorf("authenticate user, check user group: %s", err)
		}
		for _, allowGroupName := range allowedGroups {
			if groupInfo.Name == allowGroupName {
				return true, nil
			}
		}
	}

	return false, nil
}

// getUnixCreds returns info from unix socket connection about the process on the other end.
func getUnixCreds(conn net.Conn, authenticator SocketAuthenticator) (*unix.Ucred, error) {
	unixConn, ok := conn.(*net.UnixConn)
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

	isGroup, err := isInAllowedGroup(ucred)
	if err != nil {
		return err
	}

	if !isGroup {
		return ErrNoPermission
	}

	return nil
}

type FileshareAuthenticator struct {
	controllingUserUUID uint32
}

func NewFileshareAuthenticator(controlingUserUUID uint32) FileshareAuthenticator {
	return FileshareAuthenticator{
		controllingUserUUID: controlingUserUUID,
	}
}

func (f FileshareAuthenticator) Authenticate(ucred *unix.Ucred) error {
	if ucred.Uid == 0 {
		return nil
	}

	isGroup, err := isInAllowedGroup(ucred)
	if err != nil {
		return err
	}

	if !isGroup {
		return ErrNoPermission
	}

	if ucred.Uid != f.controllingUserUUID {
		return ErrNoPermission
	}

	return nil
}

// UnixSocketCredentials is used to retrieve linux user ID from unix socket connection between client and daemon
// Implements credentials.TransportCredentials to be passed to gRPC server initialization
type UnixSocketCredentials struct {
	authenticator SocketAuthenticator
}

func NewUnixSocketCredentials(authenticator SocketAuthenticator) UnixSocketCredentials {
	return UnixSocketCredentials{
		authenticator: authenticator,
	}
}

// ServerHandshake is called when client connects to daemon.
// We retrieve user ID which opened the client here.
func (u UnixSocketCredentials) ServerHandshake(c net.Conn) (net.Conn, credentials.AuthInfo, error) {
	creds, err := getUnixCreds(c, u.authenticator)
	if err != nil || creds == nil {
		return c, UcredAuth{}, err
	}

	return c, UcredAuth(*creds), nil
}

// ClientHandshake is a stub to implement credentials.TransportCredentials
func (UnixSocketCredentials) ClientHandshake(_ context.Context, _ string, c net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return c, nil, nil
}

// Info is a stub to implement credentials.TransportCredentials
func (UnixSocketCredentials) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{}
}

// Clone is a stub to implement credentials.TransportCredentials
func (UnixSocketCredentials) Clone() credentials.TransportCredentials {
	return UnixSocketCredentials{}
}

// OverrideServerName is a stub to implement credentials.TransportCredentials
func (UnixSocketCredentials) OverrideServerName(string) error {
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
