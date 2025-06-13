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

// GetDeviceType attempts to determine whether the machine is a desktop or server.
// It sequentially evaluates a series of detection strategies:
// - systemd default target
// - presence of graphical environment paths
// - XDG session type
//
// Returns the first non-unknown DeviceType detected, or DeviceTypeUnknown as a fallback.
func GetDeviceType() DeviceType {

	sources := []func() DeviceType{
		detectBySystemDefaultTarget,
		detectByGraphicalEnv,
		detectByXDGSession,
	}

	for _, s := range sources {
		if dt := s(); dt != DeviceTypeUnknown {
			return dt
		}
	}

	// fallback type if none of the methods could determine it
	return DeviceTypeUnknown
}
