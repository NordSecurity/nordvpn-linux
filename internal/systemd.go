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
	return FileExists("/run/nologin") || FileExists("/var/run/nologin")
}

// IsSystemd detect if system is running systemd
func IsSystemd() bool {
	// check name of PID1 process: "ps -p 1 -o comm="
	out, err := exec.Command("ps", "-p", "1", "-o", "comm=").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "systemd")
}
