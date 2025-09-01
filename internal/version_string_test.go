package internal

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestCleanUpVersionString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		versionString  string
		expectedResult string
		shouldFail     bool
	}{
		{
			name:           "string after version string",
			versionString:  "3.20.1+b05658b99a42e01d",
			expectedResult: "3.20.1",
		},
		{
			name:           "string before version string",
			versionString:  "b05658b99a42e01d+3.20.1",
			expectedResult: "3.20.1",
		},
		{
			name:           "string before and after version string",
			versionString:  "b05658b99a42e01d+3.20.1+b05658b99a42e01d",
			expectedResult: "3.20.1",
		},
		{
			name:           "string before and after version string no separator",
			versionString:  "b05658b99a42e01d3.20.1b05658b99a42e01d",
			expectedResult: "3.20.1",
		},
		{
			name:           "no extra strings",
			versionString:  "3.20.1",
			expectedResult: "3.20.1",
		},
		{
			name:          "garbage string",
			versionString: "aaaaa",
			shouldFail:    true,
		},
		{
			name:          "invalid version string",
			versionString: "3.aaa.1",
			shouldFail:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := CleanUpVersionString(test.versionString)

			assert.Equal(t, test.expectedResult, result)

			if test.shouldFail {
				assert.Error(t, err, "CleanUpVersionString returned an unexpected error.")
			}
		})
	}
}
