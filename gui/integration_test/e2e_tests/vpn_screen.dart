import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import '../../test/utils/test_helpers.dart';

void runVpnScreenTests() async {
  group("test subscription expired popup", () {
    testWidgets("appears when sub expired, disappears when sub active", (
      tester,
    ) async {
      final app = await tester.setupIntegrationTests();

      final mainScreen = await app.goToVpnScreen();
      await app.expireSubscription();

      await mainScreen.quickConnect();
      expect(mainScreen.isSubscriptionPopupVisible(), isTrue);

      await app.renewSubscription();
      expect(mainScreen.isSubscriptionPopupVisible(), isFalse);
    });
  });

  group("test vpn status card", () {
    testWidgets("has '- Virtual' label for virtual server", (tester) async {
      final app = await tester.setupIntegrationTests();

      // initially, we have the server info and we are not connected
      final mainScreen = await app.goToVpnScreen();
      expect(mainScreen.findServerInfoText(), equals(t.ui.connectToVpn));

      // connect
      app.connect(country: "FR", city: "Paris", isVirtualLocation: true);
      await mainScreen.waitUntilFound(find.textContaining(t.ui.connected));

      // now the server info changed
      expect(mainScreen.findServerInfoText(), contains("Virtual"));
    });

    testWidgets("has server group in status label", (tester) async {
      final app = await tester.setupIntegrationTests();

      final mainScreen = await app.goToVpnScreen();
      expect(mainScreen.findStatusLabelText(), equals(t.ui.notConnected));

      // connect to specialty server
      await mainScreen.clickSpecialtyServersTab();
      await mainScreen.clickDoubleVpnGroup();
      await mainScreen.waitUntilFound(find.textContaining(t.ui.connected));
      expect(
        mainScreen.findStatusLabelText(),
        equals("${t.ui.connected} ${t.ui.to} ${t.ui.doubleVpn}"),
      );

      await mainScreen.clickOnionOverVpn();
      await mainScreen.waitUntilFound(find.textContaining(t.ui.connected));
      expect(
        mainScreen.findStatusLabelText(),
        equals("${t.ui.connected} ${t.ui.to} ${t.ui.onionOverVpn}"),
      );

      await mainScreen.clickP2p();
      await mainScreen.waitUntilFound(find.textContaining(t.ui.connected));
      expect(
        mainScreen.findStatusLabelText(),
        equals("${t.ui.connected} ${t.ui.to} ${t.ui.p2p}"),
      );
    });
  });

  group("test servers list card", () {
    testWidgets("is updated when virtual setting changes", (tester) async {
      final app = await tester.setupIntegrationTests();

      // initially, we have the server info and we are not connected
      final vpnScreen = await app.goToVpnScreen();
      await app.changeVirtualServers(true);
      expect(await vpnScreen.serversListHasVirtualServers(), isTrue);

      await app.changeVirtualServers(false);
      expect(await vpnScreen.serversListHasVirtualServers(), isFalse);
    });
  });
}
