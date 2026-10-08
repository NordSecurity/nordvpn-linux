import 'package:nordvpn/constants.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart' as config;
import 'package:nordvpn/pb/daemon/connect.pb.dart';

extension Conversions on ConnectRequest {
  config.ServerGroup toServerGroup() {
    switch (serverGroup) {
      case doubleVpn:
        return config.ServerGroup.DOUBLE_VPN;
      case dedicatedIp:
        return config.ServerGroup.DEDICATED_IP;
      case onionOverVpn:
        return config.ServerGroup.ONION_OVER_VPN;
      case obfuscatedServers:
        return config.ServerGroup.NW_OBFUSCATED;
      default:
        return config.ServerGroup.STANDARD_VPN_SERVERS;
    }
  }
}
