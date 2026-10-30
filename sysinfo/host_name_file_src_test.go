package sysinfo

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_OSReleaseTag(t *testing.T) {
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
		result, err := osRelease(mockData).tag(tt.tag)

		assert.Equal(t, tt.wantErr, err != nil,
			"tag(%q) = %q, err: %v; want %q, err: %v", tt.tag, result, err, tt.expected, tt.wantErr)

		assert.Equal(t, tt.expected, result,
			"tag(%q) = %q, err: %v; want %q, err: %v", tt.tag, result, err, tt.expected, tt.wantErr)
	}

	out, err := osRelease("").tag("NAME")
	assert.Empty(t, out, "expected empty output, got: %q", out)

	assert.NotNil(t, err, "expected an error for non-existing source, but got nil")
}
