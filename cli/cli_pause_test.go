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
			name:           "success 1h",
			pauseDuration:  "1h",
			expectedResult: 3600,
		},
		{
			name:           "success 5m",
			pauseDuration:  "5m",
			expectedResult: 300,
		},
		{
			name:              "invalid interval",
			pauseDuration:     "17m",
			shouldReturnError: true,
		},
		{
			name:              "invalid argument",
			pauseDuration:     "aaaaa",
			shouldReturnError: true,
		},
		{
			name:              "invalid argument(no value)",
			pauseDuration:     "",
			shouldReturnError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := pauseArgToDuration(test.pauseDuration)

			assert.Equal(t, test.expectedResult, result)
			if test.shouldReturnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
