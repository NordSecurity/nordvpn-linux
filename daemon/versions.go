package daemon

import (
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
)

func ParseDebianVersions(data []byte) []string {
	// Get information about nordvpn package only
	nordPattern := regexp.MustCompile(`Package: nordvpn\nVersion: .*`)
	matches := nordPattern.FindAllString(string(data), -1)

	// Get version numbers
	versionPattern := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}(-\d{1,3})?`)
	matches = versionPattern.FindAllString(strings.Join(matches, "\n"), -1)

	for i := range matches {
		if !strings.Contains(matches[i], "-") {
			matches[i] += "-0"
		}
	}

	matches = validateVersionStrings(matches)
	return matches
}

func ParseRpmVersions(data []byte) []string {
	// get release and version info
	versionPattern := regexp.MustCompile(`rel="\d{1,3}" ver=".*"`)
	matches := versionPattern.FindAllString(string(data), -1)

	for i := range matches {
		// split to ["rel=", releaseInt, " ver=", versionString, ""]
		quoteSplit := strings.Split(matches[i], "\"")

		matches[i] = quoteSplit[3] + "-" + quoteSplit[1]
	}

	matches = validateVersionStrings(matches)
	return matches
}

func validateVersionStrings(versions []string) []string {
	validated := make([]string, 0, len(versions))
	versionPattern := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}-\d{1,3}$`)

	for _, ver := range versions {
		if versionPattern.MatchString(ver) {
			validated = append(validated, ver)
		}
	}
	return validated
}

func StringsToVersions(v []string) []semver.Version {
	var versions []semver.Version
	for _, z := range v {
		versions = append(versions, *semver.New(z))
	}
	return versions
}

func GetLatestVersion(versions []semver.Version) semver.Version {
	if len(versions) == 0 {
		return semver.Version{}
	}
	newest := versions[0]
	if len(versions) == 1 {
		return newest
	}

	for _, ver := range versions[1:] {
		if newest.LessThan(ver) {
			newest = ver
		}
	}
	return newest
}
