package mock

type RemoteConfigMock struct {
	NordWhisperEnabled bool
	GetNordWhisperErr  error
}

func NewRemoteConfigMock() *RemoteConfigMock {
	return &RemoteConfigMock{}
}

func (r *RemoteConfigMock) GetTelioConfig(version string) (string, error) {
	return "", nil
}

func (r *RemoteConfigMock) GetNordWhisperEnabled(version string) (bool, error) {
	return r.NordWhisperEnabled, r.GetNordWhisperErr
}
