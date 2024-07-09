package norduser

import (
	"fmt"
	"os"
	"strings"
)

func findVariable(name string, environment string) string {
	variables := strings.Split(environment, "\000")
	for _, variable := range variables {
		split := strings.Split(variable, "=")
		if len(split) < 1 {
			continue
		}

		if split[0] != name {
			continue
		}

		if len(split) != 2 {
			return ""
		}

		return split[1]
	}

	return ""
}

func findEnvVariableForPID(pid uint32, name string) (string, error) {
	path := fmt.Sprintf("/proc/%d/environ", pid)
	environment, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	return findVariable(name, string(environment)), nil
}
