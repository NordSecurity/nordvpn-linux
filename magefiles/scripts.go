package main

import (
	"go/build"
	"log"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
)

type gitInfo struct {
	commitHash string
	versionTag string
}

func getEnv() (map[string]string, error) {
	env, err := readVarsFromFile(".env")
	if err != nil {
		return nil, err
	}
	if env["ARCH"] == "" {
		env["ARCH"] = build.Default.GOARCH
	}
	return env, nil
}

func readVarsFromFile(filename string) (map[string]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
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

// TODO: replace with information coming from the Go toolchain
func getGitInfo() (*gitInfo, error) {
	hash, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, err
	}

	version, err := sh.Output("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		return nil, err
	}

	return &gitInfo{
		commitHash: hash,
		versionTag: version,
	}, nil
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
