package iprule

import (
	"net"
	"slices"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/stretchr/testify/assert"
)

func TestFwmarkRule(t *testing.T) {
	category.Set(t, category.Route)

	prioID, err := calculateRulePriority()
	assert.NoError(t, err)
	assert.Greater(t, prioID, uint(0))

	tblID, err := calculateCustomTableID()
	assert.NoError(t, err)
	assert.Greater(t, tblID, uint(0))

	err = addFwmarkRule(0, prioID, tblID)
	assert.Error(t, err)

	var fwmarkval uint32 = 0xe1f1
	assert.Greater(t, fwmarkval, uint32(0))

	err = addFwmarkRule(fwmarkval, prioID, tblID)
	assert.NoError(t, err)

	fndTblID, err := findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, tblID, fndTblID)

	err = removeFwmarkRule(fwmarkval)
	assert.NoError(t, err)

	fndTblID, err = findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), fndTblID)
}

func TestMultiFwmarkRule(t *testing.T) {
	category.Set(t, category.Route)

	prioID, err := calculateRulePriority()
	assert.NoError(t, err)
	assert.Greater(t, prioID, uint(0))

	tblID, err := calculateCustomTableID()
	assert.NoError(t, err)
	assert.Greater(t, tblID, uint(0))

	var fwmarkval uint32 = 0xe1f1
	assert.Greater(t, fwmarkval, uint32(0))

	// add 1st rule
	err = addFwmarkRule(fwmarkval, prioID, tblID)
	assert.NoError(t, err)

	// check
	fndTblID, err := findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, tblID, fndTblID)

	prioID2, err := calculateRulePriority()
	assert.NoError(t, err)
	assert.Greater(t, prioID2, uint(0))

	tblID2, err := calculateCustomTableID()
	assert.NoError(t, err)
	assert.Greater(t, tblID2, uint(0))

	// add 2nd rule
	err = addFwmarkRule(fwmarkval, prioID2, tblID2)
	assert.NoError(t, err)

	// check
	fndTblID, err = findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, tblID, fndTblID)

	// will remove one
	err = removeFwmarkRule(fwmarkval)
	assert.NoError(t, err)

	fndTblID, err = findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, tblID2, fndTblID)

	err = removeFwmarkRule(fwmarkval)
	assert.NoError(t, err)

	fndTblID, err = findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), fndTblID)
}

func TestSuppressRule(t *testing.T) {
	category.Set(t, category.Route)

	prioID, err := calculateRulePriority()
	assert.NoError(t, err)
	assert.Greater(t, prioID, uint(0))

	tblID, err := calculateCustomTableID()
	assert.NoError(t, err)
	assert.Greater(t, tblID, uint(0))

	var fwmarkval uint32 = 0xe1f1
	assert.Greater(t, fwmarkval, uint32(0))

	fndTblID, err := findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), fndTblID)

	err = addFwmarkRule(fwmarkval, prioID, tblID)
	assert.NoError(t, err)

	prioID2, err := calculateRulePriority()
	assert.NoError(t, err)
	assert.Greater(t, prioID2, uint(0))

	fnd, err := checkSuppressRule()
	assert.NoError(t, err)
	assert.False(t, fnd)

	err = addSuppressRule(prioID2, true)
	assert.NoError(t, err)

	fnd, err = checkSuppressRule()
	assert.NoError(t, err)
	assert.True(t, fnd)

	err = removeSuppressRule()
	assert.NoError(t, err)

	err = addSuppressRule(prioID2, false)
	assert.NoError(t, err)

	fnd, err = checkSuppressRule()
	assert.NoError(t, err)
	assert.True(t, fnd)

	err = removeSuppressRule()
	assert.NoError(t, err)

	fnd, err = checkSuppressRule()
	assert.NoError(t, err)
	assert.False(t, fnd)

	err = removeFwmarkRule(fwmarkval)
	assert.NoError(t, err)

	fndTblID, err = findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), fndTblID)
}

func TestCustomTable(t *testing.T) {
	category.Set(t, category.Route)

	prioID, err := calculateRulePriority()
	assert.NoError(t, err)
	assert.Greater(t, prioID, uint(0))

	tblID, err := calculateCustomTableID()
	assert.NoError(t, err)
	assert.Greater(t, tblID, uint(0))

	var fwmarkval uint32 = 0xe1f1
	assert.Greater(t, fwmarkval, uint32(0))

	err = addFwmarkRule(fwmarkval, prioID, tblID)
	assert.NoError(t, err)

	fndTblID, err := findFwmarkRule(fwmarkval)
	assert.NoError(t, err)
	assert.Equal(t, tblID, fndTblID)

	// prev prio id is in use, so, next one should be less (less is higher prio)
	prioID2, err := calculateRulePriority()
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

	err = router.addAllowlistRules(allowlist)
	assert.NoError(t, err, "Unexpected error when adding allowlist rules: %s", err)

	rulesAfterAdd, err := netlink.RuleList(netlink.FAMILY_V4)

	assert.NoError(t, err, "Failed to save new state for ip rules: %s", err)
	// ignore rules that existed before the test for validation purposes
	rulesAfterAdd = slices.DeleteFunc(rulesAfterAdd, func(rule netlink.Rule) bool {
		idx := slices.Index(rulesBeforeAdd, rule)
		return idx != -1
	})

	expectedSubnets := []*net.IPNet{
		nil, // in case of netlink, nil is equivalent to default subnet(35.74.174.235/0 and 4.246.215.86/0)
		{
			IP:   net.IPv4(103, 238, 215, 0),
			Mask: net.CIDRMask(24, 32),
		},
		{
			IP:   net.IPv4(237, 164, 3, 235),
			Mask: net.CIDRMask(32, 32),
		},
	}

	for _, subnet := range expectedSubnets {
		ruleIdx := slices.IndexFunc(rulesAfterAdd, func(rule netlink.Rule) bool {
			return subnet.String() == rule.Dst.String()
		})
		assert.NotEqual(t, -1, ruleIdx, "Desired subnet %s not added to ip rules.", subnet.String())
	}

	router.removeAllowSubnetRules()
	rulesAfterRemove, err := netlink.RuleList(netlink.FAMILY_V4)
	assert.NoError(t, err, "Failed to save prevous state for ip rules: %s", err)

	assert.Equal(t, len(rulesBeforeAdd), len(rulesAfterRemove),
		"IP rules were not restored to the previous state after removal.")

	for ruleIndex, rule := range rulesBeforeAdd {
		assert.Equal(t, rule, rulesAfterRemove[ruleIndex],
			"IP rules were not restored to the previous state after removal.")
	}
}

func rule(opts ...func(*netlink.Rule)) netlink.Rule {
	r := netlink.NewRule()
	for _, opt := range opts {
		opt(r)
	}
	return *r
}

func withPriority(p int) func(*netlink.Rule) {
	return func(r *netlink.Rule) {
		r.Priority = p
	}
}

func withTable(t int) func(*netlink.Rule) {
	return func(r *netlink.Rule) {
		r.Table = t
	}
}

func withMark(m int) func(*netlink.Rule) {
	return func(r *netlink.Rule) {
		r.Mark = m
	}
}

func withCloudflareSource() func(*netlink.Rule) {
	ipNet := &net.IPNet{
		IP:   net.ParseIP("1.1.1.1"),
		Mask: net.CIDRMask(32, 32),
	}
	return func(r *netlink.Rule) {
		r.Src = ipNet
	}
}

func TestFindRulePriorityCandidate(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		rules       []netlink.Rule
		expectError bool
		expectedID  uint
	}{
		{
			name: "misconfig only local rule",
			rules: []netlink.Rule{
				// The version of netlink used by us returns -1 instead of 0 for local rule
				rule(withPriority(-1), withTable(unix.RT_TABLE_LOCAL)),
			},
			expectError: true,
		},
		{
			name: "misconfig main rule with too high priority",
			rules: []netlink.Rule{
				rule(withPriority(-1), withTable(unix.RT_TABLE_LOCAL)),
				rule(withPriority(1), withTable(unix.RT_TABLE_MAIN)),
			},
			expectError: true,
		},
		{
			name: "local and main rules",
			rules: []netlink.Rule{
				rule(withPriority(-1), withTable(unix.RT_TABLE_LOCAL)),
				rule(withPriority(32766), withTable(unix.RT_TABLE_MAIN)),
			},
			expectError: false,
			expectedID:  32765,
		},
		{
			name: "local main and default rules",
			rules: []netlink.Rule{
				rule(withPriority(-1), withTable(unix.RT_TABLE_LOCAL)),
				rule(withPriority(32766), withTable(unix.RT_TABLE_MAIN)),
				rule(withPriority(32767), withTable(unix.RT_TABLE_DEFAULT)),
			},
			expectError: false,
			expectedID:  32765,
		},
		{
			name: "different configs 1",
			rules: []netlink.Rule{
				rule(withPriority(-1), withTable(unix.RT_TABLE_LOCAL)),
				rule(withPriority(2), withTable(unix.RT_TABLE_MAIN), withCloudflareSource()),
				rule(withPriority(3), withTable(200), withCloudflareSource()),
				rule(withPriority(4), withTable(unix.RT_TABLE_MAIN), withMark(0xAA)),

				rule(withPriority(32766), withTable(unix.RT_TABLE_MAIN)),
				rule(withPriority(32767), withTable(unix.RT_TABLE_DEFAULT)),
			},
			expectError: false,
			expectedID:  32765,
		},
		{
			name: "different configs 2",
			rules: []netlink.Rule{
				rule(withPriority(-1), withTable(unix.RT_TABLE_LOCAL)),
				rule(withPriority(2), withTable(unix.RT_TABLE_MAIN), withCloudflareSource()),
				rule(withPriority(3), withTable(200), withCloudflareSource()),
				rule(withPriority(4), withTable(unix.RT_TABLE_MAIN), withMark(0xAA)),

				rule(withPriority(32766), withTable(unix.RT_TABLE_MAIN)),
				rule(withPriority(32767), withTable(unix.RT_TABLE_DEFAULT)),

				rule(withPriority(32770), withTable(unix.RT_TABLE_MAIN), withMark(0xBB)),
			},
			expectError: false,
			expectedID:  32765,
		},
		{
			name: "different configs 3",
			rules: []netlink.Rule{
				rule(withPriority(-1), withTable(unix.RT_TABLE_LOCAL)),
				rule(withPriority(2), withTable(unix.RT_TABLE_MAIN), withCloudflareSource()),
				rule(withPriority(3), withTable(200), withCloudflareSource()),
				rule(withPriority(4), withTable(unix.RT_TABLE_MAIN), withMark(0xAA)),

				rule(withPriority(32764), withTable(unix.RT_TABLE_MAIN), withMark(0xAB)),
				rule(withPriority(32765), withTable(unix.RT_TABLE_MAIN), withMark(0xAC)),
				rule(withPriority(32766), withTable(unix.RT_TABLE_MAIN)),
				rule(withPriority(32767), withTable(unix.RT_TABLE_DEFAULT)),

				rule(withPriority(32770), withTable(unix.RT_TABLE_MAIN), withMark(0xBB)),
			},
			expectError: false,
			expectedID:  32763,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := findRulePriorityCandidate(tt.rules)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestIsFromAllLookupMainRule(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		testRule netlink.Rule
		expected bool
	}{
		{
			name:     "main rule",
			testRule: rule(withPriority(32766), withTable(unix.RT_TABLE_MAIN)),
			expected: true,
		},
		{
			name:     "main rule with different priority",
			testRule: rule(withPriority(100), withTable(unix.RT_TABLE_MAIN)),
			expected: true,
		},
		{
			name:     "different rule 1",
			testRule: rule(withPriority(-1), withTable(unix.RT_TABLE_LOCAL)),
			expected: false,
		},
		{
			name:     "different rule 2",
			testRule: rule(withPriority(32766), withTable(unix.RT_TABLE_MAIN), withCloudflareSource()),
			expected: false,
		},
		{
			name:     "different rule 3",
			testRule: rule(withPriority(32766), withTable(unix.RT_TABLE_MAIN), withMark(0xAA)),
			expected: false,
		},
		{
			name:     "different rule 4",
			testRule: rule(withPriority(100), withTable(unix.RT_TABLE_MAIN), withCloudflareSource()),
			expected: false,
		},
		{
			name:     "different rule 5",
			testRule: rule(withPriority(32766), withTable(unix.RT_TABLE_DEFAULT), withCloudflareSource()),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isFromAllLookupMainRule(tt.testRule))
		})
	}
}
