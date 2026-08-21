package daemon

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestTruncateMiddle(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		testName       string
		input          string
		charsToShow    int
		expectedString string
	}{
		{
			testName:       "regular token",
			input:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			charsToShow:    edgeCharsToShow,
			expectedString: "012345...abcdef",
		},
		{
			testName:       "invalid token",
			input:          "123",
			charsToShow:    edgeCharsToShow,
			expectedString: "Invalid",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			result := truncateMiddle(test.input, test.charsToShow)
			assert.Equal(t, result, test.expectedString)
		})
	}
}
