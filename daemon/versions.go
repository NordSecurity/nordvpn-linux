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

	// match <package ... name="nordvpn" ...> ... </package>
	packageRe := regexp.MustCompile(`(?s)<package[^>]+name="nordvpn"[^>]*>(.*?)</package>`)
	packageMatches := packageRe.FindAllStringSubmatch(string(data), -1)

	versions := make([]string, 0, len(packageMatches))
	for _, pkgMatch := range packageMatches {
		pkgContent := pkgMatch[1]

		// match <version ...> inside the package
		versionTagRe := regexp.MustCompile(`<version\s+([^>]+)\/?>`)
		versionTags := versionTagRe.FindAllStringSubmatch(pkgContent, -1)

		for _, tagMatch := range versionTags {
			attrs := tagMatch[1]

			// extract key="value" pairs
			kvRe := regexp.MustCompile(`(\w+)="([^"]+)"`)
			kvPairs := kvRe.FindAllStringSubmatch(attrs, -1)

			matchMap := make(map[string]string)
			for _, kv := range kvPairs {
				matchMap[kv[1]] = kv[2]
			}

			ver := matchMap["ver"]
			rel := matchMap["rel"]
			if ver != "" && rel != "" {
				versions = append(versions, ver+"-"+rel)
			}
		}
	}

	return validateVersionStrings(versions)
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
