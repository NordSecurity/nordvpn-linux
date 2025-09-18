package logger

import (
	"net/http"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestMaskIPRouteOutput(t *testing.T) {
	input4 := `default dev nordtun table 205 scope link\n
default via 180.144.168.176 dev wlp0s20f3 proto dhcp metric 20600\n
172.31.100.100/24 dev nordtun proto kernel scope link src 192.168.200.203\n
114.237.30.247/16 dev wlp0s20f3 scope link metric 1000\n
local 10.128.10.7 dev wlp0s20f3 table local proto kernel scope link src 26.14.182.220`

	maskedInput4 := maskIPRouteOutput(input4)

	expectedOutput4 := `default dev nordtun table 205 scope link\n
default via *** dev wlp0s20f3 proto dhcp metric 20600\n
172.31.100.100/24 dev nordtun proto kernel scope link src 192.168.200.203\n
***/16 dev wlp0s20f3 scope link metric 1000\n
local 10.128.10.7 dev wlp0s20f3 table local proto kernel scope link src ***`

	assert.Equal(t, expectedOutput4, maskedInput4)

	input6 := `default dev nordtun table 205 scope link\n
	default via fd31:482b:86d9:7142::1 dev wlp0s20f3 proto dhcp metric 20600\n
	8d02:d70f:76b4:162e:d12f:b0e6:204a:59d1 dev nordtun proto kernel scope link src 24ef:7163:ffd8:4ee7:16f8:008b:e52b:0a68\n
	1e66:9f56:66b5:b846:8d27:d0b5:0821:c819 dev wlp0s20f3 scope link metric 1000\n
	local fc81:9a6e:dcf2:20a7::2 dev wlp0s20f3 table local proto kernel scope link src fdf3:cbf9:573c:8e15::3`

	maskedInput6 := maskIPRouteOutput(input6)

	expectedOutput6 := `default dev nordtun table 205 scope link\n
	default via fd31:482b:86d9:7142::1 dev wlp0s20f3 proto dhcp metric 20600\n
	*** dev nordtun proto kernel scope link src ***\n
	*** dev wlp0s20f3 scope link metric 1000\n
	local fc81:9a6e:dcf2:20a7::2 dev wlp0s20f3 table local proto kernel scope link src fdf3:cbf9:573c:8e15::3`

	assert.Equal(t, expectedOutput6, maskedInput6)
}

func TestGetSystemInfo(t *testing.T) {
	category.Set(t, category.Integration)
	str := getSystemInfo()
	assert.Contains(t, str, "OS Info:")
	assert.Contains(t, str, "System Info:")
}

func TestGetNetworkInfo(t *testing.T) {
	category.Set(t, category.Route, category.Firewall)
	str := getNetworkInfo()
	assert.Contains(t, str, "Routes for ipv4")
	assert.Contains(t, str, "IP rules for ipv4")
	assert.Contains(t, str, "IP tables for ipv4")
}

const (
	ResponseHeadersBinary = `Accept-Ranges:[bytes] Age:[1603] Alt-Svc:[h3=":443"; ma=86400] Cache-Control:[public, max-age=86400] Cf-Cache-Status:[HIT] Cf-Ray:[8e5e2a3f8b1d4d84-FRA] Connection:[keep-alive] Content-Length:[1285] Content-Type:[application/octet-stream] Date:[Thu, 21 Nov 2024 05:08:59 GMT] Etag:["673b8ffa-505"] Expires:[Fri, 22 Nov 2024 05:08:59 GMT] Last-Modified:[Mon, 18 Nov 2024 19:05:30 GMT] Server:[cloudflare] Set-Cookie:[__cf_bm=qoPfqfJQjE5Dz0TtNSosvHfFHUZgFFBVVBQlSUJXvFs-1732165739-1.0.1.1-WzZaWHlOfYorXL61QyiQhQ.8aYflwBkrKlbcxPBRma55M3iKH9XUmyB6L2hCfRDvyIFleG05LwQ0RM5h7Brnzmnw8DU9sPjm3lRAnIxB0Y4; path=/; expires=Thu, 21-Nov-24 05:38:59 GMT; domain=.nordvpn.com; HttpOnly; Secure; SameSite=None] Strict-Transport-Security:[max-age=31536000; includeSubDomains; preload] Vary:[Accept-Encoding] X-Frame-Options:[SAMEORIGIN] X-Robots-Tag:[noindex, nofollow, nosnippet, noarchive]`
	ResponseHeadersText   = `Alt-Svc:[h3=":443"; ma=86400] Cf-Cache-Status:[DYNAMIC] Cf-Ray:[8e5e2a3f5b71918f-FRA] Content-Length:[184] Content-Security-Policy:[frame-ancestors "none"] Content-Type:[application/json;charset=utf-8] Date:[Thu, 21 Nov 2024 05:08:59 GMT] Priority:[u=3,i=?0] Server:[cloudflare] Server-Timing:[cfExtPri] Set-Cookie:[__cf_bm=6NjoQboqV5FhNttmRuWZCfUlwjuioCXFASuNlsRKl68-1732165739-1.0.1.1-50WHx4PPBB2E1vIXZbqmTg.5kWEbh_XykR0jX80RT5nQCQvUgH.smH1GYe6DbJT2GGBtZAohrjftnPUoR5C9NwKKMEW7eeP2JCav.aWi.dY; path=/; expires=Thu, 21-Nov-24 05:38:59 GMT; domain=.nordvpn.com; HttpOnly; Secure; SameSite=None] Strict-Transport-Security:[max-age=31536000; includeSubDomains; preload] Vary:[Accept-Encoding] X-Accept-Before:[1732208939] X-Authorization:[key-id="rsa-key-1",algorithm="rsa-sha256"] X-Digest:[223cc5928a75a1ec1c806f87c04590b0bf3e5e58b12d684c56400cd61cff344b] X-Frame-Options:[deny] X-Signature:[SSvjRicb7tR4adIM22T9RI+n8DVoM28ZRF/QlmdDk6qppatA5wxrH2l1zdS2wQ5u8VLlqjjX+/3ZPr4Wb2GnwJFZztR6YVAXxhsrJt39XsUvnZnfkKpS4mTQBbme+rTAqYGegJYJ/qqKFxcdkgqQUwneAOcbeGqn9S89AkRf7rnaKqW04Iot+DO/ltJhxzSgnsyC0AAUknHmhFpvGzxRI2IclQMLdNZ/2bmyOQGpySAKjKBsj/qxygIx4+yTp+B9vrvhWGVXyKXCk6FQ2TPSgS5jiyhwwTHDtyJhfOjs4RldK603MIScII9lgng4b9Vf9WfFYfVv4WDqPFhHcXR+AIhTquqBE34Fa/ad3QZT8syO4oZS3dCg2GjKmCVX01NavVvLZQQyaIXBwJyY3piP/nZMOPUhYYRsCyOSLGsj18EFnZRuUzTlyJC/HsofamHPZoPYX27ZreDd0Pd0el30KhdWWm8zz79yIUvxm5cmdk1VnI5SCQ5hcizKh9ufpG4vq9XJYxEwkdPnhffFhr7QBeZRTIeDflT13HPtw+J8eKEVFpLAcglj0lSTxYPrMJuh83p/phG/KkaSgV3Fyy8ZdQYRnC8lMBV3hp5ioileuh6eZxLXOYjOtZZqQQYTMCVsAkDUyJ4liUxON0vdpNS2cqoEN5ysVdw2O11wVmqV0r8=]`
)

func setupResponseHeaders(headersStr string) http.Header {
	result := http.Header{} // map[string]string
	for _, pair := range strings.Split(headersStr, " ") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(parts[1], "[]")
			result.Add(key, value)
		}
	}
	return result
}

func Test_dataRequestAPIToString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                  string
		responseHeadersString string
		responseBody          []byte
		responseBodyBinary    bool
	}{
		{
			name:                  "Binary response",
			responseHeadersString: ResponseHeadersBinary,
			responseBody:          []byte{0x01, 0x02, 0x03, 0x04},
			responseBodyBinary:    true,
		},
		{
			name:                  "Non binary response",
			responseHeadersString: ResponseHeadersText,
			responseBody:          []byte("Text body"),
			responseBodyBinary:    false,
		},
	}

	for _, test := range tests {
		data := events.DataRequestAPI{
			Request: &http.Request{
				Header: http.Header{},
			},
			Response: &http.Response{
				Header: setupResponseHeaders(test.responseHeadersString),
			},
		}
		rc := dataRequestAPIToString(data, nil, test.responseBody, false)

		if !test.responseBodyBinary {
			assert.True(t, strings.Contains(rc, string(test.responseBody)))
		}
	}
}
