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
import 'package:nordvpn/widgets/popups_listener.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';

void main() {
  const popupId = 987654;
  const popupTitle = "Decision popup";
  const popupMessage = "Do you want to apply this change?";
  const yesButton = "Yes";
  const noButton = "No";
  const firstScreen = "First screen";
  const secondScreen = "Second screen";
  const pageButton = "Page button";

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
  }) {
    return DecisionPopupMetadata(
      id: popupId,
      title: popupTitle,
      message: (_) => popupMessage,
      noButtonText: noButton,
      yesButtonText: yesButton,
      yesAction: yesAction,
      noAction: noAction,
    );
  }

  // Mirrors the production layout from main.dart: the listener is above the
  // navigator, so that the dialogs are displayed on top of it and a popup
  // survives a navigation between the routed pages.
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
          builder: (_, child) =>
              Scaffold(body: PopupsListener(child: child!)),
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
      await tester.pumpAndSettle();

      // The dialog is closed while the action is still running.
      expect(find.text(popupMessage), findsNothing);
      expect(container.read(popupActionsProvider), contains(popupId));

      completer.complete();
      await tester.pumpAndSettle();

      expect(readAfterAwait, isTrue);
      expect(tester.takeException(), isNull);
      expect(container.read(popupActionsProvider), isEmpty);
    });

    // The action keeps running without any indicator on top of the app, so that
    // the user can continue using the screen while the daemon is working.
    testWidgets('the screen is not blocked while the action runs', (
      tester,
    ) async {
      var pageTaps = 0;
      // Never completed, the action stays in progress until the end of the test.
      final completer = Completer<void>();
      final container = await pumpListener(
        tester,
        page: TextButton(
          onPressed: () => pageTaps++,
          child: const Text(pageButton),
        ),
      );

      await showPopup(
        tester,
        container,
        decisionMetadata(yesAction: (_) => completer.future),
      );

      await tester.tap(find.text(yesButton));
      await tester.pumpAndSettle();

      expect(container.read(popupActionsProvider), contains(popupId));
      // Nothing is drawn over the app while the action runs.
      expect(find.byType(ModalBarrier), findsNothing);

      // The page below still receives the input.
      await tester.tap(find.text(pageButton));
      await tester.pumpAndSettle();
      expect(pageTaps, 1);
    });

    // The action outlives the screen which started it, so its error has to be
    // displayed wherever the user is when the daemon answers.
    testWidgets('error is displayed after navigating to another screen', (
      tester,
    ) async {
      final container = await pumpListener(
        tester,
        page: const Text(firstScreen),
      );
      final completer = Completer<void>();

      await showPopup(
        tester,
        container,
        decisionMetadata(
          yesAction: (ref) async {
            await completer.future;
            // This is how the controllers report the daemon status, with their
            // own container scoped ref.
            ref.read(popupsProvider.notifier).show(DaemonStatusCode.configError);
          },
        ),
      );

      await tester.tap(find.text(yesButton));
      await tester.pumpAndSettle();
      expect(find.text(popupMessage), findsNothing);

      // The user leaves the screen which started the action while it still runs.
      unawaited(
        goRouterKey.currentState!.push(
          MaterialPageRoute<void>(builder: (_) => const Text(secondScreen)),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text(secondScreen), findsOneWidget);

      completer.complete();
      await tester.pumpAndSettle();

      expect(find.text(t.ui.settingsWereNotSaved), findsOneWidget);
      expect(find.text(secondScreen), findsOneWidget);

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
      container.read(popupActionsProvider.notifier).state = {1};
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
