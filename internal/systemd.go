package internal

import (
	"os/exec"
	"strings"
)

// IsSystemShutdown tries to determine if systemd shutdown or reboot is being executed
func IsSystemShutdown() bool {
	return FileExists("/run/nologin") || FileExists("/var/run/nologin")
}

// IsServiceActive check if given service is active
func IsServiceActive(service string) bool {
	out, err := exec.Command(SystemctlExec, "is-active", service).Output()
	if err != nil {
		return false
	}
	return "active" == strings.Trim(strings.Trim(string(out), "\n"), " ")
}
