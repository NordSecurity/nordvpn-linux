import 'package:flutter_test/flutter_test.dart';

import 'app_ctl.dart';
import 'test_helpers.dart';

abstract class ScreenHandle {
  final AppCtl app;

  ScreenHandle(this.app);

  Future<void> waitUntilFound(
    Finder finder, {
    Duration timeout = const Duration(seconds: 5),
    Duration interval = const Duration(milliseconds: 100),
    Matcher matcher = findsOneWidget,
    bool finishAnimations = true,
  }) async {
    await app.tester.pumpUntilFound(
      finder,
      timeout: timeout,
      interval: interval,
      matcher: matcher,
    );
    if (finishAnimations) {
      await app.tester.pumpAndSettle();
    }
  }

  Future<void> waitFor(
    bool Function() condition, {
    Duration timeout = const Duration(seconds: 5),
    Duration interval = const Duration(milliseconds: 100),
    bool finishAnimations = true,
  }) async {
    await app.tester.pumpUntilTrue(
      condition,
      timeout: timeout,
      interval: interval,
    );
    if (finishAnimations) {
      await app.tester.pumpAndSettle();
    }
  }
}
