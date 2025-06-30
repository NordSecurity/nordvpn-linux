package sysinfo

import (
	"os"
	"strings"
)

// getDesktopEnvironment retrieves currently used desktop environment type otherwise 'none'
func getDesktopEnvironment(readEnv envReader) string {
	de := readEnv("XDG_CURRENT_DESKTOP")
	if de == "" {
		de = readEnv("DESKTOP_SESSION")
	}

	de = strings.TrimSpace(de)
	if de == "" {
		return EnvValueUnset
	}

	if strings.Contains(de, ":") {
		tokens := strings.Split(de, ":")
		de = tokens[1]
	}

	return strings.ToLower(de)
}

// GetDisplayDesktopEnvironment retrieves the current desktop environment.
// This function only works in user sessions where the environment is populated with XDG info.
func GetDisplayDesktopEnvironment() string {
	return getDesktopEnvironment(os.Getenv)
}
