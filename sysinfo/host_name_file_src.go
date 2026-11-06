package sysinfo

import (
	"bytes"
	"fmt"
	"os"
)

const etcOSReleaseFile = "/etc/os-release"

type osRelease []byte

func readOSRelease() (osRelease, error) {
	return os.ReadFile(etcOSReleaseFile)
}

func (o osRelease) tag(tag string) (string, error) {
	for line := range bytes.SplitSeq(o, []byte("\n")) {
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
