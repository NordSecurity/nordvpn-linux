package internal

import (
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set"
)

func StringsToInterfaces(strings []string) []interface{} {
	interfaces := make([]interface{}, len(strings))
	for i, s := range strings {
		interfaces[i] = s
	}
	return interfaces
}

func SetToStrings(set mapset.Set) []string {
	var stringSlice []string
	if set == nil {
		return stringSlice
	}
	for item := range set.Iter() {
		item, _ := item.(string)
		stringSlice = append(stringSlice, item)
	}
	return stringSlice
}

func Title(name string) string {
	splits := strings.Split(name, " ")
	titled := ""
	for _, v := range splits {
		if len(v) == 0 {
			continue
		}
		titled += strings.Title(v) + "_"
	}
	return strings.TrimRight(titled, "_")
}

func SnakeCase(name string) string {
	splits := strings.Split(name, " ")
	lower := ""
	for _, v := range splits {
		if len(v) == 0 {
			continue
		}
		lower += strings.ToLower(v) + "_"
	}
	return strings.TrimRight(lower, "_")
}

func StringsContains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func StringsGetNext(haystack []string, needle string) string {
	var i int
	for i = range haystack {
		if haystack[i] == needle {
			break
		}
	}
	return haystack[(i+1)%len(haystack)]
}

func IntsToStrings(numbers []int) []string {
	if !(len(numbers) > 0) {
		return nil
	}
	strs := make([]string, 0, len(numbers))
	for _, n := range numbers {
		strs = append(strs, strconv.Itoa(n))
	}
	return strs
}
