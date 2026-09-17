import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/servers_list_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/vpn/server_list_item_factory.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';
import '../utils/fakes.dart';
import '../utils/finders.dart';
import '../utils/test_helpers.dart';

final _connectRequests = <ConnectArguments>[];

Future<void> _recordConnectRequest(ConnectArguments args) async {
  _connectRequests.add(args);
}

// Builds the servers list card, opens the specialty servers tab and returns the
// obfuscated group tile (if listed there).
Future<Finder> _obfuscatedTile(
  WidgetTester tester, {
  required Technology technology,
}) async {
  _connectRequests.clear();
  final serversList = await tester.mockedServersList(technology: technology);

  await tester.setupWidgetTest(
    ServersListCard(onSelected: _recordConnectRequest),
    overrides: [
      serversListControllerProvider.overrideWithBuild(
        (ref, notifier) => serversList,
      ),
      vpnStatusControllerProvider.overrideWithBuild(
        (ref, notifier) => fakeVpnStatus(),
      ),
      accountControllerProvider.overrideWithBuild((ref, notifier) => null),
    ],
  );

  await tester.tap(specialtyServersTab());
  await tester.pumpAndSettle();

  final tile = obfuscatedGroupTile();
  await tester.ensureVisible(tile);
  await tester.pumpAndSettle();
  return tile;
}

String? _textIn(WidgetTester tester, Finder tile, Key key) {
  final text = find.descendant(of: tile, matching: find.byKey(key));
  expect(text, findsOne);
  return tester.widget<Text>(text).data;
}

void main() {
  setUpAll(() async {
    SharedPreferencesAsyncPlatform.instance = FakeSharedPreferencesAsync();
    await initServiceLocator();
  });

  testWidgets("obfuscated group is labelled as a specialty group", (
    tester,
  ) async {
    final tile = await _obfuscatedTile(
      tester,
      technology: Technology.NORDWHISPER,
    );

    expect(
      _textIn(tester, tile, ServerListItemFactory.specialtyTitleKey),
      t.ui.obfuscated,
    );
    expect(
      _textIn(tester, tile, ServerListItemFactory.specialtyDescriptionKey),
      t.ui.obfuscatedServersDesc,
    );
  });

  testWidgets("obfuscated group connects when the daemon reports its servers", (
    tester,
  ) async {
    final tile = await _obfuscatedTile(
      tester,
      technology: Technology.NORDWHISPER,
    );

    await tester.tap(tile);
    await tester.pumpAndSettle();

    expect(_connectRequests.single.specialtyGroup, ServerType.obfuscated);
  });

  testWidgets("obfuscated group is listed but inactive without its servers", (
    tester,
  ) async {
    final tile = await _obfuscatedTile(tester, technology: Technology.NORDLYNX);

    // the group stays on the list so that the user knows it exists
    expect(
      _textIn(tester, tile, ServerListItemFactory.specialtyTitleKey),
      t.ui.obfuscated,
    );

    await tester.tap(tile);
    await tester.pumpAndSettle();

    expect(_connectRequests, isEmpty);
  });
}
