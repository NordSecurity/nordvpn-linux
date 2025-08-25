import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart';

extension Convert on ServerGroup {
  bool get isStandardVpn => this == ServerGroup.STANDARD_VPN_SERVERS;

  ServerType? toSpecialtyType() {
    return switch (this) {
      ServerGroup.DOUBLE_VPN => ServerType.doubleVpn,
      ServerGroup.ONION_OVER_VPN => ServerType.onionOverVpn,
      ServerGroup.DEDICATED_IP => ServerType.dedicatedIP,
      ServerGroup.P2P => ServerType.p2p,
      ServerGroup.STANDARD_VPN_SERVERS => ServerType.standardVpn,
      ServerGroup.OBFUSCATED => ServerType.obfuscated,
      _ => null,
    };
  }
}
