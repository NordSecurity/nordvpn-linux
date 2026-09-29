import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';

import '../../test/utils/test_helpers.dart';

void runObfuscatedServersTests() async {
  group("test obfuscated servers", () {
    testWidgets("VPN card status has obfuscated", (tester) async {
      final app = await tester.setupIntegrationTests();

      final vpnScreen = await app.goToVpnScreen();
      await app.setTechnology(Technology.NORDWHISPER);

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
      await vpnScreen.scrollToObfuscatedGroup();

      // the tile is listed under every technology, but offers servers only under NordWhisper
      expect(vpnScreen.isObfuscatedGroupOffered(), isFalse);

      await app.setTechnology(Technology.NORDWHISPER);
      await vpnScreen.waitFor(() => vpnScreen.isObfuscatedGroupOffered());

      await app.setTechnology(Technology.NORDLYNX);
      await vpnScreen.clickSpecialtyServersTab();
      await vpnScreen.scrollToObfuscatedGroup();
      await vpnScreen.waitFor(() => !vpnScreen.isObfuscatedGroupOffered());
    });
  });
}
