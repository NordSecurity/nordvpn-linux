//go:build mage

package main

import (
	"errors"
	"fmt"
	"go/build"
	"io/fs"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/coreos/go-semver/semver"
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

// TODO: replace with information coming from the Go toolchain
func getGitInfo() (*gitInfo, error) {
	hash, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, err
	}

	version, err := getLatestVersion()
	if err != nil {
		return nil, err
	}

	return &gitInfo{
		commitHash: hash,
		versionTag: version,
	}, nil
}

func getLatestVersion() (string, error) {
	version, err := getLatestVersionGit()
	if err != nil {
		return "", err
	}
	if version != "" {
		return version, nil
	}
	fmt.Println("Unable to determine version from git tags. Using changelog.")

	version, err = getLatestVersionChangelog()
	if err != nil {
		return "", err
	}
	if version != "" {
		return version, nil
	}
	fmt.Println("Unable to determine version the changelog. Using 0.0.0")

	return "0.0.0", nil
}

func getLatestVersionGit() (string, error) {
	version, err := sh.Output("git", "describe", "--tags", "--abbrev=0")
	if err != nil && !strings.Contains(err.Error(), "exit code 128") {
		return "", err
	}
	return version, nil
}

func getLatestVersionChangelog() (string, error) {
	changelogs, err := os.ReadDir("contrib/changelog/prod")
	if err != nil {
		return "", err
	}
	changelog := slices.MaxFunc(changelogs, func(a fs.DirEntry, b fs.DirEntry) int {
		return changelogMdToSemver(a.Name()).Compare(changelogMdToSemver(b.Name()))
	})
	if changelog == nil {
		return "", nil
	}
	return changelogMdToSemver(changelog.Name()).String(), nil
}

func changelogMdToSemver(filename string) semver.Version {
	if !strings.HasSuffix(filename, ".md") {
		return semver.Version{}
	}
	ver, err := semver.NewVersion(strings.Split(strings.TrimSuffix(filename, ".md"), "_")[0])
	if err != nil && ver == nil {
		return semver.Version{}
	}
	return *ver
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
