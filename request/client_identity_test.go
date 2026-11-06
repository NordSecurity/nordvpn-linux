package request

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientIdentity_UserAgent(t *testing.T) {
	category.Set(t, category.Unit)

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
			identity := NewClientIdentity(tt.version, "", sysinfo.HostInfo{OSPrettyName: tt.distroName})
			assert.Equal(t, tt.want, identity.UserAgent())
		})
	}
}

func TestClientIdentity_NordAgent(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		buildType features.TypeOfBuild
		hostInfo  sysinfo.HostInfo
		expected  map[string]string
	}{
		{
			name:      "all fields are there",
			buildType: features.BuildTypeInternal,
			hostInfo:  sysinfo.HostInfo{OSPrettyName: "Ubuntu 24.04", Architecture: "x86_64"},
			expected: map[string]string{
				"App":            "NordVPN",
				"Platform":       "Linux",
				"AppVersion":     "5.4.0",
				"OS":             "Ubuntu 24.04",
				"DeviceInfo":     "x86_64",
				"AdditionalInfo": "internal",
			},
		},
		{
			name:      "optional fields unavailable",
			buildType: features.BuildTypeInternal,
			expected: map[string]string{
				"App":            "NordVPN",
				"Platform":       "Linux",
				"AppVersion":     "5.4.0",
				"AdditionalInfo": "internal",
			},
		},
		{
			name:      "open source build",
			buildType: features.BuildTypeOSS,
			hostInfo:  sysinfo.HostInfo{OSPrettyName: "Ubuntu 24.04", Architecture: "x86_64"},
			expected: map[string]string{
				"App":            "NordVPN",
				"Platform":       "Linux",
				"AppVersion":     "5.4.0",
				"OS":             "Ubuntu 24.04",
				"DeviceInfo":     "x86_64",
				"AdditionalInfo": "oss",
			},
		},
		{
			name:      "undefined build type",
			buildType: features.BuildTypeUndefined,
			expected: map[string]string{
				"App":        "NordVPN",
				"Platform":   "Linux",
				"AppVersion": "5.4.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := NewClientIdentity("5.4.0", tt.buildType, tt.hostInfo).NordAgent()

			var got map[string]string
			require.NoError(t, json.Unmarshal([]byte(value), &got))
			assert.Equal(t, tt.expected, got)
		})
	}
}
