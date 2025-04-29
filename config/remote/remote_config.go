package remote

// RemoteConfigGetter get values from remote config
type RemoteConfigGetter interface {
	GetTelioConfig() (string, error)
}
