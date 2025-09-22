import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/service_locator.dart';

import '../../test/utils/finders.dart';
import '../../test/utils/test_helpers.dart';

void runQuickConnectTest(
  String name,
  Technology technology,
  Protocol protocol, {
  bool? obfuscate,
  String? country,
  String? server,
  String? serverCountry,
}) {
  testWidgets("- $name", (tester) async {
    final settings = Settings(technology: technology, protocol: protocol);

    if (obfuscate != null) {
      settings.obfuscate = obfuscate;
    }

    final app = await tester.setupIntegrationTests(appSettings: settings);

    final mainScreen = await app.goToVpnScreen();

    await tester.pumpUntilFound(
      find.text(t.ui.quickConnect),
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
      find.textContaining(t.ui.connected),
      timeout: Duration(seconds: 10),
    );

    await tester.pumpUntilFound(
      find.text(t.ui.disconnect),
      timeout: Duration(seconds: 10),
    );

    if (obfuscate != null) {
      final isConnectedToObfuscated = find.descendant(
        of: vpnStatusCard(),
        matching: find.textContaining(t.ui.obfuscated),
      );
      expect(isConnectedToObfuscated, findsOneWidget);
    } else {
      await tester.pumpUntilFound(
        find.text(t.ui.connected),
        timeout: Duration(seconds: 10),
      );
    }

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
    runQuickConnectTest('nordlynx', Technology.NORDLYNX, Protocol.UDP);

    // Manual TCID: LVPN-6634
    runQuickConnectTest(
      'nordwhisper',
      Technology.NORDWHISPER,
      Protocol.Webtunnel,
    );

    // Manual TCID: LVPN-6273
    runQuickConnectTest('openvpn tcp', Technology.OPENVPN, Protocol.TCP);

    // Manual TCID: LVPN-6276
    runQuickConnectTest(
      'openvpn obfuscation tcp',
      Technology.OPENVPN,
      Protocol.TCP,
      obfuscate: true,
    );

    // Manual TCID: LVPN-6274
    runQuickConnectTest('openvpn udp', Technology.OPENVPN, Protocol.UDP);

    // Manual TCID: LVPN-6275
    runQuickConnectTest(
      'openvpn obfuscation udp',
      Technology.OPENVPN,
      Protocol.UDP,
      obfuscate: true,
    );
  });
  group("Quick connect Smoke Tests", () {
    // Manual TCID: LVPN-6362
    runQuickConnectTest(
      'nordlynx specific country',
      Technology.NORDLYNX,
      Protocol.UDP,
      country: "France",
    );

    // Manual TCID: LVPN-7524
    runQuickConnectTest(
      'nordwhisper specific country',
      Technology.NORDWHISPER,
      Protocol.Webtunnel,
      country: "France",
    );

    // Manual TCID: LVPN-6363
    runQuickConnectTest(
      'openvpn tcp specific country',
      Technology.OPENVPN,
      Protocol.TCP,
      country: "France",
    );

    // Manual TCID: LVPN-6366
    runQuickConnectTest(
      'openvpn obfuscation tcp specific country',
      Technology.OPENVPN,
      Protocol.TCP,
      obfuscate: true,
      country: "Canada",
    );

    // Manual TCID: LVPN-6364
    runQuickConnectTest(
      'openvpn udp specific country',
      Technology.OPENVPN,
      Protocol.UDP,
      country: "France",
    );

    // Manual TCID: LVPN-6365
    runQuickConnectTest(
      'openvpn obfuscation udp specific country',
      Technology.OPENVPN,
      Protocol.UDP,
      obfuscate: true,
      country: "Canada",
    );
  });
  group("Quick connect Smoke Tests", () {
    // Manual TCID: LVPN-7716
    runQuickConnectTest(
      'nordlynx specific server',
      Technology.NORDLYNX,
      Protocol.UDP,
      server: "#12",
      serverCountry: "Germany",
    );

    // Manual TCID: LVPN-7717
    runQuickConnectTest(
      'nordwhisper specific server',
      Technology.NORDWHISPER,
      Protocol.Webtunnel,
      server: "#12",
      serverCountry: "Germany",
    );

    // Manual TCID: LVPN-7719
    runQuickConnectTest(
      'openvpn tcp specific server',
      Technology.OPENVPN,
      Protocol.TCP,
      server: "#12",
      serverCountry: "Germany",
    );

    // Manual TCID: LVPN-7721
    runQuickConnectTest(
      'openvpn obfuscated tcp specific server',
      Technology.OPENVPN,
      Protocol.TCP,
      obfuscate: true,
      server: "#2",
      serverCountry: "Italy",
    );

    // Manual TCID: LVPN-7718
    runQuickConnectTest(
      'openvpn udp specific server',
      Technology.OPENVPN,
      Protocol.UDP,
      server: "#12",
      serverCountry: "Germany",
    );

    // Manual TCID: LVPN-7720
    runQuickConnectTest(
      'openvpn obfuscated udp specific server',
      Technology.OPENVPN,
      Protocol.UDP,
      obfuscate: true,
      server: "#2",
      serverCountry: "Italy",
    );
  });
}
