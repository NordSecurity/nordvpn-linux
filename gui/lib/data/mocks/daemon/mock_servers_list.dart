import 'dart:async';

import 'package:flutter/cupertino.dart';
import 'package:nordvpn/data/mocks/daemon/connect_arguments_extension.dart';
import 'package:nordvpn/pb/daemon/connect.pb.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart' as config;
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:fixnum/fixnum.dart';

final class ServerData {
  final Server server;
  final String countryCode;
  final String cityName;

  ServerData({
    required this.server,
    required this.countryCode,
    required this.cityName,
  });
}

// Store information about the current servers list for the mocked daemon
// Listens on the settings changed and regenerates the list when needed
final class MockServersList {
  final StreamController<AppState> stream;
  Settings? _settings;
  late final StreamSubscription<AppState> _appStateSub;

  MockServersList(this.stream) {
    _serversList = _generateServersList();
    _appStateSub = stream.stream.listen((value) {
      if (value.hasSettingsChange()) {
        if (_settings == value.settingsChange) {
          return;
        }
        _settings = value.settingsChange;
        _serversList = _generateServersList(
          obfuscated: _settings!.obfuscate,
          hasVirtualServers: _settings!.virtualLocation,
        );
      }
    });
  }

  void dispose() => _appStateSub.cancel();

  late ServersResponse _serversList;
  String? error;

  ServersResponse get serversList => _serversList;

  // store DIP servers to faster identify
  late List<Server> _dipServers;
  List<Server> get dipServers => _dipServers;

  set setServersList(ServersResponse list) {
    _serversList = list;
    stream.add(AppState(updateEvent: UpdateEvent.SERVERS_LIST_UPDATE));
  }

  ServersResponse _generateServersList({
    bool hasVirtualServers = true,
    bool obfuscated = false,
  }) {
    debugPrint(
      "Servers list changed hasVirtualServers=$hasVirtualServers - obfuscated=$obfuscated",
    );

    _dipServers = [];

    const obfuscatedGroups = [config.ServerGroup.OBFUSCATED];

    const standardGroups = [
      config.ServerGroup.P2P,
      config.ServerGroup.STANDARD_VPN_SERVERS,
    ];

    final countries = <ServerCountry>[];
    final countryNames = {
      "IT": "Italy",
      "CA": "Canada",
      "FR": "France",
      "DE": "Germany",
      "LT": "Lithuania",
      "US": "USA",
      "IN": "India",
      "ES": "Spain",
      "AU": "Austria",
    };

    Map<String, List<Map<String, List<config.ServerGroup>>>> locations =
        obfuscated
        ? {
            "IT": [
              {"Rome": obfuscatedGroups},
            ],
            "CA": [
              {"Toronto": obfuscatedGroups},
            ],
          }
        : {
            "FR": [
              {"Paris": standardGroups},
              {
                "Marseille": [config.ServerGroup.ONION_OVER_VPN],
              },
            ],
            "DE": [
              {"Berlin": standardGroups},
              {"Frankfurt": standardGroups},
              {"Hamburg": standardGroups},
            ],
            "LT": [
              {"Vilnius": standardGroups},
            ],
            "IN": [
              {"Mumbai": standardGroups},
            ],
            "US": [
              {"Los Angeles": standardGroups},
              {
                "New York": [config.ServerGroup.DOUBLE_VPN],
              },
            ],
            "ES": [
              {
                "Madrid": [config.ServerGroup.DOUBLE_VPN],
              },
              {"Barcelona": standardGroups},
            ],
            "AT": [
              {
                "Vienna": [config.ServerGroup.DEDICATED_IP],
              },
            ],
            "BE": [
              {
                "Bruxelles": [config.ServerGroup.DEDICATED_IP],
              },
            ],
          };

    final technologies = obfuscated
        ? [Technology.OBFUSCATED_OPENVPN_TCP, Technology.OBFUSCATED_OPENVPN_UDP]
        : [Technology.NORDLYNX, Technology.OPENVPN_TCP, Technology.OPENVPN_UDP];

    var serverId = 0;
    for (final countryCode in locations.keys) {
      int serverCounter = 1;
      final cities = <ServerCity>[];
      final isVirtual = (countryCode == "IN");
      if (!hasVirtualServers && isVirtual) {
        continue;
      }

      for (final cityInfo in locations[countryCode]!) {
        final cityName = cityInfo.keys.first;
        final groups = cityInfo.values.first;

        final isDipGroup = groups.contains(config.ServerGroup.DEDICATED_IP);

        // generate some cities
        final servers = <Server>[];
        for (int i = 0; i < (isDipGroup ? 1 : 10); i++) {
          final server = Server(
            id: Int64(serverId),
            hostName: "$countryCode${serverCounter++}.nordvpn.com",
            virtual: isVirtual,
            technologies: technologies,
            serverGroups: groups,
          );
          serverId += 1;
          servers.add(server);

          if (isDipGroup) {
            _dipServers.add(server);
          }
        }

        cities.add(ServerCity(cityName: cityName, servers: servers));
      }

      countries.add(
        ServerCountry(
          countryCode: countryCode,
          countryName: countryNames[countryCode],
          cities: cities,
        ),
      );
    }

    return ServersResponse(servers: ServersMap(serversByCountry: countries));
  }

  ServerData? findServer(ConnectRequest args) {
    if (error != null) {
      throw error!;
    }

    if (serversList.servers.serversByCountry.isEmpty) {
      return null;
    }

    final group = args.serverGroup.isEmpty ? null : args.toServerGroup();
    final tag = args.serverTag.toLowerCase();

    for (final country in serversList.servers.serversByCountry) {
      for (final city in country.cities) {
        for (final server in city.servers) {
          if ((tag == "") ||
              (tag == country.countryCode.toLowerCase()) ||
              (tag ==
                  "${_normalizeName(country.countryCode)} ${_normalizeName(city.cityName)}") ||
              server.hostName.contains(args.serverTag)) {
            for (final g in server.serverGroups) {
              if ((group == null) || (group == g)) {
                return ServerData(
                  server: server,
                  countryCode: country.countryCode,
                  cityName: city.cityName,
                );
              }
            }
          }
        }
      }
    }
    return null;
  }
}

String _normalizeName(String name) {
  return name.toLowerCase().replaceAll(" ", "_");
}
