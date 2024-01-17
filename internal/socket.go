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

// UnixSocketCredentials is used to retrieve linux user ID from unix socket connection between client and daemon
// Implements credentials.TransportCredentials to be passed to gRPC server initialization
type UnixSocketCredentials struct{}

// ServerHandshake is called when client connects to daemon.
// We retrieve user ID which opened the client here.
func (UnixSocketCredentials) ServerHandshake(c net.Conn) (net.Conn, credentials.AuthInfo, error) {
	creds, err := getUnixCreds(c)
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

// getUnixCreds returns info from unix socket connection about the process on the other end.
func getUnixCreds(conn net.Conn) (*unix.Ucred, error) {
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

	if err := authenticateUser(ucred); err != nil {
		return nil, err
	}

	return ucred, nil
}

func authenticateUser(ucred *unix.Ucred) error {
	// root?
	if ucred.Uid == 0 {
		return nil
	}
	userInfo, err := user.LookupId(fmt.Sprintf("%d", ucred.Uid))
	if err != nil {
		return fmt.Errorf("authenticate user, lookup user info: %s", err)
	}
	// user belongs to 'nordvpn' or 'sudo' group?
	groups, err := userInfo.GroupIds()
	if err != nil {
		return fmt.Errorf("authenticate user, check user groups: %s", err)
	}
	for _, groupId := range groups {
		groupInfo, err := user.LookupGroupId(groupId)
		if err != nil {
			return fmt.Errorf("authenticate user, check user group: %s", err)
		}
		if groupInfo.Name == "nordvpn" || groupInfo.Name == "sudo" {
			return nil
		}
	}
	return fmt.Errorf("requesting user does not have permissions")
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
