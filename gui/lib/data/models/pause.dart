import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';

enum PauseLength {
  // values in seconds
  mins5(5 * 60, UIEvent_ItemValue.PAUSE_5_MIN),
  mins15(15 * 60, UIEvent_ItemValue.PAUSE_15_MIN),
  mins30(30 * 60, UIEvent_ItemValue.PAUSE_30_MIN),
  hour1(60 * 60, UIEvent_ItemValue.PAUSE_1_HOUR),
  hours24(24 * 60 * 60, UIEvent_ItemValue.PAUSE_24_HOURS);

  const PauseLength(this.seconds, this.eventValue);
  final int seconds;
  final UIEvent_ItemValue eventValue;
}
