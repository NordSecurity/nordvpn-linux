package internal

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var notAlphanumeric = regexp.MustCompile(`[^0-9a-zA-Z \-_]+`)

func StringsToInterfaces(strings []string) []any {
	interfaces := make([]any, len(strings))
	for i, s := range strings {
		interfaces[i] = s
	}
	return interfaces
}

func Title(name string) string {
	name = RemoveNonAlphanumeric(name)
	name = strings.Join(strings.Fields(name), " ")
	titled := cases.Title(language.English, cases.NoLower).String(name)
	return strings.ReplaceAll(titled, " ", "_")
}

func SnakeCase(name string) string {
	name = RemoveNonAlphanumeric(name)
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

func RemoveNonAlphanumeric(name string) string {
	return notAlphanumeric.ReplaceAllString(name, "")
}

func StringsContains(haystack []string, needle string) bool {
	return slices.Contains(haystack, needle)
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
	if len(numbers) <= 0 {
		return nil
	}
	strs := make([]string, 0, len(numbers))
	for _, n := range numbers {
		strs = append(strs, strconv.Itoa(n))
	}
	return strs
}

func CopyStringSlice(src []string) []string {
	return append(make([]string, 0, len(src)), src...)
}
