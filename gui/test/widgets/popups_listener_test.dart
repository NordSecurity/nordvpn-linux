import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/data/repository/uievent_repository.dart';
import 'package:nordvpn/router/router.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/theme.dart';
import 'package:nordvpn/widgets/popups_listener.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';
import 'package:url_launcher_platform_interface/url_launcher_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';
import '../utils/mock_url_launcher.dart';
import '../utils/test_helpers.dart';

// Records analytics calls instead of sending them to the daemon.
final class _RecordingUiEventRepository implements UiEventRepository {
  int shownCount = 0;
  int learnMoreCount = 0;

  @override
  void reportChangeSettings() {}

  @override
  void reportGetHelp() {}

  @override
  void reportSessionLimitShown() => shownCount++;

  @override
  void reportSessionLimitLearnMore() => learnMoreCount++;
}

// Stays loading forever, so the account controller (listened to by
// PopupsListener) returns null without ever touching gRPC or timers.
final class _FakeGrpcConnectionController extends GrpcConnectionController {
  @override
  FutureOr<bool> build() => Completer<bool>().future;
}

void main() {
  // any daemon status code without dedicated metadata gets a generic
  // info popup — used as the "other popup" occupying the visible slot
  const otherPopupId = 9999;

  late MockUrlLauncher urlLauncher;
  late _RecordingUiEventRepository uiEvents;

  setUpAll(() async {
    SharedPreferencesAsyncPlatform.instance = FakeSharedPreferencesAsync();
    await initServiceLocator();
  });

  setUp(() {
    urlLauncher = MockUrlLauncher();
    UrlLauncherPlatform.instance = urlLauncher;
    uiEvents = _RecordingUiEventRepository();
  });

  Future<ProviderContainer> pumpListener(WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          uiEventRepositoryProvider.overrideWithValue(uiEvents),
          grpcConnectionControllerProvider.overrideWith(
            _FakeGrpcConnectionController.new,
          ),
        ],
        // pin the text scale like setupWidgetTest does; this pump is
        // hand-rolled because setupWidgetTest hosts the child inside the
        // navigator, while PopupsListener must wrap it (as in main.dart) —
        // it calls showDialog during build, which is only legal towards a
        // descendant navigator.
        child: Builder(
          builder: (context) {
            final data = MediaQuery.of(context);
            return MediaQuery(
              data: data.copyWith(textScaler: TextScaler.linear(1)),
              child: MaterialApp(
                theme: lightTheme(),
                builder: (context, child) => PopupsListener(child: child!),
                navigatorKey: goRouterKey,
                home: const Scaffold(body: SizedBox()),
              ),
            );
          },
        ),
      ),
    );
    await tester.pump();
    return ProviderScope.containerOf(
      tester.element(find.byType(PopupsListener)),
    );
  }

  Future<void> dismissViaBarrier(WidgetTester tester) async {
    await tester.tapAt(const Offset(5, 5));
    await tester.pumpAndSettle();
  }

  group('PopupsListener session-limit analytics', () {
    testWidgets('fires one show event per display, including re-shows', (
      tester,
    ) async {
      final container = await pumpListener(tester);

      container
          .read(popupsProvider.notifier)
          .show(DaemonStatusCode.connectionLimitReached);
      await tester.pumpAndSettle();

      expect(
        tester.findPopupWithId(DaemonStatusCode.connectionLimitReached),
        findsOneWidget,
      );
      expect(uiEvents.shownCount, 1);

      await dismissViaBarrier(tester);
      expect(
        tester.findPopupWithId(DaemonStatusCode.connectionLimitReached),
        findsNothing,
      );

      container
          .read(popupsProvider.notifier)
          .show(DaemonStatusCode.connectionLimitReached);
      await tester.pumpAndSettle();

      expect(uiEvents.shownCount, 2);
    });

    testWidgets('a rebuild while the popup is open adds no event or dialog', (
      tester,
    ) async {
      final container = await pumpListener(tester);

      container
          .read(popupsProvider.notifier)
          .show(DaemonStatusCode.connectionLimitReached);
      await tester.pumpAndSettle();
      expect(uiEvents.shownCount, 1);

      tester.element(find.byType(PopupsListener)).markNeedsBuild();
      await tester.pumpAndSettle();

      expect(uiEvents.shownCount, 1);
      expect(
        tester.findPopupWithId(DaemonStatusCode.connectionLimitReached),
        findsOneWidget,
      );
    });

    testWidgets('a duplicate of the visible popup is not counted', (
      tester,
    ) async {
      final container = await pumpListener(tester);

      container
          .read(popupsProvider.notifier)
          .show(DaemonStatusCode.connectionLimitReached);
      await tester.pumpAndSettle();

      // same code again while its popup is visible — dropped, not queued
      container
          .read(popupsProvider.notifier)
          .show(DaemonStatusCode.connectionLimitReached);
      await tester.pumpAndSettle();

      expect(uiEvents.shownCount, 1);
      expect(
        tester.findPopupWithId(DaemonStatusCode.connectionLimitReached),
        findsOneWidget,
      );
    });

    testWidgets('a queued popup is counted once, when actually displayed', (
      tester,
    ) async {
      final container = await pumpListener(tester);

      container.read(popupsProvider.notifier).show(otherPopupId);
      await tester.pumpAndSettle();
      expect(tester.findPopupWithId(otherPopupId), findsOneWidget);

      // arrives while another popup owns the slot — queued, not displayed
      container
          .read(popupsProvider.notifier)
          .show(DaemonStatusCode.connectionLimitReached);
      await tester.pumpAndSettle();
      expect(uiEvents.shownCount, 0);

      // closing the first popup promotes the queued one
      await dismissViaBarrier(tester);

      expect(
        tester.findPopupWithId(DaemonStatusCode.connectionLimitReached),
        findsOneWidget,
      );
      expect(uiEvents.shownCount, 1);
    });

    testWidgets('tapping the help link fires one click event and launches', (
      tester,
    ) async {
      final container = await pumpListener(tester);

      container
          .read(popupsProvider.notifier)
          .show(DaemonStatusCode.connectionLimitReached);
      await tester.pumpAndSettle();

      await tester.tapOnText(find.textRange.ofSubstring('Open help guide'));
      await tester.pumpAndSettle();

      expect(uiEvents.learnMoreCount, 1);
      // the click must not be double-reported as another show
      expect(uiEvents.shownCount, 1);
      expect(
        urlLauncher.launchedUrls.single,
        startsWith('https://support.nordvpn.com/'),
      );
    });
  });
}
