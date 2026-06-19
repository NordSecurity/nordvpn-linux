package mock

type RemoteConfigMock struct {
	NordWhisperEnabled bool
	GetNordWhisperErr  error
	FeatureToggles     map[string]bool
}

func NewRemoteConfigMock() *RemoteConfigMock {
	return &RemoteConfigMock{
		FeatureToggles: make(map[string]bool),
	}
}

func (r *RemoteConfigMock) AddFeatureToggle(featureName string, toggle bool) {
	r.FeatureToggles[featureName] = toggle
}

func (r *RemoteConfigMock) GetTelioConfig() (string, error) {
	return "", nil
}

func (r *RemoteConfigMock) GetNordWhisperEnabled() (bool, error) {
	return r.NordWhisperEnabled, r.GetNordWhisperErr
}

func (r *RemoteConfigMock) IsFeatureEnabled(featureName string) bool {
	if toggle, exists := r.FeatureToggles[featureName]; exists {
		return toggle
	}
	return false
}

func (r *RemoteConfigMock) GetFeatureParam(_, _ string) (string, error) { return "", nil }
func (r *RemoteConfigMock) LoadConfig() error                           { return nil }
