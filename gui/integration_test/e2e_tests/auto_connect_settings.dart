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
      "disables 'Connect now' button when connected to selected location",
      (tester) async {
        final app = await tester.setupIntegrationTests();
        final screen = await app.goToAutoConnectSettingsScreen();

        // connect right after opening auto-connect settings - to Fastest server
        expect(screen.isConnectNowButtonEnabled(), isTrue);
        await screen.connectNow();
        // 'Connect now' is now disabled after connecting
        await screen.waitFor(() => !screen.isConnectNowButtonEnabled());
        expect(screen.isConnectNowButtonEnabled(), isFalse);

        // pick some location, connect now is enabled again
        await screen.clickListTile(withText: "Spain");
        expect(screen.isConnectNowButtonEnabled(), isTrue);
        // connect to this selected location and it becomes disabled
        await screen.connectNow();
        expect(screen.isConnectNowButtonEnabled(), isFalse);

        // connect to different location and 'Connect now' is enabled again
        app.connect(country: "FR", city: "Paris");
        await screen.waitFor(screen.isConnectNowButtonEnabled);

        expect(screen.isConnectNowButtonEnabled(), isTrue);
      },
    );
  });
}
