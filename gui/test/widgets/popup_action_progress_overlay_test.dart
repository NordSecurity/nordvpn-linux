import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/providers/popup_actions_provider.dart';
import 'package:nordvpn/theme/theme.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/popup_action_progress_overlay.dart';

void main() {
  const popupId = 4242;
  const pageText = "page content";

  Future<ProviderContainer> pumpOverlay(WidgetTester tester) async {
    final container = ProviderContainer();
    addTearDown(container.dispose);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: MaterialApp(
          theme: lightTheme(),
          home: const Scaffold(
            body: PopupActionProgressOverlay(child: Text(pageText)),
          ),
        ),
      ),
    );

    return container;
  }

  group('PopupActionProgressOverlay', () {
    testWidgets('indicator is not displayed for fast actions', (tester) async {
      final container = await pumpOverlay(tester);

      await container
          .read(popupActionsProvider.notifier)
          .run((_) async {}, popupId: popupId);
      await tester.pump(PopupActions.indicatorDelay * 2);

      expect(find.byType(LoadingIndicator), findsNothing);
      expect(find.text(pageText), findsOneWidget);
    });

    testWidgets('indicator is displayed while a slow action runs', (
      tester,
    ) async {
      final container = await pumpOverlay(tester);
      final completer = Completer<void>();

      final done = container.read(popupActionsProvider.notifier).run(
        (_) => completer.future,
        popupId: popupId,
        progressMessage: "applying",
      );

      await tester.pump(
        PopupActions.indicatorDelay - const Duration(milliseconds: 1),
      );
      expect(find.byType(LoadingIndicator), findsNothing);
      final barriersWithoutIndicator = find
          .byType(ModalBarrier)
          .evaluate()
          .length;

      await tester.pump(const Duration(milliseconds: 2));
      expect(find.byType(LoadingIndicator), findsOneWidget);
      expect(find.text("applying"), findsOneWidget);
      // The input is blocked while the change is applied.
      expect(
        find.byType(ModalBarrier).evaluate().length,
        barriersWithoutIndicator + 1,
      );

      completer.complete();
      // The indicator stays until it was displayed long enough.
      await tester.pump();
      expect(find.byType(LoadingIndicator), findsOneWidget);

      await tester.pump(PopupActions.minIndicatorDuration);
      await tester.pump();
      expect(find.byType(LoadingIndicator), findsNothing);

      await done;
    });

    testWidgets('indicator without message displays only the spinner', (
      tester,
    ) async {
      final container = await pumpOverlay(tester);
      final completer = Completer<void>();

      final done = container
          .read(popupActionsProvider.notifier)
          .run((_) => completer.future, popupId: popupId);

      await tester.pump(PopupActions.indicatorDelay * 2);
      expect(find.byType(LoadingIndicator), findsOneWidget);
      expect(
        find.descendant(
          of: find.byType(LoadingIndicator),
          matching: find.byType(Text),
        ),
        findsNothing,
      );

      completer.complete();
      await tester.pump(PopupActions.minIndicatorDuration);
      await tester.pump();
      await done;
    });
  });
}
