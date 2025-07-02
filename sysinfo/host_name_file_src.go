package sysinfo

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const etcOSReleaseFile = "/etc/os-release"

// GetHostOSName retrieves the standard name of the currently running operating system.
func GetHostOSName() (string, error) {
	return readOSReleaseTag("NAME")
}

// readOSReleaseTag opens the 'etcOSReleaseFile' file and retrieves the specified tag.
func readOSReleaseTag(tag string) (string, error) {
	file, err := os.Open(etcOSReleaseFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return readTagFromOSRelease(file, tag)
}

func readTagFromOSRelease(r io.Reader, tag string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	for _, line := range bytes.Split(data, []byte("\n")) {
		key, value, ok := bytes.Cut(line, []byte("="))
		if !ok {
			continue
		}

		if string(key) == tag {
			return string(bytes.Trim(value, "\"")), nil
		}
	}

	return "", fmt.Errorf("unsupported tag: %v", tag)
}
