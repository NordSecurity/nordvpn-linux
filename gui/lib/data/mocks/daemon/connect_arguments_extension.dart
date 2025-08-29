import 'package:nordvpn/pb/daemon/config/group.pbenum.dart' as config;
import 'package:nordvpn/pb/daemon/connect.pb.dart';

extension Conversions on ConnectRequest {
  config.ServerGroup toServerGroup() {
    switch (serverGroup) {
      case "Double_vpn":
        return config.ServerGroup.DOUBLE_VPN;
      case "Dedicated_IP":
        return config.ServerGroup.DEDICATED_IP;
      case "Onion_Over_VPN":
        return config.ServerGroup.ONION_OVER_VPN;
      case "p2p":
        return config.ServerGroup.P2P;
      case "Obfuscated_Servers":
        return config.ServerGroup.OBFUSCATED;
      default:
        return config.ServerGroup.STANDARD_VPN_SERVERS;
    }
  }
}
