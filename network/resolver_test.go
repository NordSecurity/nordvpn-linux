package network

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/test/category"
)

func TestAllowlistIP(t *testing.T) {
	category.Set(t, category.Unit)

	resolver := NewResolver(&dns.NameServers{}, 0x1234)
	resolver.Resolve("nordvpn.com")

}
