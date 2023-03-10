/*
Package kernel provides functions to get/set kernel parameters
*/
package kernel

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// SetParameter set linux kernel parameter value
func SetParameter(key string, val int) error {
	// #nosec G204 -- user input is not passed in
	if out, err := exec.Command("sysctl", "-w", fmt.Sprintf("%s=%d", key, val)).CombinedOutput(); err != nil {
		return fmt.Errorf("setting %s to %d: %w: %s", key, val, err, string(out))
	}
	return nil
}

// Parameter returns current sysctl parameters map (one or multiple values)
func Parameter(prm string) (map[string]int, error) {
	// expect e.g.: "net.ipv4.ip_forward = 1"
	out, err := exec.Command("sysctl", prm).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("listing sysctl rules error: %w: %s", err, string(out))
	}
	return parametersFrom(out)
}

// parametersFrom parse output and put values into map
func parametersFrom(output []byte) (map[string]int, error) {
	rules := map[string]int{}
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		// https://man7.org/linux/man-pages/man5/sysctl.conf.5.html
		parts := strings.Split(line, " = ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid output line: %s", line)
		}
		key, strVal := parts[0], parts[1]
		val, err := strconv.Atoi(strings.Trim(strVal, " "))
		if err != nil {
			return nil, fmt.Errorf("parsing value of %s: %w. expected integer, got: %s", key, err, strVal)
		}
		rules[key] = val
	}
	return rules, nil
}
