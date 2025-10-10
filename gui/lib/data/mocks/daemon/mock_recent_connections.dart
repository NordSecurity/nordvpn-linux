import 'package:nordvpn/data/mocks/daemon/mock_servers_list.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart' as cfg;
import 'package:nordvpn/pb/daemon/recent_connections.pb.dart';
import 'package:nordvpn/pb/daemon/server_selection_rule.pbenum.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';

class MockRecentConnections {
  final MockServersList _serversList;

  MockRecentConnections(this._serversList);

  ServerCountry? _findCountry(List<ServerCountry> servers, String code) {
    for (var server in servers) {
      if (server.countryCode == code) {
        return server;
      }
    }
    return null;
  }

  ServerCity? _findCity(ServerCountry country, String name) {
    for (var city in country.cities) {
      if (city.cityName == name) {
        return city;
      }
    }
    return null;
  }

  List<RecentConnectionModel> getConnections() {
    final servers = _serversList.serversList.servers.serversByCountry;
    final List<RecentConnectionModel> recentConnections = [];

    if (servers.isEmpty) {
      return [];
    }

    // 1. A city connection
    var country = _findCountry(servers, 'LT');
    if (country != null) {
      var city = _findCity(country, 'Vilnius');
      if (city != null) {
        recentConnections.add(RecentConnectionModel(
          city: city.cityName,
          connectionType: ServerSelectionRule.CITY,
          country: country.countryName,
          countryCode: country.countryCode,
          group: cfg.ServerGroup.UNDEFINED,
        ));
      }
    }

    // 2. A country connection
    country = _findCountry(servers, 'JP');
    if (country != null) {
      recentConnections.add(RecentConnectionModel(
        connectionType: ServerSelectionRule.COUNTRY,
        country: country.countryName,
        countryCode: country.countryCode,
        group: cfg.ServerGroup.UNDEFINED,
      ));
    }

    // 3. A group connection
    recentConnections.add(RecentConnectionModel(
      connectionType: ServerSelectionRule.GROUP,
      group: cfg.ServerGroup.ASIA_PACIFIC,
    ));

    // 4. Another city
    country = _findCountry(servers, 'US');
    if (country != null) {
      var city = _findCity(country, 'New York');
      if (city != null) {
        recentConnections.add(RecentConnectionModel(
          city: city.cityName,
          connectionType: ServerSelectionRule.CITY,
          country: country.countryName,
          countryCode: country.countryCode,
          group: cfg.ServerGroup.UNDEFINED,
        ));
      }
    }

    // 5. Another country
    country = _findCountry(servers, 'DE');
    if (country != null) {
      recentConnections.add(RecentConnectionModel(
        connectionType: ServerSelectionRule.COUNTRY,
        country: country.countryName,
        countryCode: country.countryCode,
        group: cfg.ServerGroup.UNDEFINED,
      ));
    }

    // 6. Another group
    recentConnections.add(RecentConnectionModel(
      connectionType: ServerSelectionRule.GROUP,
      group: cfg.ServerGroup.EUROPE,
    ));

    // 7. Another city
    country = _findCountry(servers, 'GB');
    if (country != null) {
      var city = _findCity(country, 'London');
      if (city != null) {
        recentConnections.add(RecentConnectionModel(
          city: city.cityName,
          connectionType: ServerSelectionRule.CITY,
          country: country.countryName,
          countryCode: country.countryCode,
          group: cfg.ServerGroup.UNDEFINED,
        ));
      }
    }

    // 8. Another country
    country = _findCountry(servers, 'CA');
    if (country != null) {
      recentConnections.add(RecentConnectionModel(
        connectionType: ServerSelectionRule.COUNTRY,
        country: country.countryName,
        countryCode: country.countryCode,
        group: cfg.ServerGroup.UNDEFINED,
      ));
    }

    // 9. Another group
    recentConnections.add(RecentConnectionModel(
      connectionType: ServerSelectionRule.GROUP,
      group: cfg.ServerGroup.THE_AMERICAS,
    ));

    // 10. Another city
    country = _findCountry(servers, 'FR');
    if (country != null) {
      var city = _findCity(country, 'Paris');
      if (city != null) {
        recentConnections.add(RecentConnectionModel(
          city: city.cityName,
          connectionType: ServerSelectionRule.CITY,
          country: country.countryName,
          countryCode: country.countryCode,
          group: cfg.ServerGroup.UNDEFINED,
        ));
      }
    }

    return recentConnections;
  }
}
