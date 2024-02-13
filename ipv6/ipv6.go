// Package ipv6 provides toggles for IPv6 part of the TCP/IP stack.
package ipv6

import (
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/kernel"
)

// Blocker provides idempotent blocking and unblocking.
type Blocker interface {
	Block() error
	Unblock() error
}

type Ipv6 struct {
	sysctlSetter *kernel.SysctlSetterImpl
	sync.Mutex
}

const netIPv6KernelParameterName = "net.ipv6.conf.all.disable_ipv6"

func NewIpv6() *Ipv6 {
	return &Ipv6{
		sysctlSetter: kernel.NewSysctlSetter(netIPv6KernelParameterName, 1, 0),
	}
}

// Block ipv6 and backup previous settings if there is no backup.
func (i *Ipv6) Block() error {
	i.Lock()
	defer i.Unlock()
	if i.sysctlSetter.IsEnabled() {
		return i.sysctlSetter.Set()
	}
	log.Println(internal.InfoPrefix, "IPv6 module is not enabled")
	return nil
}

// Unblock Ipv6 and restore previous settings from backup.
func (i *Ipv6) Unblock() error {
	i.Lock()
	defer i.Unlock()
	if i.sysctlSetter.IsEnabled() {
		return i.sysctlSetter.Unset()
	}
	return nil
}
