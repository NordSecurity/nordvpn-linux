package sysinfo

import (
	"os"
	"strings"
)

// getDisplayServer retrieves currently used display protocol otherwise none
func getDisplayProtocol(readEnv envReader) string {
	ds := readEnv("XDG_SESSION_TYPE")
	ds = strings.TrimSpace(ds)
	if ds == "" {
		return EnvValueUnset
	}

	return strings.ToLower(ds)
}

// GetDisplayProtocol retrieves the current display protocol.
// This function only works in user sessions where the environment is populated with XDG info.
func GetDisplayProtocol() string {
	return getDisplayProtocol(os.Getenv)
}
