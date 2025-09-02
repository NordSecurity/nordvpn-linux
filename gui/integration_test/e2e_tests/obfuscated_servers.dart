import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import '../../test/utils/test_helpers.dart';

void runObfuscatedServersTests() async {
  group("test obfuscated servers search", () {
    testWidgets("obfuscation messages are displayed", (tester) async {
      final app = await tester.setupIntegrationTests();

      final vpnScreen = await app.goToVpnScreen();
      await vpnScreen.clickSearch();
      expect(vpnScreen.isObfuscationWarningDisplayed(), isFalse);
      await vpnScreen.searchServer("invalid server name");
      expect(await vpnScreen.isObfuscationNoResultsFound(), isFalse);

      // switch to obfuscated servers
      await app.setObfuscatedServers(true);
      await vpnScreen.searchServer("");
      expect(vpnScreen.isObfuscationWarningDisplayed(), isTrue);
      await vpnScreen.searchServer("invalid server name");
      expect(await vpnScreen.isObfuscationNoResultsFound(), isTrue);
    });

    testWidgets("VPN card status has obfuscated", (tester) async {
      final app = await tester.setupIntegrationTests();

      final vpnScreen = await app.goToVpnScreen();
      await app.setObfuscatedServers(true);

      await vpnScreen.quickConnect();

      await vpnScreen.waitUntilFound(find.textContaining(t.ui.connected));
      expect(
        vpnScreen.findStatusLabelText(),
        equals("${t.ui.connected} ${t.ui.to} ${t.ui.obfuscated}"),
      );
    });
  });
}
