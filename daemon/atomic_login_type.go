package daemon

import (
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// atomicLoginType manages login type and timing information with thread-safe access.
// It uses a single RWMutex for synchronization of both the login type value and timestamp.
type atomicLoginType struct {
	loginType            pb.LoginType
	lastLoginAttemptTime time.Time
	mu                   sync.RWMutex
}

func NewAtomicLoginType() *atomicLoginType {
	return &atomicLoginType{
		loginType: pb.LoginType_LoginType_UNKNOWN,
	}
}

func (a *atomicLoginType) get() pb.LoginType {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.loginType
}

func (a *atomicLoginType) set(loginType pb.LoginType) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loginType = loginType
}

func (a *atomicLoginType) reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loginType = pb.LoginType_LoginType_UNKNOWN
	a.lastLoginAttemptTime = time.Time{}
}

func (a *atomicLoginType) isUnknown() bool {
	return a.get() == pb.LoginType_LoginType_UNKNOWN
}

func (a *atomicLoginType) wasStarted() bool {
	return !a.isUnknown()
}

func (a *atomicLoginType) isAltered(loginType pb.LoginType) bool {
	return a.get() != loginType
}

func (a *atomicLoginType) setLoginAttemptTime(t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastLoginAttemptTime = t
}

func (a *atomicLoginType) getLoginAttemptTime() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastLoginAttemptTime
}
