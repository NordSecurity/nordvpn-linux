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

type Distro interface {
	// ReleaseName returns the name of the currently running distribution.
	ReleaseName() (string, error)
	// ReleasePrettyName returns the pretty name of the currently running distribution.
	ReleasePrettyName() (string, error)
	// KernelName returns just the name of the currently running kernel.
	KernelName() string
	// KernelFull returns full set of information about the currently running kernel.
	KernelFull() string
}

type DistroImpl struct{}

func NewDistro() Distro {
	return &DistroImpl{}
}

func (DistroImpl) ReleaseName() (string, error) {
	data, err := os.ReadFile(osReleaseFile)
	if err != nil {
		return "", err
	}

	var release osRelease
	release.unmarshalText(data)
	return release.Name, nil
}

func (DistroImpl) ReleasePrettyName() (string, error) {
	data, err := os.ReadFile(osReleaseFile)
	if err != nil {
		return "", err
	}

	var release osRelease
	release.unmarshalText(data)
	return release.PrettyName, nil
}

func (DistroImpl) KernelName() string { return uname("-sr") }

func (DistroImpl) KernelFull() string { return uname("-a") }

func (o *osRelease) unmarshalText(text []byte) {
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
}

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
