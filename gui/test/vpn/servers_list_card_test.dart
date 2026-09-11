import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/mocks/daemon/mock_servers_list.dart';
import 'package:nordvpn/data/models/servers_list.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/servers_list_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/theme.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:nordvpn/widgets/custom_list_tile.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';
import '../utils/fakes.dart';
import '../utils/finders.dart';
import '../utils/test_helpers.dart';

Future<ServersList> _serversList({required Technology technology}) async {
  final appState = StreamController<AppState>();
  addTearDown(appState.close);
  final mockServersList = MockServersList(appState);
  addTearDown(mockServersList.dispose);

  appState.add(
    AppState(
      settingsChange: Settings(technology: technology, virtualLocation: true),
    ),
  );
  await pumpEventQueue();

  final container = ProviderContainer(
    overrides: [
      serversListControllerProvider.overrideWithBuild(
        (ref, notifier) => ServersList.empty(),
      ),
    ],
  );
  addTearDown(container.dispose);

  container
      .read(serversListControllerProvider.notifier)
      .onServersListChanged(mockServersList.serversList);

  return container.read(serversListControllerProvider).value!;
}

void main() {
  setUpAll(() async {
    SharedPreferencesAsyncPlatform.instance = FakeSharedPreferencesAsync();
    await initServiceLocator();
  });

  group("obfuscated servers tile", () {
    Future<Finder> openSpecialtyServers(
      WidgetTester tester, {
      required Technology technology,
    }) async {
      final serversList = (await tester.runAsync(
        () => _serversList(technology: technology),
      ))!;

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            serversListControllerProvider.overrideWithBuild(
              (ref, notifier) => serversList,
            ),
            vpnStatusControllerProvider.overrideWithBuild(
              (ref, notifier) => fakeVpnStatus(),
            ),
            accountControllerProvider.overrideWithBuild(
              (ref, notifier) => null,
            ),
          ],
          child: MaterialApp(
            theme: lightTheme(),
            home: Scaffold(body: ServersListCard(onSelected: (_) async {})),
          ),
        ),
      );
      await tester.pumpAndSettleWithTimeout();

      await tester.tap(specialtyServersTab());
      await tester.pumpAndSettle();

      return obfuscatedGroupTile();
    }

    testWidgets("is offered under NordWhisper", (tester) async {
      final tile = await openSpecialtyServers(
        tester,
        technology: Technology.NORDWHISPER,
      );

      expect(tester.widget<CustomListTile>(tile).enabled, isTrue);
      expect(
        find.descendant(of: tile, matching: find.text(t.ui.obfuscated)),
        findsOne,
      );
      expect(
        find.descendant(
          of: tile,
          matching: find.text(t.ui.obfuscatedServersDesc),
        ),
        findsOne,
      );
      expect(
        find.descendant(of: tile, matching: find.byType(IconButton)),
        findsOne,
      );
    });

    testWidgets("is not selectable under any other technology", (tester) async {
      final tile = await openSpecialtyServers(
        tester,
        technology: Technology.NORDLYNX,
      );

      expect(tester.widget<CustomListTile>(tile).enabled, isFalse);
      expect(
        find.descendant(of: tile, matching: find.byType(IconButton)),
        findsNothing,
      );
    });
  });
}
