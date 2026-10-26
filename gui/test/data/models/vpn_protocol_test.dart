import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/pb/daemon/config/vpn_protocol.pbenum.dart' as pb;

void main() {
  const pairs = {
    pb.VPNProtocol.VPN_PROTOCOL_NORDLYNX: VpnProtocol.nordlynx,
    pb.VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP: VpnProtocol.openVpnUdp,
    pb.VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP: VpnProtocol.openVpnTcp,
    pb.VPNProtocol.VPN_PROTOCOL_NORDWHISPER: VpnProtocol.nordWhisper,
  };

  test('vpnProtocolFromPb maps every daemon protocol', () {
    pairs.forEach((daemon, gui) {
      expect(vpnProtocolFromPb(daemon), gui, reason: '$daemon');
    });
  });

  test('vpnProtocolFromPb maps UNSPECIFIED to unknown', () {
    expect(
      vpnProtocolFromPb(pb.VPNProtocol.VPN_PROTOCOL_UNSPECIFIED),
      VpnProtocol.unknown,
    );
  });

  test('toPb is the inverse of vpnProtocolFromPb', () {
    pairs.forEach((daemon, gui) {
      expect(gui.toPb(), daemon, reason: '$gui');
    });
  });

  test('toPb asserts on unknown', () {
    expect(() => VpnProtocol.unknown.toPb(), throwsA(isA<AssertionError>()));
  });

  test('every daemon protocol has a GUI mapping', () {
    for (final p in pb.VPNProtocol.values) {
      if (p == pb.VPNProtocol.VPN_PROTOCOL_UNSPECIFIED) continue;
      expect(vpnProtocolFromPb(p), isNot(VpnProtocol.unknown), reason: '$p');
    }
  });
}
