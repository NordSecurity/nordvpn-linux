package mock

type SysctlSetterMock struct {
	IsSet    bool
	SetErr   error
	UnsetErr error
}

func (s *SysctlSetterMock) Set() error {
	if s.SetErr != nil {
		return s.SetErr
	}
	s.IsSet = true
	return nil
}

func (s *SysctlSetterMock) Unset() error {
	if s.UnsetErr != nil {
		return s.UnsetErr
	}
	s.IsSet = false
	return nil
}
