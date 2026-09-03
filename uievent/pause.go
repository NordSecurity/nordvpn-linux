package uievent

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// PauseDurationToItemValue maps a pause duration in seconds to the
// corresponding UIEvent item value. Durations which do not match one of
// the predefined pause options resolve to ITEM_VALUE_UNSPECIFIED.
func PauseDurationToItemValue(seconds uint32) pb.UIEvent_ItemValue {
	switch seconds {
	case internal.PauseSeconds5Min:
		return pb.UIEvent_PAUSE_5_MIN
	case internal.PauseSeconds15Min:
		return pb.UIEvent_PAUSE_15_MIN
	case internal.PauseSeconds30Min:
		return pb.UIEvent_PAUSE_30_MIN
	case internal.PauseSeconds1Hour:
		return pb.UIEvent_PAUSE_1_HOUR
	case internal.PauseSeconds24Hour:
		return pb.UIEvent_PAUSE_24_HOURS
	default:
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}
}
