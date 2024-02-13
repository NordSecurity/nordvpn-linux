package kernel

import (
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type SysctlSetter interface {
	Set() error
	Unset() error
}

type SysctlSetterImpl struct {
	paramName     string
	desiredValue  int
	unwantedValue int
	changed       bool
}

func NewSysctlSetter(
	paramName string,
	desiredValue int,
	unwantedValue int,
) *SysctlSetterImpl {
	return &SysctlSetterImpl{
		paramName:     paramName,
		desiredValue:  desiredValue,
		unwantedValue: unwantedValue,
		changed:       false,
	}
}

func (s *SysctlSetterImpl) Set() error {
	if !s.changed {
		values, err := Parameter(s.paramName)
		if err != nil {
			return fmt.Errorf(
				"retrieving the value of '%s': %w",
				s.paramName,
				err,
			)
		}

		if values[s.paramName] != s.desiredValue {
			s.changed = true
		}
	}

	// always set the new value, even if the values is already set
	// otherwise when a new USB adapter is connected, even if net.ipv6.conf.all.disable_ipv6=1 the new adaptor will have IPv6
	// so it needs to be set again, to disabled IPv6 also for the new interface
	if err := SetParameter(s.paramName, s.desiredValue); err != nil {
		return fmt.Errorf(
			"setting the value of '%s' to %d: %w",
			s.paramName,
			s.desiredValue,
			err,
		)
	}

	return nil
}

func (s *SysctlSetterImpl) Unset() error {
	if s.changed {
		err := SetParameter(s.paramName, s.unwantedValue)
		if err != nil {
			return fmt.Errorf(
				"setting the value of '%s' to '%d': %w",
				s.paramName,
				s.unwantedValue,
				err,
			)
		}
		s.changed = false
	}
	return nil
}

func (s *SysctlSetterImpl) Exists() bool {
	if !internal.FileExists("/proc/sys/" + strings.ReplaceAll(s.paramName, ".", "/")) {
		return false
	}

	if _, err := Parameter(s.paramName); err != nil {
		return strings.Contains(err.Error(), "cannot stat")
	}

	return true
}
