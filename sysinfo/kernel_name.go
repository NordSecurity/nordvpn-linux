package sysinfo

import (
	"os/exec"
	"strings"
)

// kernelName defines default kernel name
const kernelName = "Linux"

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
	if err != nil || out == "" {
		return kernelName
	}

	return out
}

// KernelVersion retrieves the name and release version of the currently running kernel.
func KernelVersion() string {
	return uname(defaultCmdRunner, "-sr")
}
