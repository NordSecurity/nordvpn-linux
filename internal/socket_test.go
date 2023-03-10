package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
