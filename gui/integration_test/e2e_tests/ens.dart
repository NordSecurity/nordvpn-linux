import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/pb/daemon/uievent.pb.dart';

import '../../test/utils/test_helpers.dart';

// analytics events the GUI must send to the daemon for the popup
final _sessionLimitShown = UIEvent(
  formReference: UIEvent_FormReference.GUI,
  itemName: UIEvent_ItemName.SESSION_LIMIT,
  itemType: UIEvent_ItemType.SHOW,
);
final _sessionLimitLearnMore = UIEvent(
  formReference: UIEvent_FormReference.SESSION_LIMIT_NOTIFICATION,
  itemName: UIEvent_ItemName.LEARN_MORE,
  itemType: UIEvent_ItemType.CLICK,
);

void runEnsTests() async {
  group("test ENS", () {
    testWidgets("connection limit reached popup", (tester) async {
      final app = await tester.setupIntegrationTests();

      final mainScreen = await app.goToVpnScreen();
      // set error code returned by the daemon when connecting to the VPN
      app.vpnStatus.connectingErrorStatusCode =
          DaemonStatusCode.connectionLimitReached;

      await mainScreen.quickConnect();

      await tester.pumpUntilFound(
        tester.findPopupWithId(DaemonStatusCode.connectionLimitReached),
      );
      expect(mainScreen.isConnectionLimitReachedPopupVisible(), isTrue);

      // displaying the popup reports one show ui event
      await app.waitForUiEvents([_sessionLimitShown]);
    });

    testWidgets("help guide link reports click and launches the URL", (
      tester,
    ) async {
      final app = await tester.setupIntegrationTests();
      final mainScreen = await app.goToVpnScreen();
      app.vpnStatus.connectingErrorStatusCode =
          DaemonStatusCode.connectionLimitReached;

      await mainScreen.quickConnect();
      await tester.pumpUntilFound(
        tester.findPopupWithId(DaemonStatusCode.connectionLimitReached),
      );

      await app.clickConnectionLimitReachedHelpGuideLink();

      await app.waitForUiEvents([_sessionLimitShown, _sessionLimitLearnMore]);
      expect(
        app.launchedUrls.single,
        startsWith("https://support.nordvpn.com/"),
      );
    });
  });
}
