package iprule

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

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
