import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:nordvpn/widgets/custom_list_tile.dart';

import '../../test/utils/finders.dart';
import '../../test/utils/test_helpers.dart';

void runObfuscatedServersTests() async {
  group("test obfuscated servers", () {
    testWidgets("VPN card status has obfuscated", (tester) async {
      final app = await tester.setupIntegrationTests();

      final vpnScreen = await app.goToVpnScreen();
      await app.setObfuscatedServers(true);

      await vpnScreen.quickConnect();

      await vpnScreen.waitUntilFound(find.textContaining(t.ui.secured));
      expect(vpnScreen.findStatusLabelText(), contains(t.ui.obfuscated));
    });

    // the daemon reports the obfuscated servers only under NordWhisper, so on any other
    // technology the tile has nothing to offer and stays disabled
    testWidgets("obfuscated servers are offered only for NordWhisper", (
      tester,
    ) async {
      final app = await tester.setupIntegrationTests();

      final vpnScreen = await app.goToVpnScreen();
      await vpnScreen.clickSpecialtyServersTab();

      // matches the tile only while it has servers to offer
      final offeredTile = find.byWidgetPredicate(
        (widget) =>
            widget is CustomListTile &&
            widget.key == ServerListWidgetKeys.obfuscatedVpn &&
            widget.enabled,
      );

      // the tile is listed under every technology, but offers servers only under NordWhisper
      expect(obfuscatedGroupTile(), findsOne);
      expect(offeredTile, findsNothing);

      await app.setTechnology(Technology.NORDWHISPER);
      await vpnScreen.waitUntilFound(offeredTile);

      await app.setTechnology(Technology.NORDLYNX);
      await vpnScreen.waitFor(() => offeredTile.evaluate().isEmpty);
      expect(obfuscatedGroupTile(), findsOne);
    });
  });
}
