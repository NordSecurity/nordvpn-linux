package config

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestHasSensitiveServerGroup(t *testing.T) {
	category.Set(t, category.Unit)

	cases := []struct {
		name   string
		groups []ServerGroup
		want   bool
	}{
		{
			name: "contains_dedicated_ip",
			groups: []ServerGroup{
				ServerGroup_DEDICATED_IP,
				ServerGroup_STANDARD_VPN_SERVERS,
			},
			want: true,
		},
		{
			name:   "only_dedicated_ip",
			groups: []ServerGroup{ServerGroup_DEDICATED_IP},
			want:   true,
		},
		{
			name: "no_sensitive_groups",
			groups: []ServerGroup{
				ServerGroup_STANDARD_VPN_SERVERS,
			},
			want: false,
		},
		{
			name:   "empty_slice",
			groups: []ServerGroup{},
			want:   false,
		},
		{
			name:   "nil_slice",
			groups: nil,
			want:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, HasSensitiveServerGroup(c.groups))
		})
	}
}
