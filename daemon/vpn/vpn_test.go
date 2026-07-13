package vpn

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestServerData_EndpointEqual(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		endpoint string
		other    string
		want     bool
	}{
		{
			name:     "identical IPv4 endpoints",
			endpoint: "1.2.3.4:51820",
			other:    "1.2.3.4:51820",
			want:     true,
		},
		{
			name:     "different IPv4 address",
			endpoint: "1.2.3.4:51820",
			other:    "1.2.3.5:51820",
			want:     false,
		},
		{
			name:     "different port",
			endpoint: "1.2.3.4:51820",
			other:    "1.2.3.4:51821",
			want:     false,
		},
		{
			name:     "identical IPv6 endpoints",
			endpoint: "[2001:db8::1]:51820",
			other:    "[2001:db8::1]:51820",
			want:     true,
		},
		{
			name:     "empty both sides returns false",
			endpoint: "",
			other:    "",
			want:     false,
		},
		{
			name:     "malformed receiver returns false",
			endpoint: "not-an-endpoint",
			other:    "not-an-endpoint",
			want:     false,
		},
		{
			name:     "malformed receiver with valid other returns false",
			endpoint: "not-an-endpoint",
			other:    "1.2.3.4:51820",
			want:     false,
		},
		{
			name:     "malformed other returns false",
			endpoint: "1.2.3.4:51820",
			other:    "not-an-endpoint",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ServerData{Endpoint: tt.endpoint}
			assert.Equal(t, tt.want, s.EndpointEqual(tt.other), tt.name)
		})
	}
}
