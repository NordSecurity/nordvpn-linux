package mock

type RemoteConfigMock struct {
	NordWhisperEnabled bool
	GetNordWhisperErr  error
}

func NewRemoteConfigMock() *RemoteConfigMock {
	return &RemoteConfigMock{}
}

func (r *RemoteConfigMock) GetTelioConfig() (string, error) {
	return "", nil
}

func (r *RemoteConfigMock) GetNordWhisperEnabled() (bool, error) {
	return r.NordWhisperEnabled, r.GetNordWhisperErr
}

func (r *RemoteConfigMock) IsFeatureEnabled(string) bool { return false }
func (r *RemoteConfigMock) GetFeatureParam(_, _ string)  {}
func (r *RemoteConfigMock) LoadConfig() error            { return nil }
