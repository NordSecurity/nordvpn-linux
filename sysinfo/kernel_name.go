package sysinfo

import (
	"log"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// defaultKernelName defines default kernel name
const defaultKernelName = "Linux"

// cmdRunner defines a function type for executing shell commands.
type cmdRunner func(name string, args ...string) (string, error)

// defaultCmdRunner executes a system command and returns its output.
func defaultCmdRunner(name string, args ...string) (string, error) {
	// #nosec G204 -- input is known before running the program
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// uname retrieves OS information using the provided command runner.
func uname(runner cmdRunner, flags string) string {
	out, err := runner("uname", flags)
	if err != nil {
		log.Printf("%s failed to execute 'uname %s': %v. Falling back to default: %s\n",
			internal.ErrorPrefix, flags, err, defaultKernelName)
		return defaultKernelName
	}

	if out == "" {
		log.Printf("%s 'uname %s' returned empty output. Falling back to default: %s\n",
			internal.ErrorPrefix, flags, defaultKernelName)
		return defaultKernelName
	}

	return out
}

// GetKernelVersion retrieves the name and release version of the currently running kernel.
func GetKernelVersion() string {
	return uname(defaultCmdRunner, "-sr")
}
