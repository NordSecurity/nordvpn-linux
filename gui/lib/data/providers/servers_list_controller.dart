import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/servers_list.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/i18n/country_names_service.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'servers_list_controller.g.dart';

// This class with handle the fetching, caching, filtering and constructing a tree-like used by the UI for the servers.
@riverpod
class ServersListController extends _$ServersListController
    implements ServersListObserver {
  @override
  FutureOr<ServersList> build() async {
    final isConnected = ref.watch(grpcConnectionControllerProvider).valueOrNull;
    if (isConnected != true) return ServersList.empty();

    _registerNotifications();
    return _fetchServersList();
  }

  void _registerNotifications() {
    final notification = ref.read(appStateProvider);
    notification.addServersListObserver(this);
    ref.onDispose(() {
      notification.removeServersListObserver(this);
    });
  }

  @override
  void onServersListChanged(ServersResponse servers) {
    final serversList = _groupServers(servers);
    if (state.value == serversList) {
      return;
    }

    state = AsyncData(serversList);
  }

  Future<void> refetch() async {
    try {
      state = AsyncData(await _fetchServersList());
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
    }
  }

  Future<ServersList> _fetchServersList() async {
    final vpnProvider = ref.read(vpnRepositoryProvider);
    final servers = await vpnProvider.fetchServers();
    return _groupServers(servers);
  }

  ServersList _groupServers(ServersResponse response) {
    if (response.hasError()) {
      logger.e("failed to fetch servers list ${response.error}");
      throw "failed to fetch servers list";
    }
    // Split the servers to have a hierarchically structure:
    // * each country has a list of cities, and each city has a list of servers
    // * each specialty group have its own list of countries
    Map<ServerType, Map<String, CountryServersGroup>> countries = {};

    for (final country in response.servers.serversByCountry) {
      if (country.cities.isEmpty) {
        continue;
      }

      for (final city in country.cities) {
        if (city.servers.isEmpty) {
          continue;
        }
        Map<ServerType, List<ServerInfo>> cityServerGroups = {};

        for (final server in city.servers) {
          final serverInfo = ServerInfo(
            id: server.id.toInt(),
            hostname: server.hostName,
            isVirtual: server.virtual,
          );

          for (final group in server.serverGroups) {
            final specialtyGroup = group.toSpecialtyType();
            if (specialtyGroup == null) {
              continue;
            }

            cityServerGroups
                .putIfAbsent(specialtyGroup, () => [])
                .add(serverInfo);
          }
        }

        cityServerGroups.forEach((specialty, servers) {
          countries
              .putIfAbsent(specialty, () => {})
              .putIfAbsent(country.countryCode, () {
                // update countries list
                final countryObj = sl<CountryNamesService>().register(
                  code: country.countryCode,
                  name: country.countryName,
                );

                return CountryServersGroup(
                  country: countryObj,
                  cities: [],
                  isVirtual: servers.first.isVirtual,
                );
              })
              .cities
              .add(CityServersGroup(cityName: city.cityName, servers: servers));
        });
      }
    }

    Map<ServerType, List<CountryServersGroup>> serversList = {};
    for (final entry in countries.entries) {
      serversList[entry.key] = entry.value.values.toList();
    }

    return ServersList(serversList);
  }
}
