package network

import (
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestResolveHost(t *testing.T) {
	category.Set(t, category.Integration)

	// random failures firewall still disabled - wait for it
	// map[ip6tables:[-P INPUT DROP -P FORWARD ACCEPT -P OUTPUT DROP ] iptables:[-P INPUT DROP -P FORWARD ACCEPT -P OUTPUT DROP ]] %!s(<nil>)
	time.Sleep(1 * time.Second)
	result, err := LookupAddressWithCustomDNS("google.com", "1.1.1.1", "udp")
	assert.NoError(t, err)
	assert.NotEmpty(t, result[0])
	assert.NotEmpty(t, result[1])
}

func TestPing(t *testing.T) {
	category.Set(t, category.Root)
	tests := []struct {
		addr string
		ok   bool
	}{
		{"nordvpn.com", true},
		{"1.1.1.1", true},
		{"nonexistingdomain_1234_yes.com", false},
		{"255.0.255.0.1", false},
		// assumes no ipv6
		{"2606:4700:4700::1111", false},
	}

	for _, test := range tests {
		t.Run(test.addr, func(t *testing.T) {
			err := Ping(test.addr, 1)
			if test.ok {
				assert.Nil(t, err, test.addr)
			} else {
				assert.NotNil(t, err, test.addr)
			}
		})
	}
}
