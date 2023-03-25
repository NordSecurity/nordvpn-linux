/*
Package distro provides information about the current Linux distribution.
*/
package distro

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// kernelName defines default kernel name
	kernelName    = "Linux"
	osReleaseFile = "/etc/os-release"
)

// osRelease represents contents of /etc/os-release
type osRelease struct {
	Name       string
	PrettyName string
}

func (o *osRelease) UnmarshalText(text []byte) error {
	for _, line := range bytes.Split(bytes.TrimSpace(text), []byte("\n")) {
		key, value, ok := bytes.Cut(line, []byte("="))
		if !ok {
			continue
		}

		isDoubleQuote := func(r rune) bool { return r == '"' }
		switch string(key) {
		case "NAME":
			o.Name = string(bytes.TrimFunc(value, isDoubleQuote))
		case "PRETTY_NAME":
			o.PrettyName = string(bytes.TrimFunc(value, isDoubleQuote))
		default:
			// ignore undefined fields
		}
	}
	return nil
}

// ReleaseName of the currently running distribution.
func ReleaseName() (string, error) {
	data, err := os.ReadFile(osReleaseFile)
	if err != nil {
		return "", err
	}

	var release osRelease
	if err := (&release).UnmarshalText(data); err != nil {
		return "", err
	}

	return release.Name, nil
}

// ReleasePrettyName of the currently running distribution.
func ReleasePrettyName() (string, error) {
	data, err := os.ReadFile(osReleaseFile)
	if err != nil {
		return "", err
	}

	var release osRelease
	if err := (&release).UnmarshalText(data); err != nil {
		return "", err
	}

	return release.PrettyName, nil
}

// KernelName of the currently running kernel.
func KernelName() string { return uname("-sr") }

// KernelFull name of the currently running kernel.
func KernelFull() string { return uname("-a") }

// uname returns operating system information from uname executable
func uname(flags string) string {
	// #nosec G204 -- input is known before running the program
	out, _ := exec.Command("sh", "-c", fmt.Sprintf("uname %s", flags)).Output()
	trimmed := strings.Trim(string(out), "\n")
	if trimmed == "" {
		return kernelName
	}

	return trimmed
}
