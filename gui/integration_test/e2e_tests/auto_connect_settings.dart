import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import '../../test/utils/test_helpers.dart';

void runAutoConnectSettingsTests() async {
  group("test auto-connect panel", () {
    testWidgets("has fastest server selected by default", (tester) async {
      final app = await tester.setupIntegrationTests();
      final screen = await app.goToAutoConnectSettingsScreen();
      expect(
        screen.autoConnectServerLabel(),
        equals("${t.ui.fastestServer} (${t.ui.quickConnect})"),
      );
    });

    testWidgets(
      "disables 'Secure my connection' button when connected to selected location",
      (tester) async {
        final app = await tester.setupIntegrationTests();
        final screen = await app.goToAutoConnectSettingsScreen();

        // connect right after opening auto-connect settings - to Fastest server
        expect(screen.isSecureMyConnectionButtonEnabled(), isTrue);
        await screen.secureMyConnection();
        // 'Secure my connection' is now disabled after connecting
        await app.waitUntilConnected();
        expect(screen.isSecureMyConnectionButtonEnabled(), isFalse);

        // pick some location, 'Secure my connection' is enabled again
        await screen.clickListTile(withText: "Spain");
        await app.waitForUiUpdates();
        expect(screen.isSecureMyConnectionButtonEnabled(), isTrue);
        // connect to this selected location and it becomes disabled
        await screen.secureMyConnection();
        await app.waitUntilConnected(country: "ES");
        expect(screen.isSecureMyConnectionButtonEnabled(), isFalse);

        // connect to different location and 'Secure my connection' is enabled again
        app.connect(countryCode: "FR", city: "Paris");
        await app.waitUntilConnected(country: "FR", city: "Paris");

        expect(screen.isSecureMyConnectionButtonEnabled(), isTrue);
      },
    );
  });
}
