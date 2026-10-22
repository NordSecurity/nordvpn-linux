package internal

import (
	"regexp"
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
	var lower strings.Builder
	for _, v := range splits {
		if len(v) == 0 {
			continue
		}
		lower.WriteString(strings.ToLower(v) + "_")
	}
	return strings.TrimRight(lower.String(), "_")
}

func RemoveNonAlphanumeric(name string) string {
	return notAlphanumeric.ReplaceAllString(name, "")
}
