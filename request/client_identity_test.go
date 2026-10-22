package request

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/sysinfo"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
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
			identity := NewClientIdentity(tt.version, sysinfo.HostInfo{OSPrettyName: tt.distroName})
			assert.Equal(t, tt.want, identity.UserAgent())
		})
	}
}
