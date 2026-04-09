import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';

enum PauseLength {
  // values in seconds
  mins5(5*60),
  mins15(15*60),
  mins30(30*60),
  hour1(60*60),
  hours24(24*60*60);
  
  const PauseLength(this.value);
  final int value;

  UIEvent_ItemValue toUIEventItemValue() {
    return switch (this) {
      PauseLength.mins5 => UIEvent_ItemValue.PAUSE_5_MIN,
      PauseLength.mins15 => UIEvent_ItemValue.PAUSE_15_MIN,
      PauseLength.mins30 => UIEvent_ItemValue.PAUSE_30_MIN,
      PauseLength.hour1 => UIEvent_ItemValue.PAUSE_1_HOUR,
      PauseLength.hours24 => UIEvent_ItemValue.PAUSE_24_HOURS,
    };
  }
}
