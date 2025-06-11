package sysinfo

import (
	"os"
	"os/exec"
	"strings"
)

type DeviceType string

const (
	DeviceTypeUnknown DeviceType = "unknown"
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeServer  DeviceType = "server"
)

// detectBySystemDefaultTarget determines the device type based on the default systemd target.
func detectBySystemDefaultTarget() DeviceType {
	_, err := exec.LookPath("systemctl")
	if err != nil {
		return DeviceTypeUnknown
	}

	out, err := exec.Command("systemctl", "get-default").Output()
	if err != nil {
		return DeviceTypeUnknown
	}

	switch strings.TrimSpace(string(out)) {
	case "graphical.target":
		return DeviceTypeDesktop
	case "multi-user.target":
		return DeviceTypeServer
	}

	return DeviceTypeUnknown
}

// detectByGraphicalEnv checks for the presence of common GUI-related system directories.
// Returns DeviceTypeDesktop if any are found, otherwise returns DeviceTypeUnknown.
func detectByGraphicalEnv() DeviceType {
	de := getDesktopEnvironment(os.Getenv)
	if de != "none" {
		return DeviceTypeDesktop
	}

	paths := []string{
		"/etc/X11",
		"/usr/share/xsessions",
		"/usr/share/wayland-sessions",
	}

	for _, path := range paths {
		if fi, err := os.Stat(path); err == nil && fi.IsDir() {
			return DeviceTypeDesktop
		}
	}

	return DeviceTypeUnknown
}

// detectByXDGSession evaluates the device type from the XDG_SESSION_TYPE environment variable.
// Returns DeviceTypeDesktop for "x11" or "wayland", DeviceTypeServer for "tty",
// and DeviceTypeUnknown for any other or unset value.
// Works when calling from user-level environment or its environment is propagated to the daemon
func detectByXDGSession() DeviceType {
	sessionType := getDisplayProtocol(os.Getenv)
	switch sessionType {
	case "x11", "wayland":
		return DeviceTypeDesktop
	case "tty":
		return DeviceTypeServer
	default:
		return DeviceTypeUnknown
	}
}
