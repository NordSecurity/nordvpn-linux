import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/config/protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';

// Declares the existing protocols supported by the application.
// A protocol from the GUI is equivalent to the pair (technology + protocol)
// from daemon
enum VpnProtocol {
  unknown, // this can be used to handle future case for new protocols
  nordlynx,
  openVpnUdp,
  openVpnTcp,
  nordWhisper,
}

// Convert from Technology and protocol to VpnProtocol
VpnProtocol convertToVpnProtocol(Technology technology, Protocol protocol) {
  switch (technology) {
    case Technology.NORDLYNX:
      return VpnProtocol.nordlynx;
    case Technology.OPENVPN:
      return (protocol == Protocol.TCP)
          ? VpnProtocol.openVpnTcp
          : VpnProtocol.openVpnUdp;
    case Technology.UNKNOWN_TECHNOLOGY:
      return VpnProtocol.unknown;
    case Technology.NORDWHISPER:
      return VpnProtocol.nordWhisper;
    default:
      assert(false);
      return VpnProtocol.unknown;
  }
}

(Technology technology, Protocol protocol) toTechnologyAndProtocol(
  VpnProtocol vpnProtocol,
) {
  Technology technology;
  Protocol protocol;
  switch (vpnProtocol) {
    case VpnProtocol.unknown:
      assert(false);
      logger.e("Incorrect protocol value VpnProtocol.unknown");
      technology = Technology.NORDLYNX;
      protocol = Protocol.UDP;

    case VpnProtocol.nordlynx:
      technology = Technology.NORDLYNX;
      protocol = Protocol.UDP;
      break;
    case VpnProtocol.openVpnUdp:
      technology = Technology.OPENVPN;
      protocol = Protocol.UDP;
      break;
    case VpnProtocol.openVpnTcp:
      technology = Technology.OPENVPN;
      protocol = Protocol.TCP;
      break;
    case VpnProtocol.nordWhisper:
      technology = Technology.NORDWHISPER;
      protocol = Protocol.Webtunnel;
      break;
  }

  return (technology, protocol);
}

extension VpnProtocolExt on VpnProtocol {
  bool isOpenVpn() {
    return this == VpnProtocol.openVpnTcp || this == VpnProtocol.openVpnUdp;
  }
}
