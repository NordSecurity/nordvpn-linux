package log

import (
	"fmt"
	"net"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_sanitize(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		input          []any
		expectedOutput []any
	}{
		{
			name:           "single arg, IPv4 arg is sanitized",
			input:          []any{"1.1.1.1"},
			expectedOutput: []any{"***"},
		},
		{
			name:           "single arg, IPv4 safe arg is not sanitized",
			input:          []any{DoNotSanitize("1.1.1.1")},
			expectedOutput: []any{"1.1.1.1"},
		},
		{
			name:           "multiple args, mixed types",
			input:          []any{"error: 192.168.1.1", 42, true},
			expectedOutput: []any{"error: ***", 42, true},
		},
		{
			name:           "error type with IPv4 is sanitized",
			input:          []any{fmt.Errorf("connection failed at 10.0.0.1")},
			expectedOutput: []any{"connection failed at ***"},
		},
		{
			name:           "string without dots is not processed",
			input:          []any{"no ip here"},
			expectedOutput: []any{"no ip here"},
		},
		{
			name:           "IPv4 in struct string representation is sanitized",
			input:          []any{struct{ addr string }{"172.16.0.1"}},
			expectedOutput: []any{"{***}"},
		},
		{
			name:           "multiple IPv4s in one string",
			input:          []any{"192.168.1.1 and 10.0.0.1"},
			expectedOutput: []any{"*** and ***"},
		},
		{
			name:           "numeric types pass through unchanged",
			input:          []any{uint32(5), float64(3.14), int8(-42)},
			expectedOutput: []any{uint32(5), float64(3.14), int8(-42)},
		},
		{
			name:           "LogSafe with error preserves unsanitized IP",
			input:          []any{DoNotSanitize(fmt.Errorf("failed at 1.2.3.4"))},
			expectedOutput: []any{fmt.Errorf("failed at 1.2.3.4")},
		},
		{
			name:           "net.IP type is sanitized",
			input:          []any{net.ParseIP("192.168.1.100")},
			expectedOutput: []any{"***"},
		},
		{
			name:           "net.IPAddr type is sanitized",
			input:          []any{&net.IPAddr{IP: net.ParseIP("10.0.0.50")}},
			expectedOutput: []any{"***"},
		},
		{
			name:           "net.TCPAddr type is sanitized",
			input:          []any{&net.TCPAddr{IP: net.ParseIP("172.16.0.1"), Port: 8080}},
			expectedOutput: []any{"***:8080"},
		},
		{
			name:           "net.UDPAddr type is sanitized",
			input:          []any{&net.UDPAddr{IP: net.ParseIP("203.0.113.42"), Port: 53}},
			expectedOutput: []any{"***:53"},
		},
		{
			name:           "hostname is sanitized",
			input:          []any{"abc123.nordvpn.com"},
			expectedOutput: []any{"***"},
		},
		{
			name:           "hostname mixed with IPv4 sanitizes both",
			input:          []any{"server at abc123.nordvpn.com has IP 192.168.1.1"},
			expectedOutput: []any{"server at *** has IP ***"},
		},
		{
			name:           "multiple hostnames are sanitized",
			input:          []any{"abc123.nordvpn.com and efg456.nordvpn.com"},
			expectedOutput: []any{"*** and ***"},
		},
		{
			name:           "non-nord hostname is not sanitized",
			input:          []any{"abc123.test.com"},
			expectedOutput: []any{"abc123.test.com"},
		},
		{
			name:           "full form IPv6 is sanitized",
			input:          []any{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
			expectedOutput: []any{"***"},
		},
		{
			name:           "compressed IPv6 with :: is sanitized",
			input:          []any{"2001:db8:85a3::8a2e:370:7334"},
			expectedOutput: []any{"***"},
		},
		{
			name:           "IPv6 loopback address is sanitized",
			input:          []any{"::1"},
			expectedOutput: []any{"***"},
		},
		{
			name:           "IPv6 all zeros is sanitized",
			input:          []any{"::"},
			expectedOutput: []any{"***"},
		},
		{
			name:           "IPv6 link-local address is sanitized",
			input:          []any{"fe80::1"},
			expectedOutput: []any{"***"},
		},
		{
			name:           "IPv4-mapped IPv6 address is sanitized",
			input:          []any{"::ffff:192.0.2.1"},
			expectedOutput: []any{"***"},
		},
		{
			name:           "IPv6 mixed with text is sanitized",
			input:          []any{"error at 2001:db8::1"},
			expectedOutput: []any{"error at ***"},
		},
		{
			name:           "string without IPv6 is not modified",
			input:          []any{"no ip here"},
			expectedOutput: []any{"no ip here"},
		},
		{
			name:           "IPv4 wrapped in quotes is sanitized",
			input:          []any{"\"1.1.1.1\""},
			expectedOutput: []any{"\"***\""},
		},
		{
			name: "telio event log",
			input: []any{
				"\"PublicKey\":\"***\",\"Nickname\":null,\"State\":1,\"LinkState\":1,\"IsExit\":true,\"IsVpn\":true,\"IpAddresses\":[\"10.5.0.1\",\"100.64.0.1\"]",
			},
			expectedOutput: []any{
				"\"PublicKey\":\"***\",\"Nickname\":null,\"State\":1,\"LinkState\":1,\"IsExit\":true,\"IsVpn\":true,\"IpAddresses\":[\"***\",\"***\"]",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := sanitize(test.input...)
			assert.Equal(t, test.expectedOutput, output, "Unexpected sanitization output.")
		})
	}
}
