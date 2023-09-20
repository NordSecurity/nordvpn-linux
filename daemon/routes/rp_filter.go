package routes

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/kernel"
)

// RPFilterManager should handle the setting and unsetting of the
// desired RP filter configuration value
type RPFilterManager interface {
	// Set sets the RP filter value to the one which allows policy
	// based routing if necessary
	Set() error
	// Unset sets the RP filter value to the one which was set
	// before
	Unset() error
}

type SysctlRPFilterManager struct {
	setter *kernel.SysctlSetterImpl
}

func NewSysctlRPFilterManager() *SysctlRPFilterManager {
	return &SysctlRPFilterManager{
		setter: kernel.NewSysctlSetter("net.ipv4.conf.all.rp_filter", 2, 1),
	}
}

func (s *SysctlRPFilterManager) Set() error {
	if err := s.setter.Set(); err != nil {
		return fmt.Errorf("setting rp_filter: %w", err)
	}
	return nil
}

func (s *SysctlRPFilterManager) Unset() error {
	if err := s.setter.Unset(); err != nil {
		return fmt.Errorf("unsetting rp_filter: %w", err)
	}
	return nil
}
