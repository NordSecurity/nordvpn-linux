package daemon

import (
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"

	"github.com/stretchr/testify/assert"
)

func TestAtomicLoginTypeBasicOperations(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*atomicLoginType)
		check func(*testing.T, *atomicLoginType)
	}{
		{
			name:  "initial state is unknown",
			setup: func(alt *atomicLoginType) {},
			check: func(t *testing.T, alt *atomicLoginType) {
				assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, alt.get())
				assert.True(t, alt.isUnknown())
				assert.False(t, alt.wasStarted())
			},
		},
		{
			name: "reset to unknown",
			setup: func(alt *atomicLoginType) {
				alt.set(pb.LoginType_LoginType_LOGIN)
				alt.reset()
			},
			check: func(t *testing.T, alt *atomicLoginType) {
				assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, alt.get())
				assert.True(t, alt.isUnknown())
				assert.False(t, alt.wasStarted())
			},
		},
		{
			name: "isAltered returns true when types differ",
			setup: func(alt *atomicLoginType) {
				alt.set(pb.LoginType_LoginType_LOGIN)
			},
			check: func(t *testing.T, alt *atomicLoginType) {
				assert.True(t, alt.isAltered(pb.LoginType_LoginType_SIGNUP))
				assert.True(t, alt.isAltered(pb.LoginType_LoginType_UNKNOWN))
				assert.False(t, alt.isAltered(pb.LoginType_LoginType_LOGIN))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alt := NewAtomicLoginType()
			tt.setup(alt)
			tt.check(t, alt)
		})
	}
}

func TestAtomicLoginTypeConcurrency(t *testing.T) {
	alt := NewAtomicLoginType()

	// Test concurrent reads and writes
	var wg sync.WaitGroup
	iterations := 1000

	// Writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if id%2 == 0 {
					alt.set(pb.LoginType_LoginType_LOGIN)
				} else {
					alt.set(pb.LoginType_LoginType_SIGNUP)
				}
			}
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = alt.get()
				_ = alt.isUnknown()
				_ = alt.wasStarted()
			}
		}()
	}

	// Reseters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations/10; j++ {
				alt.reset()
			}
		}()
	}

	wg.Wait()

	// Final state should be valid
	finalValue := alt.get()
	assert.Contains(t, []pb.LoginType{pb.LoginType_LoginType_UNKNOWN, pb.LoginType_LoginType_LOGIN, pb.LoginType_LoginType_SIGNUP}, finalValue)
}
