package cli

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestFormatDate(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "formats date correctly",
			input:    "2025-05-19 16:06:35",
			expected: "May 19, 2025",
			hasError: false,
		},
		{
			name:     "formats January date",
			input:    "2024-01-01 00:00:00",
			expected: "Jan 1, 2024",
			hasError: false,
		},
		{
			name:     "formats December date",
			input:    "2023-12-31 23:59:59",
			expected: "Dec 31, 2023",
			hasError: false,
		},
		{
			name:     "returns error for invalid format",
			input:    "19-05-2025",
			expected: "",
			hasError: true,
		},
		{
			name:     "returns error for empty string",
			input:    "",
			expected: "",
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := formatDate(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}
