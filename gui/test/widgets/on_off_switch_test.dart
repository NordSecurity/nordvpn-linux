import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/aurora_design.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';
import '../utils/test_helpers.dart';

void main() {
  setUpAll(() async {
    final FakeSharedPreferencesAsync store = FakeSharedPreferencesAsync();
    SharedPreferencesAsyncPlatform.instance = store;
    await initServiceLocator();
  });
  final design = AppDesign(ThemeMode.light);
  group('OnOffSwitch Widget Tests', () {
    Future<void> noOp(bool _) async {}
    Future<void> longRunning(bool _) async {
      await Future.delayed(const Duration(milliseconds: 200));
    }

    testWidgets('Initial state is off', (tester) async {
      // Build the widget.
      await tester.setupWidgetTest(OnOffSwitch(onChanged: noOp));

      // Verify initial label is "Off".
      expect(find.text('Off'), findsOneWidget);
      expect(find.text('On'), findsNothing);

      // Verify the initial state of the switch.
      final switchWidget = tester.widget<AnimatedContainer>(
        find.byType(AnimatedContainer).first,
      );
      expect(
        switchWidget.decoration,
        isA<BoxDecoration>().having(
          (d) => d.color,
          'color',
          design.semanticColors.bgSecondary,
        ),
      );
    });

    testWidgets('Clicking the switch changes its state to on', (tester) async {
      await tester.setupWidgetTest(OnOffSwitch(onChanged: noOp));

      // Tap the switch.
      await tester.tap(find.byType(GestureDetector));
      await tester.pumpUntilFound(find.text('On'));

      // Verify the label changes to "On".
      expect(find.text('On'), findsOneWidget);
      expect(find.text('Off'), findsNothing);

      // Verify the state of the switch is "On".
      final switchWidget = tester.widget<AnimatedContainer>(
        find.byType(AnimatedContainer).first,
      );
      expect(
        switchWidget.decoration,
        isA<BoxDecoration>().having(
          (d) => d.color,
          'color',
          design.semanticColors.bgAccent,
        ),
      );
    });

    testWidgets('Clicking again changes its state to off', (tester) async {
      await tester.setupWidgetTest(OnOffSwitch(onChanged: noOp));

      // Tap the switch to turn it on.
      await tester.tap(find.byType(GestureDetector));
      await tester.pumpUntilFound(find.text('On'));

      // Tap the switch to turn it off.
      await tester.tap(find.byType(GestureDetector));
      await tester.pumpUntilFound(find.text('Off'));

      // Verify the label changes back to "Off".
      expect(find.text('Off'), findsOneWidget);
      expect(find.text('On'), findsNothing);

      // Verify the state of the switch is "Off".
      final switchWidget = tester.widget<AnimatedContainer>(
        find.byType(AnimatedContainer).first,
      );
      expect(
        switchWidget.decoration,
        isA<BoxDecoration>().having(
          (d) => d.color,
          'color',
          design.semanticColors.bgSecondary,
        ),
      );
    });

    testWidgets('Clicking on switch triggers callback', (tester) async {
      bool toggleValue = false;
      // ignore: prefer_function_declarations_over_variables
      final onChanged = (_) async => toggleValue = !toggleValue;
      await tester.setupWidgetTest(OnOffSwitch(onChanged: onChanged));

      // Tap the switch
      await tester.tap(find.byType(GestureDetector));

      // Verify the callback was called
      expect(toggleValue, isTrue);

      // Tap the switch again
      await tester.tap(find.byType(GestureDetector));

      // Verify the toggle
      expect(toggleValue, isFalse);
    });

    testWidgets('Loading indicator appears for long callback', (tester) async {
      await tester.setupWidgetTest(OnOffSwitch(onChanged: longRunning));

      // Tap the switch.
      await tester.tap(find.byType(GestureDetector));
      await tester.pumpUntilFound(find.byType(CircularProgressIndicator));
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      await tester.pumpUntilFound(find.text('On'));

      // Verify the label changes to "On".
      expect(find.text('On'), findsOneWidget);
      expect(find.text('Off'), findsNothing);

      await tester.pumpUntilFound(find.byType(AnimatedPositioned));
      // Verify the state of the switch is "On".
      final switchWidget = tester.widget<AnimatedContainer>(
        find.byType(AnimatedContainer).first,
      );
      expect(
        switchWidget.decoration,
        isA<BoxDecoration>().having(
          (d) => d.color,
          'color',
          design.semanticColors.bgAccent,
        ),
      );
    });

    testWidgets('Label is changed when loading is finished', (tester) async {
      await tester.setupWidgetTest(OnOffSwitch(onChanged: longRunning));

      // Tap the switch.
      await tester.tap(find.byType(GestureDetector));
      await tester.pumpUntilFound(find.byType(CircularProgressIndicator));
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      await tester.pumpUntilFound(find.text('On'));

      // The label changes when the loader is finished
      expect(find.byType(CircularProgressIndicator), findsNothing);
      // Verify the label changed to "On".
      expect(find.text('On'), findsOneWidget);
      expect(find.text('Off'), findsNothing);
    });
  });
}
