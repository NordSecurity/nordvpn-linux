package sysinfo

import "strings"

// getDisplayServer retrieves currently used display protocol otherwise none
func getDisplayProtocol(readEnv envReader) string {
	ds := readEnv("XDG_SESSION_TYPE")
	ds = strings.TrimSpace(ds)
	if ds == "" {
		return "none"
	}

	return strings.ToLower(ds)
}
