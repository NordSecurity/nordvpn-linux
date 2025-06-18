package sysinfo

type envReader func(string) string

// EnvValueUnset represents an unset or missing environment variable value.
const EnvValueUnset = "none"
