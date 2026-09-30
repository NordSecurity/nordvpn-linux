import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/config/vpn_protocol.pbenum.dart' as pb;

// VPN protocols supported by the application, mirroring the daemon's pb.VPNProtocol.
enum VpnProtocol {
  unknown, // this can be used to handle future case for new protocols
  nordlynx,
  openVpnUdp,
  openVpnTcp,
  nordWhisper,
}

// Convert from the daemon's VPN protocol to VpnProtocol
VpnProtocol vpnProtocolFromPb(pb.VPNProtocol vpnProtocol) {
  switch (vpnProtocol) {
    case pb.VPNProtocol.VPN_PROTOCOL_NORDLYNX:
      return VpnProtocol.nordlynx;
    case pb.VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP:
      return VpnProtocol.openVpnUdp;
    case pb.VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP:
      return VpnProtocol.openVpnTcp;
    case pb.VPNProtocol.VPN_PROTOCOL_NORDWHISPER:
      return VpnProtocol.nordWhisper;
    default:
      return VpnProtocol.unknown;
  }
}

extension VpnProtocolExt on VpnProtocol {
  bool isOpenVpn() {
    return this == VpnProtocol.openVpnTcp || this == VpnProtocol.openVpnUdp;
  }

  // Convert to the daemon's VPN protocol
  pb.VPNProtocol toPb() {
    switch (this) {
      case VpnProtocol.unknown:
        assert(false);
        logger.e("Incorrect protocol value VpnProtocol.unknown");
        return pb.VPNProtocol.VPN_PROTOCOL_UNSPECIFIED;
      case VpnProtocol.nordlynx:
        return pb.VPNProtocol.VPN_PROTOCOL_NORDLYNX;
      case VpnProtocol.openVpnUdp:
        return pb.VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP;
      case VpnProtocol.openVpnTcp:
        return pb.VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP;
      case VpnProtocol.nordWhisper:
        return pb.VPNProtocol.VPN_PROTOCOL_NORDWHISPER;
    }
  }

  String displayName() {
    switch (this) {
      case VpnProtocol.unknown:
        assert(false);
        logger.e("Incorrect protocol value VpnProtocol.unknown");
        return t.ui.protocol;

      case VpnProtocol.nordlynx:
        return t.ui.nordLynx;
      case VpnProtocol.openVpnUdp:
        return t.ui.openVpnUdp;
      case VpnProtocol.openVpnTcp:
        return t.ui.openVpnTcp;

      case VpnProtocol.nordWhisper:
        return t.ui.nordWhisper;
    }
  }
}
