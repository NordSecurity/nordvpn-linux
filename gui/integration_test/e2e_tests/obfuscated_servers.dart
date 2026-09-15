import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';

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
  });
}
