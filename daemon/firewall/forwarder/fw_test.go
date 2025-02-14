package forwarder

import (
	"fmt"
	"net/netip"
	"os/exec"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes/netlink"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func commandFunc(command string, arg ...string) ([]byte, error) {
	return exec.Command(command, arg...).CombinedOutput()
}

func TestFiltering(t *testing.T) {
	category.Set(t, category.Route)

	rc, err := checkFilteringRule(meshSrcSubnet, commandFunc)
	assert.NoError(t, err)
	assert.False(t, rc)

	err = enableFiltering(commandFunc)
	assert.NoError(t, err)

	rc, err = checkFilteringRule(meshSrcSubnet, commandFunc)
	assert.NoError(t, err)
	assert.True(t, rc)

	err = clearFiltering(commandFunc)
	assert.NoError(t, err)

	rc, err = checkFilteringRule(meshSrcSubnet, commandFunc)
	assert.NoError(t, err)
	assert.False(t, rc)

	_, intf, err := netlink.Retriever{}.Retrieve(netip.Prefix{}, 0)

	assert.NoError(t, err)
	assert.NotEmpty(t, intf.Name)

	ip := netip.MustParsePrefix("100.77.1.1/32")

	interfaceNames := []string{"eth0"}
	err = resetForwardTraffic(
		[]TrafficPeer{{ip, true, false}},
		interfaceNames,
		commandFunc,
		false,
		false,
		[]string{})
	assert.NoError(t, err)

	rc, err = checkFilteringRule(ip.String(), commandFunc)
	assert.NoError(t, err)
	assert.True(t, rc)

	err = clearFiltering(commandFunc)
	assert.NoError(t, err)
}

func TestResetPeersTraffic(t *testing.T) {
	category.Set(t, category.Route)

	defer clearFiltering(commandFunc)
	err := enableFiltering(commandFunc)
	assert.NoError(t, err)

	ip1 := netip.MustParsePrefix("100.77.1.1/32")
	ip2 := netip.MustParsePrefix("200.88.2.2/32")
	tests := []struct {
		peers                        []TrafficPeer
		ruleExists                   []bool
		ruleAbovePrivateSubnetsBlock []bool
		ruleIsLocalOnly              []bool
	}{
		{
			peers:                        []TrafficPeer{{ip1, false, false}, {ip2, false, false}},
			ruleExists:                   []bool{false, false},
			ruleAbovePrivateSubnetsBlock: []bool{false, false},
			ruleIsLocalOnly:              []bool{false, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, true, false}, {ip2, false, false}},
			ruleExists:                   []bool{true, false},
			ruleAbovePrivateSubnetsBlock: []bool{false, false},
			ruleIsLocalOnly:              []bool{false, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, false, true}, {ip2, false, false}},
			ruleExists:                   []bool{true, false},
			ruleAbovePrivateSubnetsBlock: []bool{true, false},
			ruleIsLocalOnly:              []bool{true, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, false, false}, {ip2, true, false}},
			ruleExists:                   []bool{false, true},
			ruleAbovePrivateSubnetsBlock: []bool{false, false},
			ruleIsLocalOnly:              []bool{false, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, false, false}, {ip2, false, true}},
			ruleExists:                   []bool{false, true},
			ruleAbovePrivateSubnetsBlock: []bool{false, true},
			ruleIsLocalOnly:              []bool{false, true},
		},
		{
			peers:                        []TrafficPeer{{ip1, true, true}, {ip2, false, false}},
			ruleExists:                   []bool{true, false},
			ruleAbovePrivateSubnetsBlock: []bool{true, false},
			ruleIsLocalOnly:              []bool{false, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, true, false}, {ip2, true, false}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{false, false},
			ruleIsLocalOnly:              []bool{false, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, true, false}, {ip2, false, true}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{false, true},
			ruleIsLocalOnly:              []bool{false, true},
		},
		{
			peers:                        []TrafficPeer{{ip1, false, true}, {ip2, true, false}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{true, false},
			ruleIsLocalOnly:              []bool{true, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, false, true}, {ip2, false, true}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{true, true},
			ruleIsLocalOnly:              []bool{true, true},
		},
		{
			peers:                        []TrafficPeer{{ip1, false, false}, {ip2, true, true}},
			ruleExists:                   []bool{false, true},
			ruleAbovePrivateSubnetsBlock: []bool{false, true},
			ruleIsLocalOnly:              []bool{false, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, true, true}, {ip2, true, false}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{true, false},
			ruleIsLocalOnly:              []bool{false, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, true, true}, {ip2, false, true}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{true, true},
			ruleIsLocalOnly:              []bool{false, true},
		},
		{
			peers:                        []TrafficPeer{{ip1, true, false}, {ip2, true, true}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{false, true},
			ruleIsLocalOnly:              []bool{false, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, false, true}, {ip2, true, true}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{true, true},
			ruleIsLocalOnly:              []bool{true, false},
		},
		{
			peers:                        []TrafficPeer{{ip1, true, true}, {ip2, true, true}},
			ruleExists:                   []bool{true, true},
			ruleAbovePrivateSubnetsBlock: []bool{true, true},
			ruleIsLocalOnly:              []bool{false, false},
		},
	}

	ruleExists := func(ip netip.Prefix) bool {
		exists, err := checkFilteringRule(ip.String(), commandFunc)
		assert.NoError(t, err)
		return exists
	}

	ruleAbovePrivateSubnetsBlock := func(ip netip.Prefix) bool {
		line1, err := checkFilteringRulesLine([]string{ip.String()}, commandFunc)
		assert.NoError(t, err)
		line2, err := checkFilteringRulesLine([]string{meshSrcSubnet}, commandFunc)
		assert.NoError(t, err)
		return line1 != -1 && line1 < line2
	}

	ruleIsLocalOnly := func(ip netip.Prefix) bool {
		for _, localIP := range []string{"10.0.0.0", "172.16.0.0", "192.168.0.0", "169.254.0.0"} {
			line, err := checkFilteringRulesLine([]string{ip.String(), localIP}, commandFunc)
			assert.NoError(t, err)
			if line == -1 {
				return false
			}
		}
		return true
	}

	interfaceNames := []string{"eth0"}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test.peers), func(t *testing.T) {
			err = resetForwardTraffic(test.peers, interfaceNames, commandFunc, false, false, []string{})
			assert.NoError(t, err)

			for i, peer := range test.peers {
				assert.Equal(t, test.ruleExists[i], ruleExists(peer.IP))
				assert.Equal(t, test.ruleAbovePrivateSubnetsBlock[i], ruleAbovePrivateSubnetsBlock(peer.IP))
				assert.Equal(t, test.ruleIsLocalOnly[i], ruleIsLocalOnly(peer.IP))
			}
		})
	}
}
