package request

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockOsHostName struct {
	prettyName string
	err        error
}

func (m MockOsHostName) GetHostOSPrettyName() (string, error) { return m.prettyName, m.err }

func TestGetUserAgentValue(t *testing.T) {
	tests := []struct {
		name    string
		version string
		mock    MockOsHostName
		want    string
		wantErr string
	}{
		{
			name:    "successful user agent generation",
			version: "1.3.37",
			mock:    MockOsHostName{prettyName: "Ubuntu 22.04", err: nil},
			want:    "NordApp Linux/1.3.37 (Ubuntu 22.04)",
		},
		{
			name:    "empty version",
			version: "",
			mock:    MockOsHostName{prettyName: "Ubuntu 22.04", err: nil},
			want:    "NordApp Linux/ (Ubuntu 22.04)",
		},
		{
			name:    "distro error",
			version: "1.3.37",
			mock:    MockOsHostName{prettyName: "", err: errors.New("os error")},
			wantErr: "determining device os: os error",
		},
		{
			name:    "with special characters in distro name",
			version: "1.3.37",
			mock:    MockOsHostName{prettyName: "Debian GNU/Linux 11 (bullseye)", err: nil},
			want:    "NordApp Linux/1.3.37 (Debian GNU/Linux 11 (bullseye))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserAgentValue(tt.version, tt.mock.GetHostOSPrettyName)

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
