package request

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserAgentValue(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		distroName string
		want       string
	}{
		{
			name:       "successful user agent generation",
			version:    "1.3.37",
			distroName: "Ubuntu 22.04",
			want:       "NordApp Linux/1.3.37 (Ubuntu 22.04)",
		},
		{
			name:       "empty version",
			version:    "",
			distroName: "Ubuntu 22.04",
			want:       "NordApp Linux/ (Ubuntu 22.04)",
		},
		{
			name:       "with special characters in distro name",
			version:    "1.3.37",
			distroName: "Debian GNU/Linux 11 (bullseye)",
			want:       "NordApp Linux/1.3.37 (Debian GNU/Linux 11 (bullseye))",
		},
		{
			name:       "empty distro name falls back to unknown",
			version:    "1.3.37",
			distroName: "",
			want:       "NordApp Linux/1.3.37 (unknown)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetUserAgentValue(tt.version, tt.distroName))
		})
	}
}
