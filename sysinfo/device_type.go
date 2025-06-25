package sysinfo

import (
	"fmt"
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

const (
	defaultTargetGraphical = "graphical.target"
	defaultTargetServer    = "multi-user.target"
	sessionTypeX11         = "x11"
	sessionTypeWayland     = "wayland"
	sessionTypeTTY         = "tty"
)

type deviceTypeDetector interface {
	Get() (SystemDeviceType, error)
}

// GetDeviceType returns the system's device type by checking systemd target,
// desktop environment presence, and session type, in that order.
func GetDeviceType() SystemDeviceType {
	detectors := []deviceTypeDetector{
		newSystemdTargetDetector(),
		newGraphicalEnvDetector(),
		newXDGSessionDetector(),
	}

	for _, detector := range detectors {
		if deviceType, err := detector.Get(); err == nil && deviceType != SystemDeviceTypeUnknown {
			return deviceType
		}
	}

	return SystemDeviceTypeUnknown
}

// -------------------------------------
// Systemd Target Detector
// -------------------------------------

type systemdTargetDetector struct {
	detectTarget func() (string, error)
}

func (d systemdTargetDetector) Get() (SystemDeviceType, error) {
	target, err := d.detectTarget()
	if err != nil {
		return SystemDeviceTypeUnknown, fmt.Errorf("detecting systemd target: %w", err)
	}

	switch strings.TrimSpace(target) {
	case defaultTargetGraphical:
		return SystemDeviceTypeDesktop, nil
	case defaultTargetServer:
		return SystemDeviceTypeServer, nil
	}

	return SystemDeviceTypeUnknown, nil
}

func newSystemdTargetDetector() deviceTypeDetector {
	return &systemdTargetDetector{
		detectTarget: func() (string, error) {
			if _, err := exec.LookPath("systemctl"); err != nil {
				return "", err
			}
			out, err := exec.Command("systemctl", "get-default").Output()
			return string(out), err
		},
	}
}

// -------------------------------------
// Graphical Environment Detector
// -------------------------------------

type fileInfoFunc func(name string) (os.FileInfo, error)

type graphicalEnvDetector struct {
	detectEnv func() (string, error)
	statPath  fileInfoFunc
}

func (d graphicalEnvDetector) Get() (SystemDeviceType, error) {
	env, err := d.detectEnv()
	if err != nil {
		return SystemDeviceTypeUnknown, fmt.Errorf("detecting graphical env: %w", err)
	}
	if env != EnvValueUnset {
		return SystemDeviceTypeDesktop, nil
	}

	guiPaths := []string{
		"/etc/X11",
		"/usr/share/xsessions",
		"/usr/share/wayland-sessions",
	}
	for _, path := range guiPaths {
		if info, err := d.statPath(path); err == nil && info.IsDir() {
			return SystemDeviceTypeDesktop, nil
		}
	}

	return SystemDeviceTypeUnknown, nil
}

func newGraphicalEnvDetector() deviceTypeDetector {
	return &graphicalEnvDetector{
		detectEnv: func() (string, error) {
			return getDesktopEnvironment(os.Getenv), nil
		},
		statPath: os.Stat,
	}
}

// -------------------------------------
// XDG Session Detector
// -------------------------------------

type xdgSessionDetector struct {
	detectSession func() (string, error)
}

func (d xdgSessionDetector) Get() (SystemDeviceType, error) {
	session, err := d.detectSession()
	if err != nil {
		return SystemDeviceTypeUnknown, fmt.Errorf("detecting XDG session: %w", err)
	}

	switch session {
	case sessionTypeX11, sessionTypeWayland:
		return SystemDeviceTypeDesktop, nil
	case sessionTypeTTY:
		return SystemDeviceTypeServer, nil
	}

	return SystemDeviceTypeUnknown, nil
}

func newXDGSessionDetector() deviceTypeDetector {
	return &xdgSessionDetector{
		detectSession: func() (string, error) {
			return getDisplayProtocol(os.Getenv), nil
		},
	}
}
