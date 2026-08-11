import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import '../../test/utils/test_helpers.dart';

void runEnsTests() async {
  group("test ENS", () {
    testWidgets("connection limit reached popup", (tester) async {
      final app = await tester.setupIntegrationTests();

      final mainScreen = await app.goToVpnScreen();
      // set error code returned by the daemon to be
      app.vpnStatus.connectingErrorStatusCode =
          DaemonStatusCode.connectionLimitReached;

      await mainScreen.quickConnect();

      final txtFinder = find.text(t.ui.connectionLimitReachedTitle);
      await tester.pumpUntilFound(txtFinder);
      expect(txtFinder.evaluate().isNotEmpty, isTrue);
      expect(
        find.text(t.ui.connectionLimitReachedTitle).evaluate().isNotEmpty,
        isTrue,
      );
    });
  });
}
