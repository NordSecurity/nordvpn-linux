package internal

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/peer"
)

func TestAuthType(t *testing.T) {
	ucred := UcredAuth{Pid: 123, Uid: 456, Gid: 789}
	assert.Equal(t, "123:456:789", ucred.AuthType())
}

func FuzzUcredFromContext(f *testing.F) {
	f.Add(int32(0), uint32(0), uint32(0))
	f.Add(int32(1000), uint32(1000), uint32(1000))
	f.Add(int32(-1), uint32(1), uint32(1))
	f.Fuzz(func(t *testing.T, a int32, b uint32, c uint32) {
		ucred := UcredAuth{a, b, c}
		ctx := peer.NewContext(context.Background(), &peer.Peer{
			AuthInfo: ucred,
		})
		resUcred, err := UcredFromContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, ucred.Pid, resUcred.Pid)
		assert.Equal(t, ucred.Gid, resUcred.Gid)
		assert.Equal(t, ucred.Uid, resUcred.Uid)
	})
}

func Test_authenticateUser(t *testing.T) {
	// we need to execute this test on tester docker image
	category.Set(t, category.Integration)

	// This test assumes there is a access to system users and groups
	// and there is a user with ID=1000 (usually this is the first user
	// created after install if it is a host system) and this user
	// should belong to the `nordvpn` group.
	tests := []struct {
		name          string
		uid           int
		grps          []string
		authenticator SocketAuthenticator
		wantErr       bool
	}{
		{
			name:          "test1: existing user id 1000",
			uid:           1000,
			grps:          []string{"nordvpn"},
			wantErr:       false,
			authenticator: DaemonAuthenticator{},
		},
		{
			name:          "test2: non existing user",
			uid:           9000,
			grps:          []string{"nordvpn"},
			wantErr:       true,
			authenticator: DaemonAuthenticator{},
		},
		{
			name:          "test3: existing user id 1000, fileshare authentication",
			uid:           1000,
			grps:          []string{"nordvpn"},
			wantErr:       false,
			authenticator: FileshareAuthenticator{controllingUserUUID: 1000},
		},
		{
			name:          "test4: non exisiting user, fileshare authentication",
			uid:           9000,
			grps:          []string{"nordvpn"},
			wantErr:       true,
			authenticator: FileshareAuthenticator{controllingUserUUID: 1000},
		},
		{
			name:          "test5: non controlling user, fileshare authentication",
			uid:           1000,
			grps:          []string{"nordvpn"},
			wantErr:       true,
			authenticator: FileshareAuthenticator{controllingUserUUID: 2000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.authenticator.Authenticate(&unix.Ucred{Uid: uint32(tt.uid)})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type MockNetConn struct{}

func (c MockNetConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (c MockNetConn) Write(b []byte) (n int, err error)  { return 0, nil }
func (c MockNetConn) Close() error                       { return nil }
func (c MockNetConn) LocalAddr() net.Addr                { return nil }
func (c MockNetConn) RemoteAddr() net.Addr               { return nil }
func (c MockNetConn) SetDeadline(t time.Time) error      { return nil }
func (c MockNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (c MockNetConn) SetWriteDeadline(t time.Time) error { return nil }

type MockWrappConnection struct{ net.Conn }

func Test_ExtractConnection(t *testing.T) {
	category.Set(t, category.Unit)

	// test with object
	c1 := MockNetConn{}
	c2 := MockWrappConnection{Conn: c1}
	c3 := extractConnection(c2)
	assert.NotNil(t, c3)

	// test with pointer
	var c01 net.Conn = &MockNetConn{}
	c02 := &MockWrappConnection{Conn: c01}
	c03 := extractConnection(c02)
	assert.NotNil(t, c03)
}

type otherAuthInfo struct{}

func (otherAuthInfo) AuthType() string { return "other" }

func TestUcredFromContext(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		ctx     context.Context
		want    unix.Ucred
		wantErr bool
	}{
		{
			name:    "no peer in context",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "peer with nil auth info",
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				AuthInfo: nil,
			}),
			wantErr: true,
		},
		{
			name: "peer with wrong auth info type",
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				AuthInfo: otherAuthInfo{},
			}),
			wantErr: true,
		},
		{
			name: "peer with ucred auth info",
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				AuthInfo: UcredAuth{Pid: 123, Uid: 456, Gid: 789},
			}),
			want: unix.Ucred{Pid: 123, Uid: 456, Gid: 789},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UcredFromContext(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
