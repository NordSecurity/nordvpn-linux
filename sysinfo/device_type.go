package sysinfo

import (
	"os"
	"os/exec"
	"strings"
)

type SystemDeviceType string

const (
	SystemDeviceTypeUnknown SystemDeviceType = "unknown"
	SystemDeviceTypeDesktop SystemDeviceType = "desktop"
	SystemDeviceTypeServer  SystemDeviceType = "server"
)

// detectBySystemDefaultTarget determines the device type based on the default systemd target.
func detectBySystemDefaultTarget() SystemDeviceType {
	_, err := exec.LookPath("systemctl")
	if err != nil {
		return SystemDeviceTypeUnknown
	}

	out, err := exec.Command("systemctl", "get-default").Output()
	if err != nil {
		return SystemDeviceTypeUnknown
	}

	switch strings.TrimSpace(string(out)) {
	case "graphical.target":
		return SystemDeviceTypeDesktop
	case "multi-user.target":
		return SystemDeviceTypeServer
	}

	return SystemDeviceTypeUnknown
}

// detectByGraphicalEnv checks for the presence of common GUI-related system directories.
// Returns SystemDeviceTypeDesktop if any are found, otherwise returns SystemDeviceTypeUnknown.
func detectByGraphicalEnv() SystemDeviceType {
	de := getDesktopEnvironment(os.Getenv)
	if de != EnvValueUnset {
		return SystemDeviceTypeDesktop
	}

	paths := []string{
		"/etc/X11",
		"/usr/share/xsessions",
		"/usr/share/wayland-sessions",
	}

	for _, path := range paths {
		if fi, err := os.Stat(path); err == nil && fi.IsDir() {
			return SystemDeviceTypeDesktop
		}
	}

	return SystemDeviceTypeUnknown
}

// detectByXDGSession evaluates the device type from the XDG_SESSION_TYPE environment variable.
// Returns SystemDeviceTypeDesktop for "x11" or "wayland", SystemDeviceTypeServer for "tty",
// and SystemDeviceTypeUnknown for any other or unset value.
// Works when calling from user-level environment or its environment is propagated to the daemon
func detectByXDGSession() SystemDeviceType {
	sessionType := getDisplayProtocol(os.Getenv)
	switch sessionType {
	case "x11", "wayland":
		return SystemDeviceTypeDesktop
	case "tty":
		return SystemDeviceTypeServer
	default:
		return SystemDeviceTypeUnknown
	}
}

// SystemDeviceType attempts to determine whether the machine is a desktop or server.
// It sequentially evaluates a series of detection strategies:
// - systemd default target
// - presence of graphical environment paths
// - XDG session type
//
// Returns the first non-unknown SystemDeviceType detected, or SystemDeviceTypeUnknown as a fallback.
func DeviceType() SystemDeviceType {
	sources := []func() SystemDeviceType{
		detectBySystemDefaultTarget,
		detectByGraphicalEnv,
		detectByXDGSession,
	}

	for _, s := range sources {
		if dt := s(); dt != SystemDeviceTypeUnknown {
			return dt
		}
	}

	// fallback type if none of the methods could determine it
	return SystemDeviceTypeUnknown
}
