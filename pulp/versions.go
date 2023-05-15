package pulp

import (
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"

	"golang.org/x/mod/semver"
)

func addPrefix(str string) string {
	if strings.HasPrefix(str, "v") {
		return str
	}
	return "v" + str
}

func removePrefix(str string) string {
	return strings.TrimPrefix(str, "v")
}

// transform the input slice and apply a given
// function to each element in the copy.
func transform(slice []string, fn func(string) string) []string {
	ret := []string{}
	for _, val := range slice {
		ret = append(ret, fn(val))
	}
	return ret
}

// unique returns a slice with only unique elements.
func unique(slice []string) []string {
	set := map[string]bool{}
	for _, val := range slice {
		set[val] = true
	}

	ret := []string{}
	for elem := range set {
		ret = append(ret, elem)
	}
	return ret
}

// last elements from the slice up to a given count.
func last(slice []string, count uint) []string {
	length := len(slice)
	if length == 0 || count == 0 {
		return []string{}
	}

	if int(count) > length {
		dst := make([]string, length)
		copy(dst, slice)
		return dst
	}

	lastIndex := length - 1
	ret := []string{}
	for i := 0; i < int(count); i++ {
		ret = append(ret, slice[lastIndex-i])
	}
	return ret
}

func deleteFrom(versions []string, count uint) []string {
	versions = transform(versions, addPrefix)
	semver.Sort(versions) // requires v prefix to work correctly
	uniqueVersions := unique(transform(versions, semver.MajorMinor))
	semver.Sort(uniqueVersions)
	toKeep := last(uniqueVersions, count)

	toDelete := internal.Filter(versions, func(s string) bool {
		for _, prefix := range toKeep {
			if strings.HasPrefix(s, prefix) {
				return false
			}
		}
		return true
	})

	return transform(toDelete, removePrefix)
}
