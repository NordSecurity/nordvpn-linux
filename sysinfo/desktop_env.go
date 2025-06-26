package sysinfo

import "strings"

// getDesktopEnvironment retrieves currently used desktop environment type otherwise 'none'
func getDesktopEnvironment(readEnv envReader) string {
	de := readEnv("XDG_CURRENT_DESKTOP")
	if de == "" {
		de = readEnv("DESKTOP_SESSION")
	}

	de = strings.TrimSpace(de)
	if de == "" {
		return "none"
	}

	if strings.Contains(de, ":") {
		tokens := strings.Split(de, ":")
		de = tokens[1]
	}

	return strings.ToLower(de)
}
