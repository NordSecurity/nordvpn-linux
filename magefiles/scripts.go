package main

import (
	"os"
	"strings"

	"github.com/magefile/mage/sh"
)

type gitInfo struct {
	commitHash string
	versionTag string
}

func getEnv() (map[string]string, error) {
	content, err := os.ReadFile(".env")
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
