package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPauseArgumentsParsing(t *testing.T) {
	tests := []struct {
		name                        string
		pauseDuration               string
		expectedPauseDurationResult uint32
		shouldReturnError           bool
	}{
		{
			name:                        "success 5m",
			pauseDuration:               "5m",
			expectedPauseDurationResult: 300,
		},
		{
			name:                        "success 15m",
			pauseDuration:               "15m",
			expectedPauseDurationResult: 900,
		},
		{
			name:                        "success 30m",
			pauseDuration:               "30m",
			expectedPauseDurationResult: 1800,
		},
		{
			name:                        "success 1h",
			pauseDuration:               "1h",
			expectedPauseDurationResult: 3600,
		},
		{
			name:                        "success 24h",
			pauseDuration:               "24h",
			expectedPauseDurationResult: 86400,
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

			assert.Equal(t, test.expectedPauseDurationResult, result)
			if test.shouldReturnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
