package cli

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/stretchr/testify/assert"
)

func TestPauseArgumentsParsing(t *testing.T) {
	tests := []struct {
		name                        string
		pauseDuration               string
		expectedPauseDurationResult uint32
		expectedUIEventResult       pb.UIEvent_ItemValue
		shouldReturnError           bool
	}{
		{
			name:                        "success 5m",
			pauseDuration:               "5m",
			expectedPauseDurationResult: 300,
			expectedUIEventResult:       pb.UIEvent_PAUSE_5_MIN,
		},
		{
			name:                        "success 15m",
			pauseDuration:               "15m",
			expectedPauseDurationResult: 900,
			expectedUIEventResult:       pb.UIEvent_PAUSE_5_MIN,
		},
		{
			name:                        "success 30m",
			pauseDuration:               "30m",
			expectedPauseDurationResult: 1800,
			expectedUIEventResult:       pb.UIEvent_PAUSE_5_MIN,
		},
		{
			name:                        "success 1h",
			pauseDuration:               "1h",
			expectedPauseDurationResult: 3600,
			expectedUIEventResult:       pb.UIEvent_PAUSE_1_HOUR,
		},
		{
			name:                        "success 24h",
			pauseDuration:               "24h",
			expectedPauseDurationResult: 86400,
			expectedUIEventResult:       pb.UIEvent_PAUSE_24_HOURS,
		},
		{
			name:                  "invalid interval",
			pauseDuration:         "17m",
			shouldReturnError:     true,
			expectedUIEventResult: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name:                  "invalid argument",
			pauseDuration:         "aaaaa",
			shouldReturnError:     true,
			expectedUIEventResult: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name:                  "invalid argument(no value)",
			pauseDuration:         "",
			shouldReturnError:     true,
			expectedUIEventResult: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := pauseArgToInterval(test.pauseDuration)

			assert.Equal(t, test.expectedPauseDurationResult, result)
			if test.shouldReturnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
