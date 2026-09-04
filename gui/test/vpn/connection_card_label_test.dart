import 'package:flutter/material.dart' hide ConnectionState;
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/vpn/connection_card_label.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';
import '../utils/test_helpers.dart';

void main() {
  setUpAll(() async {
    SharedPreferencesAsyncPlatform.instance = FakeSharedPreferencesAsync();
    await initServiceLocator();
  });

  VpnStatus vpnStatus({
    required VpnProtocol protocol,
    ConnectionState state = ConnectionState.CONNECTED,
    ServerGroup group = ServerGroup.UNDEFINED,
    bool isObfuscated = false,
  }) {
    return VpnStatus(
      ip: "127.0.0.1",
      hostname: "lt123.nordvpn.com",
      city: null,
      country: null,
      status: state,
      protocol: protocol,
      isVirtualLocation: false,
      isObfuscated: isObfuscated,
      connectionParameters: ConnectionParameters(
        source: ConnectionSource.MANUAL,
        group: group,
      ),
      isMeshnetRouting: false,
    );
  }

  List<String> labelTexts(WidgetTester tester) {
    final row = tester.widget<Row>(find.byKey(ConnectionCardLabel.labelKey));
    return row.children
        .whereType<Text>()
        .map((text) => text.data ?? "")
        .toList();
  }

  final cases = [
    (
      name: "NordWhisper is labelled as obfuscated",
      protocol: VpnProtocol.nordWhisper,
      group: ServerGroup.UNDEFINED,
      isObfuscated: false,
      serverType: t.ui.obfuscated,
    ),
    (
      name: "a plain NordLynx connection has no server type",
      protocol: VpnProtocol.nordlynx,
      group: ServerGroup.UNDEFINED,
      isObfuscated: false,
      serverType: null,
    ),
    (
      name: "a specialty group is labelled with its own name",
      protocol: VpnProtocol.nordlynx,
      group: ServerGroup.DOUBLE_VPN,
      isObfuscated: false,
      serverType: t.ui.doubleVpn,
    ),
    ( // todo: change this later when OVPN drops obfuscation
      name: "the obfuscated group is still labelled as obfuscated",
      protocol: VpnProtocol.openVpnTcp,
      group: ServerGroup.OBFUSCATED,
      isObfuscated: true,
      serverType: t.ui.obfuscated,
    ),
    ( // todo: change this later when OVPN drops obfuscation
      name: "the daemon obfuscated flag alone does not add a label",
      protocol: VpnProtocol.openVpnTcp,
      group: ServerGroup.UNDEFINED,
      isObfuscated: true,
      serverType: null,
    ),
  ];

  for (final testCase in cases) {
    testWidgets(testCase.name, (tester) async {
      await tester.setupWidgetTest(
        ConnectionCardLabel(
          vpnStatus: vpnStatus(
            protocol: testCase.protocol,
            group: testCase.group,
            isObfuscated: testCase.isObfuscated,
          ),
        ),
      );

      expect(labelTexts(tester), [
        t.ui.secured,
        if (testCase.serverType != null) testCase.serverType!,
      ]);
    });
  }

  testWidgets("no server type while disconnected", (tester) async {
    await tester.setupWidgetTest(
      ConnectionCardLabel(
        vpnStatus: vpnStatus(
          protocol: VpnProtocol.nordWhisper,
          state: ConnectionState.DISCONNECTED,
        ),
      ),
    );

    expect(labelTexts(tester), [t.ui.notSecured]);
  });
}
