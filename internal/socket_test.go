package internal

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestAuthType(t *testing.T) {
	ucred := UcredAuth{Pid: 123, Uid: 456, Gid: 789}
	assert.Equal(t, "123:456:789", ucred.AuthType())
}

func FuzzUcredToStringToUcred(f *testing.F) {
	f.Add(int32(0), uint32(0), uint32(0))
	f.Add(int32(1000), uint32(1000), uint32(1000))
	f.Add(int32(-1), uint32(1), uint32(1))
	f.Fuzz(func(t *testing.T, a int32, b uint32, c uint32) {
		ucred := UcredAuth{a, b, c}
		resUcred, err := StringToUcred(ucred.AuthType())
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
		name    string
		uid     int
		grps    []string
		wantErr bool
	}{
		{
			name:    "test1: existing user id 1000",
			uid:     1000,
			grps:    []string{"nordvpn"},
			wantErr: false,
		},
		{
			name:    "test2: non existing user",
			uid:     9000,
			grps:    []string{"nordvpn"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := authenticateUser(&unix.Ucred{Uid: uint32(tt.uid)}, tt.grps); (err != nil) != tt.wantErr {
				t.Errorf("authenticateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
