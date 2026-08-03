package pause

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

func PauseIntervalToPauseUIEventItemValue(interval pb.PauseInverval) pb.UIEvent_ItemValue {
	switch interval {
	case pb.PauseInverval_PAUSE_5_MIN:
		return pb.UIEvent_PAUSE_5_MIN
	case pb.PauseInverval_PAUSE_15_MIN:
		return pb.UIEvent_PAUSE_15_MIN
	case pb.PauseInverval_PAUSE_30_MIN:
		return pb.UIEvent_PAUSE_30_MIN
	case pb.PauseInverval_PAUSE_1_HOUR:
		return pb.UIEvent_PAUSE_1_HOUR
	case pb.PauseInverval_PAUSE_24_HOURS:
		return pb.UIEvent_PAUSE_24_HOURS
	default:
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}
}
