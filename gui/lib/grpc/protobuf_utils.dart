import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart' as pb;
import 'package:nordvpn/pb/daemon/connect.pb.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';

extension ConnectionParamsExt on ConnectionParameters {
  ConnectRequest toConnectRequest() {
    final connectRequest = ConnectRequest();
    // connecting to specific server
    if (serverName.isNotEmpty) {
      connectRequest.serverTag = serverName;
      return connectRequest;
    }

    // connecting to some location
    if (countryCode.isNotEmpty) {
      connectRequest.serverTag = countryCode.toLowerCase();
    }

    if (city.isNotEmpty) {
      connectRequest.serverTag += " ${City(city).sanitizedName}";
    }

    final groupName = group.backendName;
    if (groupName != null) {
      connectRequest.serverGroup = groupName;
    }

    return connectRequest;
  }
}

extension _ProtobufServerGroupExt on pb.ServerGroup {
  String? get backendName {
    switch (this) {
      case pb.ServerGroup.DOUBLE_VPN:
        return doubleVpn;
      case pb.ServerGroup.DEDICATED_IP:
        return dedicatedIp;
      case pb.ServerGroup.ONION_OVER_VPN:
        return onionOverVpn;
      case pb.ServerGroup.P2P:
        return p2p;
      case pb.ServerGroup.OBFUSCATED:
        return obfuscatedServers;
      default:
        return null;
    }
  }
}
