package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
)

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../.."))
}

func TestValidateServiceFiles(t *testing.T) {
	category.Set(t, category.Unit)

	systemdAnalyze, err := exec.LookPath("systemd-analyze")
	if err != nil {
		t.Skip("systemd-analyze not found; skipping systemd unit validation")
	}

	serviceFiles := []string{
		"contrib/systemd/system/nordvpnd.service",
		"contrib/systemd/system/nordvpnd.socket",
	}

	for _, serviceFile := range serviceFiles {
		serviceFile = projectRoot() + "/" + serviceFile
		t.Run(serviceFile, func(t *testing.T) {
			if _, err := os.Stat(serviceFile); err != nil {
				t.Fatalf("service file does not exist: %v", err)
			}

			cmd := exec.Command(systemdAnalyze, "verify", serviceFile)

			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			err := cmd.Run()
			if err != nil {
				t.Fatalf("validation failed for %s: %v",
					serviceFile, err)
			}

			output := strings.TrimSpace(out.String())

			// systemd-analyze may return success even with warnings,
			// so fail on any output at all.
			if output != "" {
				t.Fatalf("systemd-analyze reported warnings/errors for %s:\n%s",
					serviceFile, output)
			}
		})
	}
}
