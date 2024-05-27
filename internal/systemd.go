package internal

import (
	"os/exec"
	"strings"
)

// IsServiceActive check if given service is active
func IsServiceActive(service string) bool {
	out, err := exec.Command(SystemctlExec, "is-active", service).Output()
	if err != nil {
		return false
	}
	return "active" == strings.Trim(strings.Trim(string(out), "\n"), " ")
}

// IsSystemShutdown detect if system is being shutdown
func IsSystemShutdown() bool {
	// https://www.freedesktop.org/software/systemd/man/latest/shutdown.html
	return FileExists("/run/nologin")
}

// IsSystemd detect if system is running systemd
func IsSystemd() bool {
	// https://www.freedesktop.org/software/systemd/man/latest/sd_booted.html
	return FileExists("/run/systemd/system")
}
