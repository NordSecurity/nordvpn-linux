import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import '../../test/utils/test_helpers.dart';

void runWarmupTests() async {
  group("warmup", () {
    // NOTE: The purpose of this test is to act as a shader warmup.
    // See: https://docs.flutter.dev/perf/shader
    testWidgets("test basic app interaction", (tester) async {
      await tester.setupIntegrationTests();

      await tester.pumpUntilFound(find.text(t.ui.rejectNonEssential));
      await tester.tap(find.text(t.ui.rejectNonEssential));

      await tester.pumpUntilFound(find.text(t.ui.logIn));
      await tester.tap(find.text(t.ui.logIn));

      await tester.pumpUntilFound(find.text(t.ui.quickConnect));
      await tester.tap(find.text(t.ui.quickConnect));

      await tester.pumpUntilFound(find.text(t.ui.disconnect));
      await tester.tap(find.text(t.ui.disconnect));
      await tester.pumpAndSettleWithTimeout();

      // settings screens
      await tester.tap(find.text(t.ui.settings));
      await tester.pumpAndSettleWithTimeout();

      await tester.tap(find.text(t.ui.general));
      await tester.pumpAndSettleWithTimeout();
    });
  });
}
