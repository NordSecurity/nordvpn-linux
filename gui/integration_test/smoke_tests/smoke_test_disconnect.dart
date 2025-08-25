import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/service_locator.dart';

import '../../test/utils/finders.dart';
import '../../test/utils/test_helpers.dart';

void runDisconnectTest(
  String name,
  Technology technology,
  Protocol protocol, {
  bool? obfuscate,
  String? country,
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

    await mainScreen.disconnect();

    await tester.pumpUntilFound(
      find.text(t.ui.quickConnect),
      timeout: Duration(seconds: 10),
    );

    await tester.pumpUntilFound(
      find.text(t.ui.notConnected),
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
    runDisconnectTest('nordlynx', Technology.NORDLYNX, Protocol.UDP);

    // Manual TCID: LVPN-7527
    runDisconnectTest(
      'nordwhisper',
      Technology.NORDWHISPER,
      Protocol.Webtunnel,
    );

    // Manual TCID: LVPN-6378
    runDisconnectTest('openvpn tcp', Technology.OPENVPN, Protocol.TCP);

    // Manual TCID: LVPN-6380
    runDisconnectTest(
      'openvpn obfuscation tcp',
      Technology.OPENVPN,
      Protocol.TCP,
      obfuscate: true,
    );

    // Manual TCID: LVPN-6379
    runDisconnectTest('openvpn udp', Technology.OPENVPN, Protocol.UDP);

    // Manual TCID: LVPN-6377
    runDisconnectTest(
      'openvpn obfuscation udp',
      Technology.OPENVPN,
      Protocol.UDP,
      obfuscate: true,
    );
  });
  group("Disconnect Smoke Tests", () {
    // Manual TCID: LVPN-6279
    runDisconnectTest(
      'nordlynx specific country',
      Technology.NORDLYNX,
      Protocol.UDP,
      country: "France",
    );

    // Manual TCID: LVPN-6638
    runDisconnectTest(
      'nordwhisper specific country',
      Technology.NORDWHISPER,
      Protocol.Webtunnel,
      country: "France",
    );

    // Manual TCID: LVPN-6360
    runDisconnectTest(
      'openvpn tcp specific country',
      Technology.OPENVPN,
      Protocol.TCP,
      country: "France",
    );

    // Manual TCID: LVPN-6358
    runDisconnectTest(
      'openvpn obfuscation tcp specific country',
      Technology.OPENVPN,
      Protocol.TCP,
      obfuscate: true,
      country: "Canada",
    );

    // Manual TCID: LVPN-6361
    runDisconnectTest(
      'openvpn udp specific country',
      Technology.OPENVPN,
      Protocol.UDP,
      country: "France",
    );

    // Manual TCID: LVPN-6359
    runDisconnectTest(
      'openvpn obfuscation udp specific country',
      Technology.OPENVPN,
      Protocol.UDP,
      obfuscate: true,
      country: "Canada",
    );
  });
}
