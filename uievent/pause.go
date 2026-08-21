package uievent

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// Pause durations, in seconds, supported by the pause UI event item values.
const (
	pauseSeconds5Min   uint32 = 5 * 60
	pauseSeconds15Min  uint32 = 15 * 60
	pauseSeconds30Min  uint32 = 30 * 60
	pauseSeconds1Hour  uint32 = 60 * 60
	pauseSeconds24Hour uint32 = 24 * 60 * 60
)

// PauseDurationToItemValue maps a pause duration in seconds to the
// corresponding UIEvent item value. Durations which do not match one of
// the predefined pause options resolve to ITEM_VALUE_UNSPECIFIED.
func PauseDurationToItemValue(seconds uint32) pb.UIEvent_ItemValue {
	switch seconds {
	case pauseSeconds5Min:
		return pb.UIEvent_PAUSE_5_MIN
	case pauseSeconds15Min:
		return pb.UIEvent_PAUSE_15_MIN
	case pauseSeconds30Min:
		return pb.UIEvent_PAUSE_30_MIN
	case pauseSeconds1Hour:
		return pb.UIEvent_PAUSE_1_HOUR
	case pauseSeconds24Hour:
		return pb.UIEvent_PAUSE_24_HOURS
	default:
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}
}
