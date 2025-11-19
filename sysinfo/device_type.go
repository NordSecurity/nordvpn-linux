package sysinfo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type SystemDeviceType string

const (
	SystemDeviceTypeUnknown   SystemDeviceType = "unknown"
	SystemDeviceTypeDesktop   SystemDeviceType = "desktop"
	SystemDeviceTypeServer    SystemDeviceType = "server"
	SystemDeviceTypeContainer SystemDeviceType = "container"
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
		newContainerDetector(),
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

// -------------------------------------
// Container Detector
// -------------------------------------

type containerDetector struct {
	checkIfFileExists    func(string) bool
	readEnv              func(string) string
	readEnvironAndCgroup func() ([]byte, []byte)
}

func (d containerDetector) Get() (SystemDeviceType, error) {
	if d.isDocker() || d.isKubernetes() || d.isLXC() {
		return SystemDeviceTypeContainer, nil
	}

	return SystemDeviceTypeUnknown, nil
}

func newContainerDetector() deviceTypeDetector {
	return &containerDetector{
		checkIfFileExists:    internal.FileExists,
		readEnv:              os.Getenv,
		readEnvironAndCgroup: readEnvironAndCgroup,
	}
}

func readEnvironAndCgroup() ([]byte, []byte) {
	environ, err := os.ReadFile("/proc/1/environ")
	if err != nil {
		return []byte{}, []byte{}
	}

	cgroup, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return environ, []byte{}
	}

	return environ, cgroup
}

// isDocker detects docker environment looking for /.dockerenv file
func (d containerDetector) isDocker() bool {
	// If /.dockerenv exists we are running inside docker
	return d.checkIfFileExists("/.dockerenv")
}

// isKubernetes detects Kubernetes environment looking for KUBERNETES_SERVICE_HOST env
func (d containerDetector) isKubernetes() bool {
	return d.readEnv("KUBERNETES_SERVICE_HOST") != ""
}

// isLXC detects traditional LXC or LXD containers via cgroup or environment variables.
func (d containerDetector) isLXC() bool {
	// Some LXC containers may have container=lxc env variable
	if val := d.readEnv("container"); strings.ToLower(val) == "lxc" {
		return true
	}

	// Check /proc/1/environ for container=lxc or container=lxd and /proc/1/cgroup for lxc or lxd
	environ, cgroup := d.readEnvironAndCgroup()
	if bytes.Contains(environ, []byte("container=lxc")) || bytes.Contains(environ, []byte("container=lxd")) ||
		bytes.Contains(cgroup, []byte("lxc")) || bytes.Contains(cgroup, []byte("lxd")) {
		return true
	}

	return false
}
