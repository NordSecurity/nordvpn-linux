package internal

import (
	"fmt"
	"regexp"
)

func CleanUpVersionString(versionString string) (string, error) {
	r := regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+`)
	result := r.FindString(versionString)
	if result == "" {
		return "", fmt.Errorf("invalid version string")
	}

	return result, nil
}
