package kernel

import "fmt"

type SysctlSetter struct {
	paramName     string
	desiredValue  int
	unwantedValue int
	changed       bool
}

func NewSysctlSetter(
	paramName string,
	desiredValue int,
	unwantedValue int,
) *SysctlSetter {
	return &SysctlSetter{
		paramName:     paramName,
		desiredValue:  desiredValue,
		unwantedValue: unwantedValue,
		changed:       false,
	}
}

func (s *SysctlSetter) Set() error {
	values, err := Parameter(s.paramName)
	if err != nil {
		return fmt.Errorf(
			"retrieving the value of '%s': %w",
			s.paramName,
			err,
		)
	}
	if values[s.paramName] == s.unwantedValue {
		err := SetParameter(s.paramName, s.desiredValue)
		if err != nil {
			return fmt.Errorf(
				"setting the value of '%s' to %d: %w",
				s.paramName,
				s.desiredValue,
				err,
			)
		}
		s.changed = true
	}
	return nil
}

func (s *SysctlSetter) Unset() error {
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
