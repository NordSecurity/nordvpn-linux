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
		setup func(*AtomicLoginType)
		check func(*testing.T, *AtomicLoginType)
	}{
		{
			name:  "initial state is unknown",
			setup: func(alt *AtomicLoginType) {},
			check: func(t *testing.T, alt *AtomicLoginType) {
				assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, alt.Get())
				assert.True(t, alt.IsUnknown())
				assert.False(t, alt.WasStarted())
			},
		},
		{
			name: "reset to unknown",
			setup: func(alt *AtomicLoginType) {
				alt.Set(pb.LoginType_LoginType_LOGIN)
				alt.Reset()
			},
			check: func(t *testing.T, alt *AtomicLoginType) {
				assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, alt.Get())
				assert.True(t, alt.IsUnknown())
				assert.False(t, alt.WasStarted())
			},
		},
		{
			name: "IsAltered returns true when types differ",
			setup: func(alt *AtomicLoginType) {
				alt.Set(pb.LoginType_LoginType_LOGIN)
			},
			check: func(t *testing.T, alt *AtomicLoginType) {
				assert.True(t, alt.IsAltered(pb.LoginType_LoginType_SIGNUP))
				assert.True(t, alt.IsAltered(pb.LoginType_LoginType_UNKNOWN))
				assert.False(t, alt.IsAltered(pb.LoginType_LoginType_LOGIN))
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

func TestAtomicLoginTypeConstructorWithInitialValue(t *testing.T) {
	tests := []struct {
		name          string
		initialValues []pb.LoginType
		expectedValue pb.LoginType
	}{
		{
			name:          "no initial value defaults to UNKNOWN",
			initialValues: []pb.LoginType{},
			expectedValue: pb.LoginType_LoginType_UNKNOWN,
		},
		{
			name:          "initial value LOGIN",
			initialValues: []pb.LoginType{pb.LoginType_LoginType_LOGIN},
			expectedValue: pb.LoginType_LoginType_LOGIN,
		},
		{
			name:          "initial value SIGNUP",
			initialValues: []pb.LoginType{pb.LoginType_LoginType_SIGNUP},
			expectedValue: pb.LoginType_LoginType_SIGNUP,
		},
		{
			name:          "multiple values uses first one",
			initialValues: []pb.LoginType{pb.LoginType_LoginType_LOGIN, pb.LoginType_LoginType_SIGNUP},
			expectedValue: pb.LoginType_LoginType_LOGIN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alt := NewAtomicLoginType(tt.initialValues...)
			assert.Equal(t, tt.expectedValue, alt.Get())
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
					alt.Set(pb.LoginType_LoginType_LOGIN)
				} else {
					alt.Set(pb.LoginType_LoginType_SIGNUP)
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
				_ = alt.Get()
				_ = alt.IsUnknown()
				_ = alt.WasStarted()
			}
		}()
	}

	// Reseters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations/10; j++ {
				alt.Reset()
			}
		}()
	}

	wg.Wait()

	// Final state should be valid
	finalValue := alt.Get()
	assert.Contains(t, []pb.LoginType{pb.LoginType_LoginType_UNKNOWN, pb.LoginType_LoginType_LOGIN, pb.LoginType_LoginType_SIGNUP}, finalValue)
}
