package daemon

import (
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// AtomicLoginType manages login type and timing information with thread-safe access.
// It uses a single RWMutex for synchronization of both the login type value and timestamp.
type AtomicLoginType struct {
	loginType            pb.LoginType
	lastLoginAttemptTime time.Time
	mu                   sync.RWMutex
}

func NewAtomicLoginType() *AtomicLoginType {
	return &AtomicLoginType{
		loginType: pb.LoginType_LoginType_UNKNOWN,
	}
}

func (a *AtomicLoginType) Get() pb.LoginType {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.loginType
}

func (a *AtomicLoginType) Set(loginType pb.LoginType) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loginType = loginType
}

func (a *AtomicLoginType) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loginType = pb.LoginType_LoginType_UNKNOWN
	a.lastLoginAttemptTime = time.Time{}
}

func (a *AtomicLoginType) IsUnknown() bool {
	return a.Get() == pb.LoginType_LoginType_UNKNOWN
}

func (a *AtomicLoginType) WasStarted() bool {
	return !a.IsUnknown()
}

func (a *AtomicLoginType) IsAltered(loginType pb.LoginType) bool {
	return a.Get() != loginType
}

func (a *AtomicLoginType) SetLoginAttemptTime(t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastLoginAttemptTime = t
}

func (a *AtomicLoginType) GetLoginAttemptTime() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastLoginAttemptTime
}
