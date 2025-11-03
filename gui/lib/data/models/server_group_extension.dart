import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart';

extension Convert on ServerGroup {
  bool get isStandardVpn => this == ServerGroup.STANDARD_VPN_SERVERS;

  ServerType? toSpecialtyType() {
    return switch (this) {
      // groups
      ServerGroup.DOUBLE_VPN => ServerType.doubleVpn,
      ServerGroup.ONION_OVER_VPN => ServerType.onionOverVpn,
      ServerGroup.DEDICATED_IP => ServerType.dedicatedIP,
      ServerGroup.P2P => ServerType.p2p,
      ServerGroup.STANDARD_VPN_SERVERS => ServerType.standardVpn,
      ServerGroup.OBFUSCATED => ServerType.obfuscated,
      // [Deprecated] Region
      ServerGroup.EUROPE => ServerType.europe,
      // [Deprecated] Region
      ServerGroup.THE_AMERICAS => ServerType.theAmericas,
      // [Deprecated] Region
      ServerGroup.ASIA_PACIFIC => ServerType.asiaPacific,
      // [Deprecated] Region
      ServerGroup.AFRICA_THE_MIDDLE_EAST_AND_INDIA =>
        ServerType.africaTheMiddleEastAndIndia,
      _ => null,
    };
  }
}
