package sysinfo

type envReader func(string) string

const (
	// EnvValueUnset represents an unset or missing environment variable value.
	EnvValueUnset = "none"
	logTag        = "[sysinfo]"
)
