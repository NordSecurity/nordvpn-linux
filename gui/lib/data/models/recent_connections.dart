// import 'package:nordvpn/data/models/server_group_extension.dart';
// import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/recent_connections.pb.dart';
import 'package:nordvpn/pb/daemon/server_selection_rule.pb.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart' as cfg;

class RecentConnection {
  final String country;
  final String city;
  final cfg.ServerGroup group;
  final String countryCode;
  final String specificServerName;
  final String specificServer;
  final ServerSelectionRule connectionType;

  RecentConnection({
    required this.country,
    required this.city,
    required this.group,
    required this.countryCode,
    required this.specificServerName,
    required this.specificServer,
    required this.connectionType,
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
    );
  }

  @override
  String toString() {
    return 'RecentConnection(country: $country, city: $city, group: $group, '
        'countryCode: $countryCode, specificServerName: $specificServerName, '
        'specificServer: $specificServer, connectionType: $connectionType)';
  }

  static const Map<cfg.ServerGroup, String> _groupTitles = {
    cfg.ServerGroup.DOUBLE_VPN: "Double VPN",
    cfg.ServerGroup.ONION_OVER_VPN: "Onion Over VPN",
    cfg.ServerGroup.STANDARD_VPN_SERVERS: "Standard VPN Servers",
    cfg.ServerGroup.P2P: "P2P",
    cfg.ServerGroup.OBFUSCATED: "Obfuscated Servers",
    cfg.ServerGroup.DEDICATED_IP: "Dedicated IP",
    cfg.ServerGroup.ULTRA_FAST_TV: "Ultra Fast TV",
    cfg.ServerGroup.ANTI_DDOS: "Anti DDOS",
    cfg.ServerGroup.NETFLIX_USA: "Netflix USA",
    cfg.ServerGroup.EUROPE: "Europe",
    cfg.ServerGroup.THE_AMERICAS: "The Americas",
    cfg.ServerGroup.ASIA_PACIFIC: "Asia Pacific",
    cfg.ServerGroup.AFRICA_THE_MIDDLE_EAST_AND_INDIA:
        "Africa The Middle East and India",
  };

  String get specialtyServer => RecentConnection._groupTitles[group] ?? "";

  String get displayName {
    switch (connectionType) {
      case ServerSelectionRule.CITY:
        return (country != "" && city != "") ? "$country $city" : "";

      case ServerSelectionRule.COUNTRY:
        return country;

      case ServerSelectionRule.SPECIFIC_SERVER:
        return specificServerName;

      case ServerSelectionRule.GROUP:
        return specialtyServer;

      case ServerSelectionRule.COUNTRY_WITH_GROUP:
        final ss = specialtyServer;
        if (ss == "" || country == "") {
          return "";
        }
        return "$ss ($country)";

      case ServerSelectionRule.SPECIFIC_SERVER_WITH_GROUP:
        if (group != cfg.ServerGroup.UNDEFINED) {
          final ss = specialtyServer;
          if (country != "" && city != "") {
            return "$ss ($country, $city)";
          } else if (country != "") {
            return "$ss ($country)";
          }
        }
        return "";

      case ServerSelectionRule.NONE:
      case ServerSelectionRule.RECOMMENDED:
        return "";

      default:
        return "";
    }
  }

  bool get isSpecialtyServer {
    return connectionType == ServerSelectionRule.GROUP;
  }

  bool get isCountryBased {
    return !isSpecialtyServer;
  }
}
