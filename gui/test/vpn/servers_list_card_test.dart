import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/servers_list_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/vpn/server_list_item_factory.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:nordvpn/widgets/custom_list_tile.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';
import '../utils/fakes.dart';
import '../utils/finders.dart';
import '../utils/test_helpers.dart';

// Builds the servers list card, opens the specialty servers tab and returns the
// obfuscated group tile (if listed there).
Future<Finder> _obfuscatedTile(WidgetTester tester) async {
  // NordWhisper is the only technology for which the daemon reports obfuscated
  // servers, and it's the only one where the tile is active
  final serversList = await tester.mockedServersList(
    technology: Technology.NORDWHISPER,
  );

  await tester.setupWidgetTest(
    ServersListCard(onSelected: (_) async {}),
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

  return obfuscatedGroupTile();
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

  testWidgets("obfuscated group is listed under specialty group tab", (
    tester,
  ) async {
    final tile = await _obfuscatedTile(tester);

    expect(tester.widget<CustomListTile>(tile).enabled, isTrue);
    expect(
      _textIn(tester, tile, ServerListItemFactory.specialtyTitleKey),
      t.ui.obfuscated,
    );
    expect(
      _textIn(tester, tile, ServerListItemFactory.specialtyDescriptionKey),
      t.ui.obfuscatedServersDesc,
    );
  });
}
