package mock

type RemoteConfigMock struct {
	QuenchEnabled bool
	GetQuenchErr  error
}

func NewRemoteConfigMock() *RemoteConfigMock {
	return &RemoteConfigMock{}
}

func (r *RemoteConfigMock) GetTelioConfig(version string) (string, error) {
	return "", nil
}

func (r *RemoteConfigMock) GetQuenchEnabled(version string) (bool, error) {
	return r.QuenchEnabled, r.GetQuenchErr
}
