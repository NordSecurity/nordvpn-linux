import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/widgets/toast.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';
import '../utils/test_helpers.dart';

void main() {
  const expectedMinutes = '00';
  const expectedZeroSeconds = '00';
  const expectedOneSecond = '01';
  const expectedTwoSeconds = '02';
  const expectedThreeSeconds = '03';

  setUpAll(() async {
    final fakeStore = FakeSharedPreferencesAsync();
    SharedPreferencesAsyncPlatform.instance = fakeStore;
    await initServiceLocator();
  });

  Widget buildToast({required Duration timeout, VoidCallback? onClose}) {
    return Toast(duration: timeout, onClose: onClose);
  }

  group('Toast', () {
    testWidgets('verify time format', (tester) async {
      await tester.setupWidgetTest(
        buildToast(timeout: const Duration(seconds: 1)),
      );

      final expectedText = t.ui.VPNResumesIn(
        minutes: expectedMinutes,
        seconds: expectedOneSecond,
      );
      expect(find.text(expectedText), findsOneWidget);
    });

    testWidgets('verify countdown updates every second', (tester) async {
      await tester.setupWidgetTest(
        buildToast(timeout: const Duration(seconds: 3)),
      );

      final initialExpectedText = t.ui.VPNResumesIn(
        minutes: expectedMinutes,
        seconds: expectedThreeSeconds,
      );
      expect(find.text(initialExpectedText), findsOneWidget);

      final expectedTextAfterSecond = t.ui.VPNResumesIn(
        minutes: expectedMinutes,
        seconds: expectedTwoSeconds,
      );
      await tester.pump(const Duration(seconds: 1));
      expect(find.text(expectedTextAfterSecond), findsOneWidget);

      final expectedTextAfterTwoSeconds = t.ui.VPNResumesIn(
        minutes: expectedMinutes,
        seconds: expectedOneSecond,
      );
      await tester.pump(const Duration(seconds: 1));
      expect(find.text(expectedTextAfterTwoSeconds), findsOneWidget);

      final expectedTextAfterThreeSeconds = t.ui.VPNResumesIn(
        minutes: expectedMinutes,
        seconds: expectedZeroSeconds,
      );
      await tester.pump(const Duration(seconds: 1));
      expect(find.text(expectedTextAfterThreeSeconds), findsOneWidget);
    });

    testWidgets('verify no countdown updates after timeout', (tester) async {
      await tester.setupWidgetTest(
        buildToast(timeout: const Duration(seconds: 1)),
      );

      final expectedText = t.ui.VPNResumesIn(
        minutes: expectedMinutes,
        seconds: expectedOneSecond,
      );
      expect(find.text(expectedText), findsOneWidget);

      final expectedTextAfterSecond = t.ui.VPNResumesIn(
        minutes: expectedMinutes,
        seconds: expectedZeroSeconds,
      );
      await tester.pump(const Duration(seconds: 1));
      expect(find.text(expectedTextAfterSecond), findsOneWidget);

      // timeout hit, no more updates shall occure from now on
      final expectedTextAfterTwoSeconds = t.ui.VPNResumesIn(
        minutes: expectedMinutes,
        seconds: expectedZeroSeconds,
      );
      await tester.pump(const Duration(seconds: 5));
      expect(find.text(expectedTextAfterTwoSeconds), findsOneWidget);
    });

    testWidgets('verify close calls callback', (tester) async {
      bool onCloseCalled = false;
      await tester.setupWidgetTest(
        buildToast(
          timeout: const Duration(seconds: 1),
          onClose: () {
            onCloseCalled = true;
          },
        ),
      );

      final svgFinder = tester.findSvgWithPath('toast_close_icon.svg');
      expect(svgFinder, findsOneWidget);

      final closeButtonFinder = find.ancestor(
        of: svgFinder,
        matching: find.byType(GestureDetector),
      );
      await tester.tap(closeButtonFinder);
      await tester.pump();

      expect(onCloseCalled, isTrue);
    });
  });

  group('Toast keyboard', () {
    for (final key in [
      LogicalKeyboardKey.enter,
      LogicalKeyboardKey.numpadEnter,
      LogicalKeyboardKey.space,
    ]) {
      testWidgets('$key closes the toast', (tester) async {
        var closed = false;
        await tester.setupWidgetTest(
          buildToast(
            timeout: const Duration(seconds: 5),
            onClose: () => closed = true,
          ),
        );
        await tester.pump(); // let postFrameCallback run focus
        await tester.sendKeyEvent(key);
        expect(closed, isTrue);
      });
    }

    for (final key in [
      LogicalKeyboardKey.arrowUp,
      LogicalKeyboardKey.arrowDown,
      LogicalKeyboardKey.arrowLeft,
      LogicalKeyboardKey.arrowRight,
      LogicalKeyboardKey.tab,
      LogicalKeyboardKey.escape,
      LogicalKeyboardKey.keyA,
      LogicalKeyboardKey.digit1,
    ]) {
      testWidgets('$key does not close the toast', (tester) async {
        var closed = false;
        await tester.setupWidgetTest(
          buildToast(
            timeout: const Duration(seconds: 5),
            onClose: () => closed = true,
          ),
        );
        await tester.pump();
        await tester.sendKeyEvent(key);
        expect(closed, isFalse);
      });
    }
  });

  group('Toast focus', () {
    testWidgets('close button gains focus after first frame', (tester) async {
      await tester.setupWidgetTest(
        buildToast(timeout: const Duration(seconds: 5), onClose: () {}),
      );
      await tester.pump();
      final closeFocus = tester
          .widgetList<Focus>(find.byType(Focus))
          .firstWhere((f) => f.focusNode?.debugLabel == 'ToastCloseButton');
      expect(closeFocus.focusNode!.hasFocus, isTrue);
    });
  });
}
