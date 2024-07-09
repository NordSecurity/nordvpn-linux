package norduser

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_findVariable(t *testing.T) {
	category.Set(t, category.Unit)

	const variable1 string = "VAR_1"
	const value1 string = "val_1"

	const variable2 string = "VAR_2"
	const value2 string = "val_2"

	const emptyVariable string = "VAR_3"

	environment := fmt.Sprintf("%s=%s\000%s=%s\000%s=",
		variable1, value1,
		variable2, value2,
		emptyVariable)

	tests := []struct {
		name          string
		variableName  string
		expectedValue string
	}{
		{
			name:          "normal variable",
			variableName:  variable1,
			expectedValue: value1,
		},
		{
			name:          "empty variable",
			variableName:  emptyVariable,
			expectedValue: "",
		},
		{
			name:          "variable doesnt exist",
			variableName:  "NO_VARIABLE",
			expectedValue: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			variableValue := findVariable(test.variableName, environment)
			assert.Equal(t, test.expectedValue, variableValue)
		})
	}
}
