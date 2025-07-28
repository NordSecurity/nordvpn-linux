//go:build mage

package main

import (
	"errors"
	"go/build"
	"log"
	"os"
	"os/exec"
	"strings"
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
	for _, line := range strings.Split(string(content), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		env[key] = value
	}

	return env, nil
}

func getVersions() (map[string]string, error) {
	versions, err := readVarsFromFile("lib-versions.env")
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func getGitVersionTag() string {
	cmd := exec.Command("git", "describe", "--tags", "--always")
	out, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(out))
}

func mergeMaps(m1, m2 map[string]string) map[string]string {
	result := make(map[string]string)

	for key, value := range m1 {
		result[key] = value
	}

	for key, value := range m2 {
		val, exists := result[key]
		if exists {
			log.Println("you are overriding:", val)
		}
		result[key] = value
	}

	return result
}
