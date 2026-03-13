import 'package:nordvpn/widgets/adaptive_scaffold/flutter_adaptive_scaffold.dart';

final class AppBreakpoints {
  AppBreakpoints._();

  static const Breakpoint small = Breakpoint(
    beginWidth: 0,
    endWidth: 500,
    beginHeight: 532,
    endHeight: 532,
  );

  static const Breakpoint medium = Breakpoint(
    beginWidth: 500,
    endWidth: 700,
    beginHeight: 532,
    endHeight: 532,
  );

  static const Breakpoint large = Breakpoint(
    beginWidth: 700,
    endWidth: null,
    beginHeight: 532,
    endHeight: 532,
  );
}
