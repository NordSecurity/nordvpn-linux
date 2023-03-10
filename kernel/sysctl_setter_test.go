package kernel

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestSet_rp_filter(t *testing.T) {
	category.Set(t, category.Root)

	param := "net.ipv4.conf.all.rp_filter"

	paramVal, err := Parameter(param)
	assert.NoError(t, err)
	fmt.Printf("~~~ 1Param[%s]=%d\n", param, paramVal[param])

	desiredVal := 2
	unwantedVal := 0

	setter := NewSysctlSetter(param, desiredVal, unwantedVal)

	err = setter.Set()
	assert.NoError(t, err)

	paramValA, err := Parameter(param)
	assert.NoError(t, err)
	fmt.Printf("~~~ 1aParam[%s]=%d\n", param, paramValA[param])
	assert.Equal(t, desiredVal, paramValA[param])

	err = setter.Unset()
	assert.NoError(t, err)

	paramValB, err := Parameter(param)
	assert.NoError(t, err)
	fmt.Printf("~~~ 1bParam[%s]=%d\n", param, paramValB[param])
	assert.Equal(t, unwantedVal, paramValB[param])

	// restore original value
	err = SetParameter(param, paramVal[param])
	assert.NoError(t, err)

	paramVal2, err := Parameter(param)
	fmt.Printf("~~~ 2Param[%s]=%d\n", param, paramVal[param])
	assert.NoError(t, err)
	assert.Equal(t, paramVal[param], paramVal2[param])
}
