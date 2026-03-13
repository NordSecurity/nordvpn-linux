import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/theme/breakpoints.dart';


void main() {
  group('Window size constants', () {
    test('MinSize greater than small and medium breakpoint', () async {
      expect(windowMinSize.width, greaterThan(AppBreakpoints.small.endWidth!));
      expect(windowMinSize.height, greaterThan(AppBreakpoints.small.endHeight!));

      expect(windowMinSize.width, greaterThan(AppBreakpoints.medium.endWidth!));
      expect(windowMinSize.height, greaterThan(AppBreakpoints.medium.endHeight!));
    });
  });

  group('Window size constants', () {
    test('MinSize equals DefaultSize', () async {
      expect(windowMinSize.width, windowDefaultSize.width);
      expect(windowMinSize.height, windowDefaultSize.height);
    });
  });
}
