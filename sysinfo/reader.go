package sysinfo

import "os"

// GetDesktopEnvironment retrieves the current desktop environment.
// This function only works in user sessions where the environment is populated with XDG info.
func GetDesktopEnvironment() string {
	return getDesktopEnvironment(os.Getenv)
}

// GetDisplayProtocol retrieves the current display protocol.
// This function only works in user sessions where the environment is populated with XDG info.
func GetDisplayProtocol() string {
	return getDisplayProtocol(os.Getenv)
}
