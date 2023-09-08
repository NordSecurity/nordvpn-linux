package mock

type SysctlSetterMock struct {
	isSet bool
}

func (s *SysctlSetterMock) Set() error {
	s.isSet = true
	return nil
}

func (s *SysctlSetterMock) Unset() error {
	s.isSet = false
	return nil
}
