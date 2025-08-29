import 'package:flutter_test/flutter_test.dart';

import '../../test/utils/test_helpers.dart';

void runConnectSettingsTests() async {
  group("test connect screen", () {
    testWidgets("can be enabled and disabled", (tester) async {
      final app = await tester.setupIntegrationTests();

      // initially, auto-connect is off
      final connectionSettings = await app.goToConnectionSettingsScreen();
      expect(connectionSettings.isAutoConnectSwitchOn(), isFalse);
      expect(connectionSettings.isAutoConnectTileEnabled(), isFalse);

      // enable
      await connectionSettings.clickAutoConnectSwitch();

      expect(connectionSettings.isAutoConnectSwitchOn(), isTrue);
      expect(connectionSettings.isAutoConnectTileEnabled(), isTrue);

      // disable again
      await connectionSettings.clickAutoConnectSwitch();

      expect(connectionSettings.isAutoConnectSwitchOn(), isFalse);
      expect(connectionSettings.isAutoConnectTileEnabled(), isFalse);
    });
  });
}
