package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPauseArgumentsParsing(t *testing.T) {
	tests := []struct {
		name              string
		pauseDuration     string
		expectedResult    uint32
		shouldReturnError bool
	}{
		{
			name:           "success hours",
			pauseDuration:  "5h",
			expectedResult: 18000,
		},
		{
			name:           "success minutes",
			pauseDuration:  "30m",
			expectedResult: 1800,
		},
		{
			name:              "invalid argument(unit in front)",
			pauseDuration:     "m30",
			shouldReturnError: true,
		},
		{
			name:              "invalid argument(unit not recognized)",
			pauseDuration:     "30s",
			shouldReturnError: true,
		},
		{
			name:              "invalid argument(no value)",
			pauseDuration:     "m",
			shouldReturnError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parsePauseArg(test.pauseDuration)

			assert.Equal(t, test.expectedResult, result)
			if test.shouldReturnError {
				assert.Error(t, err)
			}
		})
	}
}
