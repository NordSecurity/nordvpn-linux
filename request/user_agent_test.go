package request

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockDistro struct {
	prettyName string
	err        error
}

func (m mockDistro) ReleasePrettyName() (string, error) { return m.prettyName, m.err }
func (m mockDistro) ReleaseName() (string, error)       { return "", nil }
func (m mockDistro) KernelName() string                 { return "" }
func (m mockDistro) KernelFull() string                 { return "" }

func TestGetUserAgentValue(t *testing.T) {
	tests := []struct {
		name    string
		version string
		distro  mockDistro
		want    string
		wantErr string
	}{
		{
			name:    "successful user agent generation",
			version: "1.3.37",
			distro:  mockDistro{prettyName: "Ubuntu 22.04", err: nil},
			want:    "NordApp Linux/1.3.37 (Ubuntu 22.04)",
		},
		{
			name:    "empty version",
			version: "",
			distro:  mockDistro{prettyName: "Ubuntu 22.04", err: nil},
			want:    "NordApp Linux/ (Ubuntu 22.04)",
		},
		{
			name:    "distro error",
			version: "1.3.37",
			distro:  mockDistro{prettyName: "", err: errors.New("os error")},
			wantErr: "determining device os: os error",
		},
		{
			name:    "with special characters in distro name",
			version: "1.3.37",
			distro:  mockDistro{prettyName: "Debian GNU/Linux 11 (bullseye)", err: nil},
			want:    "NordApp Linux/1.3.37 (Debian GNU/Linux 11 (bullseye))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserAgentValue(tt.version, tt.distro)

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
