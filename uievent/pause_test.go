package uievent

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestPauseDurationToItemValue(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		seconds  uint32
		expected pb.UIEvent_ItemValue
	}{
		{
			name:     "5 minutes",
			seconds:  300,
			expected: pb.UIEvent_PAUSE_5_MIN,
		},
		{
			name:     "15 minutes",
			seconds:  900,
			expected: pb.UIEvent_PAUSE_15_MIN,
		},
		{
			name:     "30 minutes",
			seconds:  1800,
			expected: pb.UIEvent_PAUSE_30_MIN,
		},
		{
			name:     "1 hour",
			seconds:  3600,
			expected: pb.UIEvent_PAUSE_1_HOUR,
		},
		{
			name:     "24 hours",
			seconds:  86400,
			expected: pb.UIEvent_PAUSE_24_HOURS,
		},
		{
			name:     "zero",
			seconds:  0,
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name:     "one second below 5 minutes",
			seconds:  299,
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name:     "one second above 5 minutes",
			seconds:  301,
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name:     "arbitrary duration",
			seconds:  7200,
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
		{
			name:     "above 24 hours",
			seconds:  86401,
			expected: pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, PauseDurationToItemValue(tt.seconds))
		})
	}
}
