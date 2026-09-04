import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/pb/daemon/config/protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';

void main() {
  // The daemon reports a (technology, protocol) pair, the GUI works with a
  // single protocol value.
  final cases = [
    (
      name: "NordLynx ignores the protocol field",
      technology: Technology.NORDLYNX,
      protocol: Protocol.UDP,
      expected: VpnProtocol.nordlynx,
    ),
    (
      name: "OpenVPN over UDP",
      technology: Technology.OPENVPN,
      protocol: Protocol.UDP,
      expected: VpnProtocol.openVpnUdp,
    ),
    (
      name: "OpenVPN over TCP",
      technology: Technology.OPENVPN,
      protocol: Protocol.TCP,
      expected: VpnProtocol.openVpnTcp,
    ),
    (
      name: "NordWhisper is reported with the Webtunnel protocol",
      technology: Technology.NORDWHISPER,
      protocol: Protocol.Webtunnel,
      expected: VpnProtocol.nordWhisper,
    ),
    (
      name: "unknown technology maps to unknown",
      technology: Technology.UNKNOWN_TECHNOLOGY,
      protocol: Protocol.UNKNOWN_PROTOCOL,
      expected: VpnProtocol.unknown,
    ),
  ];

  for (final testCase in cases) {
    test(testCase.name, () {
      expect(
        convertToVpnProtocol(testCase.technology, testCase.protocol),
        testCase.expected,
      );
    });
  }
}
