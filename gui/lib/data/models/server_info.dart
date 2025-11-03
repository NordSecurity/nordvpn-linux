// specialty groups
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart' as config;

// Server representation for the GUI
final class ServerInfo {
  final int id;
  final String hostname;
  final bool isVirtual;

  ServerInfo({
    required this.id,
    required this.hostname,
    required this.isVirtual,
  });

  String get serverNumber => RegExp(r'\d+').firstMatch(hostname)?[0] ?? "";

  String serverName() {
    final endIndex = hostname.indexOf(".");
    if (endIndex == -1) {
      logger.e("failed to parse hostname $hostname");
      return "";
    }

    return hostname.substring(0, endIndex);
  }
}

enum ServerType {
  dedicatedIP,
  doubleVpn,
  onionOverVpn,
  p2p,
  standardVpn,
  obfuscated,
  // [Deprecated] Region
  europe,
  // [Deprecated] Region
  theAmericas,
  // [Deprecated] Region
  asiaPacific,
  // [Deprecated] Region
  africaTheMiddleEastAndIndia,
}

extension Daemon on ServerType {
  String? get backendName {
    switch (this) {
      case ServerType.doubleVpn:
        return doubleVpn;
      case ServerType.dedicatedIP:
        return dedicatedIp;
      case ServerType.onionOverVpn:
        return onionOverVpn;
      case ServerType.p2p:
        return p2p;
      case ServerType.obfuscated:
        return obfuscatedServers;
      case ServerType.standardVpn:
        return null;
      // [Deprecated] Regions
      case ServerType.europe:
        return europe;
      // [Deprecated] Region
      case ServerType.theAmericas:
        return theAmericas;
      // [Deprecated] Region
      case ServerType.asiaPacific:
        return asiaPacific;
      // [Deprecated] Region
      case ServerType.africaTheMiddleEastAndIndia:
        return africaTheMiddleEastAndIndia;
    }
  }

  config.ServerGroup toServerGroup() {
    switch (this) {
      case ServerType.doubleVpn:
        return config.ServerGroup.DOUBLE_VPN;
      case ServerType.dedicatedIP:
        return config.ServerGroup.DEDICATED_IP;
      case ServerType.onionOverVpn:
        return config.ServerGroup.ONION_OVER_VPN;
      case ServerType.p2p:
        return config.ServerGroup.P2P;
      case ServerType.standardVpn:
        return config.ServerGroup.STANDARD_VPN_SERVERS;
      case ServerType.obfuscated:
        return config.ServerGroup.OBFUSCATED;
      // [Deprecated] Region
      case ServerType.europe:
        return config.ServerGroup.EUROPE;
      // [Deprecated] Region
      case ServerType.theAmericas:
        return config.ServerGroup.THE_AMERICAS;
      // [Deprecated] Region
      case ServerType.asiaPacific:
        return config.ServerGroup.ASIA_PACIFIC;
      // [Deprecated] Region
      case ServerType.africaTheMiddleEastAndIndia:
        return config.ServerGroup.AFRICA_THE_MIDDLE_EAST_AND_INDIA;
    }
  }
}

const Map<config.ServerGroup, ServerType> _groupTitles = {
  config.ServerGroup.DOUBLE_VPN: ServerType.doubleVpn,
  config.ServerGroup.ONION_OVER_VPN: ServerType.onionOverVpn,
  config.ServerGroup.STANDARD_VPN_SERVERS: ServerType.standardVpn,
  config.ServerGroup.P2P: ServerType.p2p,
  config.ServerGroup.OBFUSCATED: ServerType.obfuscated,
  config.ServerGroup.DEDICATED_IP: ServerType.dedicatedIP,
  // [Deprecated] Region
  config.ServerGroup.EUROPE: ServerType.europe,
  // [Deprecated] Region
  config.ServerGroup.THE_AMERICAS: ServerType.theAmericas,
  // [Deprecated] Region
  config.ServerGroup.ASIA_PACIFIC: ServerType.asiaPacific,
  // [Deprecated] Region
  config.ServerGroup.AFRICA_THE_MIDDLE_EAST_AND_INDIA:
      ServerType.africaTheMiddleEastAndIndia,
};

ServerType? toServerType(config.ServerGroup group) {
  return _groupTitles[group];
}
