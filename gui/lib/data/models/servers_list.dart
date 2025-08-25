import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_info.dart';

// This contains all the serves available into the app
final class ServersList {
  // The countries and cities are not sorted
  final Map<ServerType, List<CountryServersGroup>> groups;

  factory ServersList.empty() => ServersList({});

  ServersList(this.groups) {
    // For now order the countries by the display name here.
    // This will not work if in the future multiple languages are supported
    // and language changes at runtime.
    groups.forEach((_, countries) {
      // Compare localized country names
      countries.sort((a, b) => a.countryName.compareTo(b.countryName));

      // sort cities from each country
      for (final country in countries) {
        country.cities.sort(
          (a, b) => a.localizedName.compareTo(b.localizedName),
        );
      }
    });
  }

  List<CountryServersGroup> get standardServersList =>
      groups[ServerType.standardVpn] ?? [];

  List<CountryServersGroup> get obfuscatedServersList =>
      groups[ServerType.obfuscated] ?? [];

  List<CountryServersGroup> specialtyServersList(ServerType type) {
    return groups[type] ?? [];
  }

  // search over all the countries from the group and if it contains
  // the searched id returns a list of countries containing found servers
  List<CountryServersGroup> findServers(Set<int> serverIds, ServerType type) {
    final countries = groups[type];
    if (countries == null || serverIds.isEmpty) {
      return [];
    }

    List<CountryServersGroup> result = [];

    for (final country in countries) {
      Map<String, List<ServerInfo>> cityServers = {};
      for (final city in country.cities) {
        for (final server in city.servers) {
          if (serverIds.contains(server.id)) {
            cityServers.putIfAbsent(city.name, () => []).add(server);
            serverIds.remove(server.id);
            if (serverIds.isEmpty) {
              break;
            }
          }
        }
        if (serverIds.isEmpty) {
          break;
        }
      }

      if (cityServers.isNotEmpty) {
        result.add(
          CountryServersGroup(
            country: country.country,
            cities: [
              for (final cityName in cityServers.keys.toList(
                growable: false,
              )..sort())
                CityServersGroup(
                  cityName: cityName,
                  servers: cityServers[cityName]!,
                ),
            ],
            isVirtual: country.isVirtual,
          ),
        );
      }
      if (serverIds.isEmpty) {
        break;
      }
    }

    return result;
  }

  bool hasSpecialtyServers(ServerType type) {
    return groups[type]?.isNotEmpty ?? false;
  }
}

// Represent the servers from a country
// A country has a list of cities. Each city has a list of servers
final class CountryServersGroup {
  final Country country;
  final List<CityServersGroup> cities;
  final bool isVirtual;

  CountryServersGroup({
    required this.country,
    required this.cities,
    required this.isVirtual,
  });

  String get countryName => country.name;
  String get code => country.code;
  String get localizedName => country.localizedName;
}

// Stores all the servers from a city
final class CityServersGroup {
  City city;
  List<ServerInfo> servers;
  CityServersGroup({required String cityName, required this.servers})
    : city = City(cityName);

  // City name received from the API or daemon, in english
  String get name => city.name;

  // The city name to display on the UI
  String get localizedName => city.localizedName;

  @override
  String toString() {
    return "CityServersGroup(city: $city, servers: ${servers.length})";
  }
}
