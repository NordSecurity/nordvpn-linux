import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/providers/popup_actions_provider.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/router/router.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/theme.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/popup_action_progress_overlay.dart';
import 'package:nordvpn/widgets/popups_listener.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';

void main() {
  const popupId = 987654;
  const popupTitle = "Decision popup";
  const popupMessage = "Do you want to apply this change?";
  const yesButton = "Yes";
  const noButton = "No";

  // Keeps the daemon connection in the loading state, so that the controllers
  // which the listener depends on don't try to reach the daemon.
  final noDaemonConnection = grpcConnectionControllerProvider.overrideWith(
    _NoDaemonConnection.new,
  );

  setUpAll(() async {
    SharedPreferencesAsyncPlatform.instance = FakeSharedPreferencesAsync();
    await initServiceLocator();
  });

  DecisionPopupMetadata decisionMetadata({
    required PopupAction yesAction,
    PopupAction? noAction,
    String? progressMessage,
  }) {
    return DecisionPopupMetadata(
      id: popupId,
      title: popupTitle,
      message: (_) => popupMessage,
      noButtonText: noButton,
      yesButtonText: yesButton,
      yesAction: yesAction,
      noAction: noAction,
      progressMessage: progressMessage,
    );
  }

  // Mirrors the production layout from main.dart: the listener and the progress
  // overlay are above the navigator, so that the dialogs are displayed on top
  // of them.
  Future<ProviderContainer> pumpListener(
    WidgetTester tester, {
    Widget page = const SizedBox.shrink(),
  }) async {
    final container = ProviderContainer(overrides: [noDaemonConnection]);
    addTearDown(container.dispose);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: MaterialApp(
          navigatorKey: goRouterKey,
          theme: lightTheme(),
          builder: (_, child) => Scaffold(
            body: PopupsListener(
              child: PopupActionProgressOverlay(child: child!),
            ),
          ),
          home: page,
        ),
      ),
    );
    await tester.pumpAndSettle();

    return container;
  }

  Future<void> showPopup(
    WidgetTester tester,
    ProviderContainer container,
    PopupMetadata metadata,
  ) async {
    container.read(popupsProvider.notifier).showWithMetadata(metadata);
    await tester.pumpAndSettle();
    expect(find.text(popupMessage), findsOneWidget);
  }

  // pumpAndSettle can't be used while the progress indicator is displayed,
  // because the spinner animates forever.
  const closeAnimation = Duration(milliseconds: 300);

  Future<void> pumpUntilIndicatorHidden(WidgetTester tester) async {
    await tester.pump();
    await tester.pump(PopupActions.minIndicatorDuration);
    await tester.pump(closeAnimation);
  }

  group('PopupsListener', () {
    // The popup is closed before the action starts, so the action outlives the
    // dialog. Running it with the WidgetRef of the popup used to throw
    // "Cannot use ref after the widget was disposed." after the first await.
    testWidgets('action using ref after an await does not crash', (
      tester,
    ) async {
      final container = await pumpListener(tester);
      final completer = Completer<void>();
      var readAfterAwait = false;

      await showPopup(
        tester,
        container,
        decisionMetadata(
          yesAction: (ref) async {
            await completer.future;
            ref.read(popupsProvider.notifier);
            readAfterAwait = true;
          },
        ),
      );

      await tester.tap(find.text(yesButton));
      await tester.pump();
      await tester.pump(closeAnimation);

      // The dialog is closed while the action is still running.
      expect(find.text(popupMessage), findsNothing);
      expect(container.read(popupActionsProvider), isNotNull);

      completer.complete();
      await pumpUntilIndicatorHidden(tester);

      expect(readAfterAwait, isTrue);
      expect(tester.takeException(), isNull);
      expect(container.read(popupActionsProvider), isNull);
    });

    testWidgets('progress indicator is displayed while the action runs', (
      tester,
    ) async {
      final container = await pumpListener(tester);
      final completer = Completer<void>();

      await showPopup(
        tester,
        container,
        decisionMetadata(
          yesAction: (_) => completer.future,
          progressMessage: "applying",
        ),
      );

      await tester.tap(find.text(yesButton));
      await tester.pump();
      await tester.pump(closeAnimation);

      expect(find.text(popupMessage), findsNothing);
      expect(find.byType(LoadingIndicator), findsOneWidget);
      expect(find.text("applying"), findsOneWidget);

      completer.complete();
      await pumpUntilIndicatorHidden(tester);

      expect(find.byType(LoadingIndicator), findsNothing);
    });

    // The error reported by the action must be displayed only after the progress
    // indicator is dismissed, never on top of it.
    testWidgets('error reported by the action is displayed after the progress', (
      tester,
    ) async {
      final container = await pumpListener(tester);
      final completer = Completer<void>();

      await showPopup(
        tester,
        container,
        decisionMetadata(
          yesAction: (ref) async {
            ref.read(popupsProvider.notifier).show(DaemonStatusCode.configError);
            await completer.future;
          },
        ),
      );

      await tester.tap(find.text(yesButton));
      await tester.pump();
      await tester.pump(closeAnimation);

      // Queued, but not displayed while the action is running.
      expect(find.byType(LoadingIndicator), findsOneWidget);
      expect(find.text(t.ui.settingsWereNotSaved), findsNothing);

      completer.complete();
      await pumpUntilIndicatorHidden(tester);

      expect(find.byType(LoadingIndicator), findsNothing);
      expect(find.text(t.ui.settingsWereNotSaved), findsOneWidget);

      // Close it, so that the queue is drained before the test ends.
      await tester.tap(find.text(t.ui.close));
      await tester.pumpAndSettle();
    });

    testWidgets('failing action displays the generic error popup', (
      tester,
    ) async {
      final container = await pumpListener(tester);

      await showPopup(
        tester,
        container,
        decisionMetadata(
          yesAction: (_) async => throw Exception("action failed"),
        ),
      );

      await tester.tap(find.text(yesButton));
      await tester.pumpAndSettle();

      expect(find.text(t.daemon.genericErrorTitle), findsOneWidget);

      await tester.tap(find.text(t.ui.gotIt));
      await tester.pumpAndSettle();
    });

    testWidgets('no button runs the no action', (tester) async {
      final container = await pumpListener(tester);
      var yesExecuted = false;
      var noExecuted = false;

      await showPopup(
        tester,
        container,
        decisionMetadata(
          yesAction: (_) => yesExecuted = true,
          noAction: (_) => noExecuted = true,
        ),
      );

      await tester.tap(find.text(noButton));
      await tester.pumpAndSettle();

      expect(noExecuted, isTrue);
      expect(yesExecuted, isFalse);
    });

    testWidgets('closing the popup with the X icon runs no action', (
      tester,
    ) async {
      final container = await pumpListener(tester);
      var yesExecuted = false;
      var noExecuted = false;

      await showPopup(
        tester,
        container,
        decisionMetadata(
          yesAction: (_) => yesExecuted = true,
          noAction: (_) => noExecuted = true,
        ),
      );

      await tester.tap(find.byType(IconButton));
      await tester.pumpAndSettle();

      expect(find.text(popupMessage), findsNothing);
      expect(yesExecuted, isFalse);
      expect(noExecuted, isFalse);
      expect(find.byType(LoadingIndicator), findsNothing);
    });

    // The popup is displayed as a side effect of build, so a rebuild while the
    // dialog is displayed must not open a second dialog for it.
    testWidgets('rebuild while the popup is displayed opens a single dialog', (
      tester,
    ) async {
      final container = await pumpListener(tester);

      await showPopup(
        tester,
        container,
        decisionMetadata(yesAction: (_) {}),
      );

      // Any provider change rebuilds the listener.
      container.read(popupActionsProvider.notifier).state = null;
      await tester.pumpAndSettle();

      expect(find.text(popupMessage), findsOneWidget);

      await tester.tap(find.text(noButton));
      await tester.pumpAndSettle();
      expect(find.text(popupMessage), findsNothing);
    });

    testWidgets('queue is drained when the popup is closed', (tester) async {
      final container = await pumpListener(tester);

      await showPopup(tester, container, decisionMetadata(yesAction: (_) {}));

      await tester.tap(find.text(noButton));
      await tester.pumpAndSettle();

      expect(container.read(popupsProvider), isNull);

      // The same popup can be displayed again.
      await showPopup(tester, container, decisionMetadata(yesAction: (_) {}));
      await tester.tap(find.text(noButton));
      await tester.pumpAndSettle();
    });
  });
}

final class _NoDaemonConnection extends GrpcConnectionController {
  @override
  Future<bool> build() => Completer<bool>().future;
}
