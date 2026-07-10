//go:build mage

package main

import (
	"errors"
	"fmt"
	"go/build"
	"maps"
	"os"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	ciJobTokenEnvVar    = "CI_JOB_TOKEN"
	ciJobTokenPassEntry = "nordvpn-linux/ci_job_token"
)

func getEnv() (map[string]string, error) {
	env, err := readVarsFromFile(".env")
	if err != nil {
		return nil, err
	}
	if env["ARCH"] == "" {
		env["ARCH"] = build.Default.GOARCH
		env["ARCHS"] = build.Default.GOARCH
	}
	return env, nil
}

func readVarsFromFile(filename string) (map[string]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	env := map[string]string{}
	for line := range strings.SplitSeq(string(content), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		env[key] = value
	}

	return env, nil
}

func ensureCIJobToken(env map[string]string) error {
	cmd := exec.Command("pass", "show", ciJobTokenPassEntry)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(
			"reading %s from pass entry %q: %w (store it with `pass insert %s`)",
			ciJobTokenEnvVar, ciJobTokenPassEntry, err, ciJobTokenPassEntry,
		)
	}

	token := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	if token == "" {
		return fmt.Errorf("pass entry %q for %s is empty", ciJobTokenPassEntry, ciJobTokenEnvVar)
	}

	env[ciJobTokenEnvVar] = token
	return nil
}

func getVersions() (map[string]string, error) {
	versions, err := readVarsFromFile("lib-versions.env")
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func mergeMaps(m1, m2 map[string]string) map[string]string {
	result := make(map[string]string)

	maps.Copy(result, m1)

	for key, value := range m2 {
		val, exists := result[key]
		if exists {
			log.Info("you are overriding:", val)
		}
		result[key] = value
	}

	return result
}
