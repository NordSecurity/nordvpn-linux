package daemon

import (
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

type AtomicLoginType struct {
	value atomic.Int32
}

func NewAtomicLoginType(initialValue ...pb.LoginType) *AtomicLoginType {
	alt := &AtomicLoginType{}
	if len(initialValue) > 0 {
		alt.Set(initialValue[0])
	}
	return alt
}

func (a *AtomicLoginType) Get() pb.LoginType {
	return pb.LoginType(a.value.Load())
}

func (a *AtomicLoginType) Set(loginType pb.LoginType) {
	a.value.Store(int32(loginType))
}

func (a *AtomicLoginType) Reset() {
	a.Set(pb.LoginType_LoginType_UNKNOWN)
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
