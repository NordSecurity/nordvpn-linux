package iprule

import (
	"net"
	"slices"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/vishvananda/netlink"

	"github.com/stretchr/testify/assert"
)

func TestFwmarkRule(t *testing.T) {
	category.Set(t, category.Route)

	prioID, err := calculateRulePriority(false)
	assert.NoError(t, err)
	assert.Greater(t, prioID, uint(0))

	tblID, err := calculateCustomTableID(false)
	assert.NoError(t, err)
	assert.Greater(t, tblID, uint(0))

	err = addFwmarkRule(0, prioID, tblID, false)
	assert.Error(t, err)

	var fwmarkval uint32 = 0xe1f1
	assert.Greater(t, fwmarkval, uint32(0))

	err = addFwmarkRule(fwmarkval, prioID, tblID, false)
	assert.NoError(t, err)

	fndTblID, err := findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, tblID, fndTblID)

	err = removeFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)

	fndTblID, err = findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), fndTblID)
}

func TestMultiFwmarkRule(t *testing.T) {
	category.Set(t, category.Route)

	prioID, err := calculateRulePriority(false)
	assert.NoError(t, err)
	assert.Greater(t, prioID, uint(0))

	tblID, err := calculateCustomTableID(false)
	assert.NoError(t, err)
	assert.Greater(t, tblID, uint(0))

	var fwmarkval uint32 = 0xe1f1
	assert.Greater(t, fwmarkval, uint32(0))

	// add 1st rule
	err = addFwmarkRule(fwmarkval, prioID, tblID, false)
	assert.NoError(t, err)

	// check
	fndTblID, err := findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, tblID, fndTblID)

	prioID2, err := calculateRulePriority(false)
	assert.NoError(t, err)
	assert.Greater(t, prioID2, uint(0))

	tblID2, err := calculateCustomTableID(false)
	assert.NoError(t, err)
	assert.Greater(t, tblID2, uint(0))

	// add 2nd rule
	err = addFwmarkRule(fwmarkval, prioID2, tblID2, false)
	assert.NoError(t, err)

	// check
	fndTblID, err = findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, tblID, fndTblID)

	// will remove one
	err = removeFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)

	fndTblID, err = findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, tblID2, fndTblID)

	err = removeFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)

	fndTblID, err = findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), fndTblID)
}

func TestSuppressRule(t *testing.T) {
	category.Set(t, category.Route)

	prioID, err := calculateRulePriority(false)
	assert.NoError(t, err)
	assert.Greater(t, prioID, uint(0))

	tblID, err := calculateCustomTableID(false)
	assert.NoError(t, err)
	assert.Greater(t, tblID, uint(0))

	var fwmarkval uint32 = 0xe1f1
	assert.Greater(t, fwmarkval, uint32(0))

	fndTblID, err := findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), fndTblID)

	err = addFwmarkRule(fwmarkval, prioID, tblID, false)
	assert.NoError(t, err)

	prioID2, err := calculateRulePriority(false)
	assert.NoError(t, err)
	assert.Greater(t, prioID2, uint(0))

	fnd, err := checkSuppressRule(false)
	assert.NoError(t, err)
	assert.False(t, fnd)

	err = addSuppressRule(prioID2, false, true)
	assert.NoError(t, err)

	fnd, err = checkSuppressRule(false)
	assert.NoError(t, err)
	assert.True(t, fnd)

	err = removeSuppressRule(false)
	assert.NoError(t, err)

	err = addSuppressRule(prioID2, false, false)
	assert.NoError(t, err)

	fnd, err = checkSuppressRule(false)
	assert.NoError(t, err)
	assert.True(t, fnd)

	err = removeSuppressRule(false)
	assert.NoError(t, err)

	fnd, err = checkSuppressRule(false)
	assert.NoError(t, err)
	assert.False(t, fnd)

	err = removeFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)

	fndTblID, err = findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), fndTblID)
}

func TestCustomTable(t *testing.T) {
	category.Set(t, category.Route)

	prioID, err := calculateRulePriority(false)
	assert.NoError(t, err)
	assert.Greater(t, prioID, uint(0))

	tblID, err := calculateCustomTableID(false)
	assert.NoError(t, err)
	assert.Greater(t, tblID, uint(0))

	var fwmarkval uint32 = 0xe1f1
	assert.Greater(t, fwmarkval, uint32(0))

	err = addFwmarkRule(fwmarkval, prioID, tblID, false)
	assert.NoError(t, err)

	fndTblID, err := findFwmarkRule(fwmarkval, false)
	assert.NoError(t, err)
	assert.Equal(t, tblID, fndTblID)

	// prev prio id is in use, so, next one should be less (less is higher prio)
	prioID2, err := calculateRulePriority(false)
	assert.NoError(t, err)
	assert.Greater(t, prioID, prioID2)
}

func TestAddAllowlistRules(t *testing.T) {
	category.Set(t, category.Route)

	rulesBeforeAdd, err := netlink.RuleList(netlink.FAMILY_V4)
	assert.NoError(t, err, "Failed to save previous state for ip rules: %s", err)

	router := Router{
		subnetToRulePriority: make(map[string]uint),
	}

	allowlist := []string{
		// two /0 subnets should be simplified to a single default route
		"35.74.174.235/0",
		"4.246.215.86/0",
		"103.238.215.35/24",
		"237.164.3.235/32",
	}

	err = router.addAllowlistRules(allowlist, false)
	assert.NoError(t, err, "Unexpected error when adding allowlist rules: %s", err)

	rulesAfterAdd, err := netlink.RuleList(netlink.FAMILY_V4)

	assert.NoError(t, err, "Failed to save new state for ip rules: %s", err)
	// ignore rules that existed before the test for validation purposes
	rulesAfterAdd = slices.DeleteFunc(rulesAfterAdd, func(rule netlink.Rule) bool {
		idx := slices.Index(rulesBeforeAdd, rule)
		return idx != -1
	})

	_, subnet2, _ := net.ParseCIDR("103.238.215.35/24")
	_, subnet3, _ := net.ParseCIDR("237.164.3.235/32")

	expectedSubnets := []*net.IPNet{
		nil, // in case of netlink, nil is equivalent to default subnet(35.74.174.235/0 and 4.246.215.86/0)
		subnet2,
		subnet3,
	}

	for _, subnet := range expectedSubnets {
		ruleIdx := slices.IndexFunc(rulesAfterAdd, func(rule netlink.Rule) bool {
			return subnet.String() == rule.Dst.String()
		})
		assert.NotEqual(t, -1, ruleIdx, "Desired subnet not added to ip rules.")
	}

	router.removeAllowSubnetRules(false)
	rulesAfterRemove, err := netlink.RuleList(netlink.FAMILY_V4)
	assert.NoError(t, err, "Failed to save prevous state for ip rules: %s", err)

	assert.Equal(t, len(rulesBeforeAdd), len(rulesAfterRemove),
		"IP rules were not restored to the previous state after removal.")

	for ruleIndex, rule := range rulesBeforeAdd {
		assert.Equal(t, rule, rulesAfterRemove[ruleIndex],
			"IP rules were not restored to the previous state after removal.")
	}
}
