import 'package:flutter_test/flutter_test.dart';

import '../../test/utils/test_helpers.dart';

void runCustomDnsTests() async {
  group("custom DNS smoke tests", () {
    testWidgets("can add DNS servers when TP is off", (tester) async {
      final app = await tester.setupIntegrationTests();

      final dnsScreen = await app.goToCustomDnsSettingsScreen();

      expect(dnsScreen.isDnsEnabled(), isFalse);
      expect(dnsScreen.isAddDnsFormEnabled(), isFalse);
      expect(dnsScreen.isAddButtonEnabled(), isFalse);
      expect(dnsScreen.isAddButtonVisible(), isTrue);
      expect(dnsScreen.isInputFieldEmpty(), isTrue);
      expect(dnsScreen.serversList(), findsNothing);

      // enable DNS toggle
      await dnsScreen.tapOnOffSwitch();
      await app.waitForUiUpdates();
      expect(dnsScreen.isDisableTpPopupDisplayed(), isFalse);
      expect(dnsScreen.isDnsEnabled(), isTrue);
      expect(dnsScreen.isAddDnsFormEnabled(), isTrue);
      expect(dnsScreen.isAddButtonEnabled(), isFalse);
      expect(dnsScreen.isAddButtonVisible(), isTrue);
      expect(dnsScreen.serversList(), findsNothing);

      await dnsScreen.enterDnsAddress("1.2");
      await app.waitForUiUpdates();
      expect(dnsScreen.isAddButtonEnabled(), isFalse);

      // enter a valid IP address
      final dnsServerAddress = "1.2.3.4";
      await dnsScreen.enterDnsAddress(dnsServerAddress);
      await app.waitForUiUpdates();
      expect(dnsScreen.isAddButtonEnabled(), isTrue);
      expect(dnsScreen.isAddButtonVisible(), isTrue);

      // tap to add the server
      await tester.tap(dnsScreen.addButton());
      await app.waitForUiUpdates(timeout: Duration(milliseconds: 500));
      expect(dnsScreen.isAddButtonVisible(), isFalse);
      await app.waitForUiUpdates();

      // the vales is added into the list and the add form fields are reset
      expect(dnsScreen.isDnsEnabled(), isTrue);
      expect(dnsScreen.isAddDnsFormEnabled(), isTrue);
      expect(dnsScreen.isAddButtonEnabled(), isFalse);
      expect(dnsScreen.isAddButtonVisible(), isTrue);
      expect(dnsScreen.isInputFieldEmpty(), isTrue);
      await app.refreshAppState();
      expect(dnsScreen.serversList(), findsOne);
      expect(dnsScreen.serversItem(dnsServerAddress), findsOne);

      // delete added server
      await dnsScreen.deleteServer(dnsServerAddress);
      await app.refreshAppState();
      // await tester.pumpAndSettleWithTimeout(duration: Duration(seconds: 1));
      expect(dnsScreen.serversList(), findsNothing);

      // disable custom DNS
      await dnsScreen.tapOnOffSwitch();
      await app.waitForUiUpdates();
      expect(dnsScreen.isDnsEnabled(), isFalse);
      expect(dnsScreen.isAddDnsFormEnabled(), isFalse);
      expect(dnsScreen.isAddButtonEnabled(), isFalse);
      expect(dnsScreen.isAddButtonVisible(), isTrue);
      expect(dnsScreen.isInputFieldEmpty(), isTrue);
      expect(dnsScreen.serversList(), findsNothing);
    });
  });

  testWidgets("enable custom DNS when TP is on", (tester) async {
    final app = await tester.setupIntegrationTests();
    app.setThreatProtection(true);

    final dnsScreen = await app.goToCustomDnsSettingsScreen();

    expect(dnsScreen.isDnsEnabled(), isFalse);
    // enable DNS toggle
    await dnsScreen.tapOnOffSwitch();
    await app.waitForUiUpdates();
    expect(dnsScreen.isDisableTpPopupDisplayed(), isTrue);
  });
}
