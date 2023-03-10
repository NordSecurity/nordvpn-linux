package ipv6

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestGetRulesFrom(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    []byte
		expected map[string]int
		hasError bool
	}{
		{
			name:     "empty",
			input:    []byte{},
			expected: map[string]int{},
			hasError: false,
		},
		{
			name:     "nil",
			input:    nil,
			expected: map[string]int{},
			hasError: false,
		},
		{
			name:     "invalid input",
			input:    []byte{1, 2},
			expected: nil,
			hasError: true,
		},
		{
			name:  "block rule",
			input: []byte("block = 1"),
			expected: map[string]int{
				"block": 1,
			},
			hasError: false,
		},
		{
			name:  "multiple rules",
			input: []byte("block = 1\nallow = 2\ndrop = 3"),
			expected: map[string]int{
				"block": 1,
				"allow": 2,
				"drop":  3,
			},
			hasError: false,
		},
		{
			name:     "multiple ruleswith invalid input",
			input:    []byte("block = 1\nallow = 2\ndrop = 3\nwhatever = &"),
			expected: nil,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parametersFrom(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.EqualValues(t, got, test.expected)
		})
	}
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
