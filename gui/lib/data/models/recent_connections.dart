import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/recent_connections.pb.dart';
import 'package:nordvpn/pb/daemon/server_selection_rule.pb.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart' as cfg;

class RecentConnection {
  // Matches server id from specific server name
  // e.g., "Lithuania #123" -> captures "#123"
  static final _serverIdRegex = RegExp(r'^[A-Za-z\s-]+(#\d+)$');

  final String country;
  final String city;
  final cfg.ServerGroup group;
  final String countryCode;
  final String specificServerName;
  final String specificServer;
  final ServerSelectionRule connectionType;
  final bool isVirtual;

  RecentConnection({
    required this.country,
    required this.city,
    required this.group,
    required this.countryCode,
    required this.specificServerName,
    required this.specificServer,
    required this.connectionType,
    required this.isVirtual,
  });

  factory RecentConnection.fromPb(RecentConnectionModel pb) {
    return RecentConnection(
      country: pb.country,
      city: pb.city,
      group: pb.group,
      countryCode: pb.countryCode,
      specificServerName: pb.specificServerName,
      specificServer: pb.specificServer,
      connectionType: pb.connectionType,
      isVirtual: pb.isVirtual,
    );
  }

  @override
  String toString() {
    return 'RecentConnection(country: $country, city: $city, group: $group, '
        'countryCode: $countryCode, specificServerName: $specificServerName, '
        'specificServer: $specificServer, connectionType: $connectionType, virtual: $isVirtual)';
  }

  static const Map<cfg.ServerGroup, String> _groupTitles = {
    cfg.ServerGroup.DOUBLE_VPN: "Double VPN",
    cfg.ServerGroup.ONION_OVER_VPN: "Onion Over VPN",
    cfg.ServerGroup.STANDARD_VPN_SERVERS: "Standard VPN Servers",
    cfg.ServerGroup.P2P: "P2P",
    cfg.ServerGroup.OBFUSCATED: "Obfuscated Servers",
    cfg.ServerGroup.DEDICATED_IP: "Dedicated IP",
    // [Deprecated] Region
    cfg.ServerGroup.EUROPE: "Europe",
    // [Deprecated] Region
    cfg.ServerGroup.THE_AMERICAS: "The Americas",
    // [Deprecated] Region
    cfg.ServerGroup.ASIA_PACIFIC: "Asia Pacific",
    // [Deprecated] Region
    cfg.ServerGroup.AFRICA_THE_MIDDLE_EAST_AND_INDIA:
        "Africa The Middle East and India",
  };

  String get specialtyServer => RecentConnection._groupTitles[group] ?? "";

  bool get isSpecialtyServer {
    return connectionType == ServerSelectionRule.GROUP;
  }

  bool get isCountryBased {
    return !isSpecialtyServer;
  }

  /// Extracts server ID from specific server name
  /// e.g., "Lithuania #123" -> "#123"
  String? get serverId {
    final match = _serverIdRegex.firstMatch(specificServerName);
    return match?[1];
  }
}
