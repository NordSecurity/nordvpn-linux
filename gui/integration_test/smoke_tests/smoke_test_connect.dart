import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/vpn_protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/service_locator.dart';

import '../../test/utils/finders.dart';
import '../../test/utils/test_helpers.dart';

void runQuickConnectTest(
  String name,
  VPNProtocol vpnProtocol, {
  String? country,
  String? server,
  String? serverCountry,
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
    } else if (server != null) {
      await mainScreen.clickSearch();
      await mainScreen.searchServer(server);
      await mainScreen.connectToCountry(serverCountry!);
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
  });
}

void main() {
  WidgetController.hitTestWarningShouldBeFatal = true;

  setUp(() async => await initServiceLocator());
  tearDown(() async => await sl.reset(dispose: true));

  // Call your existing test function
  runConnectSmokeTests();
}

void runConnectSmokeTests() {
  group("Quick connect Smoke Tests", () {
    // Manual TCID: LVPN-6271
    runQuickConnectTest('nordlynx', VPNProtocol.VPN_PROTOCOL_NORDLYNX);

    // Manual TCID: LVPN-6634
    runQuickConnectTest('nordwhisper', VPNProtocol.VPN_PROTOCOL_NORDWHISPER);

    // Manual TCID: LVPN-6273
    runQuickConnectTest('openvpn tcp', VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP);

    // Manual TCID: LVPN-6274
    runQuickConnectTest('openvpn udp', VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP);
  });
  group("Quick connect Smoke Tests", () {
    // Manual TCID: LVPN-6362
    runQuickConnectTest(
      'nordlynx specific country',
      VPNProtocol.VPN_PROTOCOL_NORDLYNX,
      country: "France",
    );

    // Manual TCID: LVPN-7524
    runQuickConnectTest(
      'nordwhisper specific country',
      VPNProtocol.VPN_PROTOCOL_NORDWHISPER,
      country: "France",
    );

    // Manual TCID: LVPN-6363
    runQuickConnectTest(
      'openvpn tcp specific country',
      VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP,
      country: "France",
    );

    // Manual TCID: LVPN-6364
    runQuickConnectTest(
      'openvpn udp specific country',
      VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP,
      country: "France",
    );
  });
  group("Quick connect Smoke Tests", () {
    // Manual TCID: LVPN-7716
    runQuickConnectTest(
      'nordlynx specific server',
      VPNProtocol.VPN_PROTOCOL_NORDLYNX,
      server: "#12",
      serverCountry: "Germany",
    );

    // Manual TCID: LVPN-7717
    runQuickConnectTest(
      'nordwhisper specific server',
      VPNProtocol.VPN_PROTOCOL_NORDWHISPER,
      server: "#12",
      serverCountry: "Germany",
    );

    // Manual TCID: LVPN-7719
    runQuickConnectTest(
      'openvpn tcp specific server',
      VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP,
      server: "#12",
      serverCountry: "Germany",
    );

    // Manual TCID: LVPN-7718
    runQuickConnectTest(
      'openvpn udp specific server',
      VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP,
      server: "#12",
      serverCountry: "Germany",
    );
  });
}
