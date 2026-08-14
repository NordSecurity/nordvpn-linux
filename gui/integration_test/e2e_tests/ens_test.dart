import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';

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

      await tester.pumpUntilFound(
        tester.findPopupWithId(DaemonStatusCode.connectionLimitReached),
      );
      expect(mainScreen.isConnectionLimitReachedVisible(), isTrue);
    });
  });
}
