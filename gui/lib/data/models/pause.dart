import 'package:nordvpn/pb/daemon/pause.pb.dart';
import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';

enum PauseLength {
  // values in seconds
  mins5(PauseInverval.PAUSE_5_MIN, UIEvent_ItemValue.PAUSE_5_MIN),
  mins15(PauseInverval.PAUSE_15_MIN, UIEvent_ItemValue.PAUSE_15_MIN),
  mins30(PauseInverval.PAUSE_30_MIN, UIEvent_ItemValue.PAUSE_30_MIN),
  hour1(PauseInverval.PAUSE_1_HOUR, UIEvent_ItemValue.PAUSE_1_HOUR),
  hours24(PauseInverval.PAUSE_24_HOURS, UIEvent_ItemValue.PAUSE_24_HOURS);

  const PauseLength(this.interval, this.eventValue);
  final PauseInverval interval;
  final UIEvent_ItemValue eventValue;
}
