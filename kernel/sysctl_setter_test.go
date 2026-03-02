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
	originalVal := paramVal[param]

	setter := NewSysctlSetter(param, desiredVal)

	err = setter.Set()
	assert.NoError(t, err)

	paramValA, err := Parameter(param)
	assert.NoError(t, err)
	fmt.Printf("~~~ 1aParam[%s]=%d\n", param, paramValA[param])
	assert.Equal(t, desiredVal, paramValA[param])
	// restore original value
	err = setter.Unset()
	assert.NoError(t, err)

	paramValB, err := Parameter(param)
	assert.NoError(t, err)
	fmt.Printf("~~~ 1bParam[%s]=%d\n", param, paramValB[param])
	assert.Equal(t, originalVal, paramValB[param])
}
