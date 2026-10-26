import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/vpn_protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/service_locator.dart';

import '../../test/utils/finders.dart';
import '../../test/utils/test_helpers.dart';

void runDisconnectTest(
  String name,
  VPNProtocol vpnProtocol, {
  String? country,
}) {
  testWidgets("- $name", (tester) async {
    final settings = Settings(vpnProtocol: vpnProtocol);

    final app = await tester.setupIntegrationTests(appSettings: settings);

    final mainScreen = await app.goToVpnScreen();

    await tester.pumpUntilFound(
      find.text(t.ui.secureMyConnection),
      timeout: Duration(seconds: 10),
    );

    if (country != null) {
      await mainScreen.connectToCountry(country);
    } else {
      await mainScreen.quickConnect();

      await tester.pumpUntilFound(
        find.text(t.ui.cancel),
        timeout: Duration(seconds: 10),
      );
    }

    await tester.pumpUntilFound(
      find.textContaining(t.ui.secured),
      timeout: Duration(seconds: 10),
    );

    await tester.pumpUntilFound(
      find.text(t.ui.pauseConnection),
      timeout: Duration(seconds: 10),
    );

    await tester.pumpUntilFound(
      find.text(t.ui.secured),
      timeout: Duration(seconds: 10),
    );

    if (country != null) {
      final isCountryConnected = find.descendant(
        of: vpnStatusCard(),
        matching: find.textContaining(country),
      );
      expect(isCountryConnected, findsOneWidget);
    }

    await mainScreen.disconnect();

    await tester.pumpUntilFound(
      find.text(t.ui.secureMyConnection),
      timeout: Duration(seconds: 10),
    );

    await tester.pumpUntilFound(
      find.text(t.ui.notSecured),
      timeout: Duration(seconds: 10),
    );
  });
}

void main() {
  WidgetController.hitTestWarningShouldBeFatal = true;

  setUp(() async => await initServiceLocator());
  tearDown(() async => await sl.reset(dispose: true));

  // Call your existing test function
  runDisconnectSmokeTests();
}

void runDisconnectSmokeTests() {
  group("Disconnect Smoke Tests", () {
    // Manual TCID: LVPN-6375
    runDisconnectTest('nordlynx', VPNProtocol.VPN_PROTOCOL_NORDLYNX);

    // Manual TCID: LVPN-7527
    runDisconnectTest('nordwhisper', VPNProtocol.VPN_PROTOCOL_NORDWHISPER);

    // Manual TCID: LVPN-6378
    runDisconnectTest('openvpn tcp', VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP);

    // Manual TCID: LVPN-6379
    runDisconnectTest('openvpn udp', VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP);
  });
  group("Disconnect Smoke Tests", () {
    // Manual TCID: LVPN-6279
    runDisconnectTest(
      'nordlynx specific country',
      VPNProtocol.VPN_PROTOCOL_NORDLYNX,
      country: "France",
    );

    // Manual TCID: LVPN-6638
    runDisconnectTest(
      'nordwhisper specific country',
      VPNProtocol.VPN_PROTOCOL_NORDWHISPER,
      country: "France",
    );

    // Manual TCID: LVPN-6360
    runDisconnectTest(
      'openvpn tcp specific country',
      VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP,
      country: "France",
    );

    // Manual TCID: LVPN-6361
    runDisconnectTest(
      'openvpn udp specific country',
      VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP,
      country: "France",
    );
  });
}
