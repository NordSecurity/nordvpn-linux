package sysinfo

import (
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_ReadTagFromOSRelease(t *testing.T) {
	category.Set(t, category.Unit)

	const mockData = `NAME="Ubuntu"
PRETTY_NAME="Ubuntu 22.04 LTS"`

	tests := []struct {
		tag      string
		expected string
		wantErr  bool
	}{
		{"NAME", "Ubuntu", false},
		{"PRETTY_NAME", "Ubuntu 22.04 LTS", false},
		{"VERSION", "", true},
	}

	for _, tt := range tests {
		reader := strings.NewReader(mockData) // Create a new reader per test iteration
		result, err := readTagFromOSRelease(reader, tt.tag)

		assert.Equal(t, tt.wantErr, err != nil,
			"readTagFromOSRelease(%q) = %q, err: %v; want %q, err: %v", tt.tag, result, err, tt.expected, tt.wantErr)

		assert.Equal(t, tt.expected, result,
			"readTagFromOSRelease(%q) = %q, err: %v; want %q, err: %v", tt.tag, result, err, tt.expected, tt.wantErr)
	}

	out, err := readTagFromOSRelease(strings.NewReader(""), "NAME")
	assert.Empty(t, out, "expected empty output, got: %q", out)

	assert.NotNil(t, err, "expected an error for non-existing source, but got nil")
}
