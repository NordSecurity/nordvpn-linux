package vpn

// LibConfigGetter is interface to acquire config for vpn implementation library
type LibConfigGetter interface {
	GetConfig() (string, error)
}

// NordWhisperFeatureConfig defines the features available for NordWhisper
type NordWhisperFeatureConfig struct {
	EnableECH bool
}

// NewNordWhisperFeatureConfig builds the default feature configuration for NordWhisper
func NewNordWhisperFeatureConfig() NordWhisperFeatureConfig {
	return NordWhisperFeatureConfig{
		EnableECH: true,
	}
}

// NordWhisperConfigGetter is interface to acquire config for NordWhisper vpn implementation
type NordWhisperConfigGetter interface {
	GetConfig() (NordWhisperFeatureConfig, error)
}
