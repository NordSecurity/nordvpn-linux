package logger

import (
	"testing"

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
	assert.Contains(t, str, "App Version:")
	assert.Contains(t, str, "OS Info:")
	assert.Contains(t, str, "System Info:")
}

func TestGetNetworkInfo(t *testing.T) {
	category.Set(t, category.Route, category.Firewall)
	str := getNetworkInfo()
	assert.Contains(t, str, "Routes for ipv4")
	assert.Contains(t, str, "Routes for ipv6")
	assert.Contains(t, str, "IP rules for ipv4")
	assert.Contains(t, str, "IP rules for ipv6")
	assert.Contains(t, str, "IP tables for ipv4")
	assert.Contains(t, str, "IP tables for ipv6")
}
