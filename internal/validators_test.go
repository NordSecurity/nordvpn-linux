package internal

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

func TestIsAddressValidAsDNSServer(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		// Valid IPv4 addresses.
		{name: "common public dns", address: "1.1.1.1", expected: true},
		{name: "google dns", address: "8.8.8.8", expected: true},
		{name: "loopback", address: "127.0.0.1", expected: true},
		{name: "private class a", address: "10.0.0.1", expected: true},
		{name: "private class c", address: "192.168.1.1", expected: true},
		{name: "all zeros", address: "0.0.0.0", expected: true},
		{name: "broadcast", address: "255.255.255.255", expected: true},
		{name: "max single octet", address: "0.0.0.255", expected: true},

		// IPv4-mapped IPv6 is written in IPv6 notation, so it is rejected.
		{name: "ipv4-mapped ipv6", address: "::ffff:192.168.1.1", expected: false},

		// Valid IPv6 addresses are rejected because To4() is nil.
		{name: "ipv6 loopback", address: "::1", expected: false},
		{name: "ipv6 unspecified", address: "::", expected: false},
		{name: "ipv6 compressed", address: "2001:db8::1", expected: false},
		{name: "ipv6 link-local", address: "fe80::1", expected: false},
		{name: "ipv6 full", address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", expected: false},

		// Malformed / non-IP input.
		{name: "empty string", address: "", expected: false},
		{name: "whitespace only", address: "   ", expected: false},
		{name: "leading space", address: " 1.1.1.1", expected: false},
		{name: "trailing space", address: "1.1.1.1 ", expected: false},
		{name: "octet out of range", address: "256.256.256.256", expected: false},
		{name: "single octet out of range", address: "192.168.1.256", expected: false},
		{name: "negative octet", address: "-1.1.1.1", expected: false},
		{name: "too few octets", address: "1.1.1", expected: false},
		{name: "too many octets", address: "1.1.1.1.1", expected: false},
		{name: "trailing dot", address: "1.1.1.1.", expected: false},
		{name: "leading dot", address: ".1.1.1.1", expected: false},
		{name: "double dot", address: "1..1.1", expected: false},
		{name: "letters in octet", address: "1.1.1.a", expected: false},
		{name: "alphabetic", address: "abc", expected: false},
		{name: "hostname", address: "example.com", expected: false},
		{name: "hex notation", address: "0x1.0x1.0x1.0x1", expected: false},
		{name: "cidr notation", address: "192.168.1.0/24", expected: false},
		{name: "cidr single host", address: "1.1.1.1/32", expected: false},
		{name: "address with port", address: "1.1.1.1:53", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAddressValidAsDNSServer(tt.address))
		})
	}
}
