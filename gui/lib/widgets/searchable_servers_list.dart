import 'package:flutter/material.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/servers_list.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/pair.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/servers_list_theme.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/vpn/server_list_item_factory.dart';
import 'package:nordvpn/widgets/custom_expansion_tile.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/interactive_list_view.dart';

enum _SearchGroup { countries, cities, specialty, servers }

// Display an adaptive searchable list view for servers.
// If there are few servers the search bar is removed and instead a static list
// is displayed
final class SearchableServersList extends StatelessWidget {
  final ServersList? serversList;
  final List<CountryServersGroup>? servers;
  final ServerType? specialtyServer;
  final bool allowServerNameSearch;
  final Function(ConnectArguments) onTap;
  final TextEditingController? searchTextController;
  final Widget? leadingWidget;
  final ServerListItemFactory itemFactory;
  late final List<Object> _serverItems;
  final bool _isObfuscatedEnabled;
  final bool withQuickConnectTile;

  SearchableServersList({
    super.key,
    required this.servers,
    this.specialtyServer,
    this.allowServerNameSearch = true,
    required this.onTap,
    this.serversList,
    this.searchTextController,
    this.leadingWidget,
    this.withQuickConnectTile = false,
    ServerListItemFactory? itemFactory,
  }) : assert((serversList != null) != (servers != null)),
       itemFactory = itemFactory ?? sl(),
       _isObfuscatedEnabled = (specialtyServer == ServerType.obfuscated) {
    _initServerItems();
  }

  void _initServerItems() {
    _serverItems = List<Object>.from(servers ?? []);
    if (_serverItems.isNotEmpty && withQuickConnectTile) {
      _serverItems.insert(0, _QuickConnectTile());
    }
  }

  SearchableServersList.forServersList({
    super.key,
    required this.serversList,
    required this.onTap,
    this.allowServerNameSearch = true,
    this.servers,
    this.specialtyServer,
    this.searchTextController,
    this.leadingWidget,
    this.withQuickConnectTile = false,
    ServerListItemFactory? itemFactory,
  }) : assert((serversList != null) != (servers != null)),
       itemFactory = itemFactory ?? sl(),
       _isObfuscatedEnabled = (specialtyServer == ServerType.obfuscated) {
    _initServerItems();
  }

  @override
  Widget build(BuildContext context) {
    if (!_shouldShowSearchableList()) {
      return ListView(
        shrinkWrap: true,
        children: [
          for (final item in _serverItems)
            _buildSearchListItem(
              context,
              item,
              specialtyGroup: specialtyServer,
              onTap: (args) =>
                  onTap(args.copyWith(specialtyGroup: specialtyServer)),
            ),
        ],
      );
    }
    return _buildSearchableList(context);
  }

  // checks if it should be a list view with search bar or a static list view
  bool _shouldShowSearchableList() {
    if (servers == null) {
      return true;
    }

    // If there are more than 2 cities or more than 1 city in a country,
    // then show the view with the search.
    // Because if a country has more than 1 server it can be expanded,
    // changing the view height => change of size
    var numberOfCities = 0;
    for (final item in servers!) {
      numberOfCities += item.cities.length;
      if (numberOfCities > 2 || item.cities.length > 1) {
        // set some big value to be sure it will display the search list view
        // and the Quick Connect button
        return true;
      }
    }

    return numberOfCities > 2;
  }

  Widget _buildSearchableList(BuildContext context) {
    final appTheme = context.appTheme;
    final hint = allowServerNameSearch
        ? t.ui.searchServersHint
        : t.ui.searchCountryAndCity;

    return InteractiveListView(
      beginSearchAfter: serverSearchAfterNumChars,
      searchHintText: hint,
      searchTextController: searchTextController,
      leadingWidget: leadingWidget,
      items: _serverItems,
      itemBuilder: (context, item) => _buildSearchListItem(
        context,
        item,
        specialtyGroup: specialtyServer,
        onTap: (args) => onTap(args.copyWith(specialtyGroup: specialtyServer)),
      ),
      filter: (query, items) => (servers != null)
          ? _filterServers(query.toLowerCase(), servers!)
          : _filterServersList(query.toLowerCase(), serversList!),
      searchBarSize: appTheme.body,
      noResultsFoundWidget: Center(child: _noResultsWidget(context)),
      showEmptyListAtStartup: servers == null,
      emptyListWidget: _warningObfuscationEnabled(context),
    );
  }

  Widget _noResultsWidget(BuildContext context) {
    final appTheme = context.appTheme;
    final serversListTheme = context.serversListTheme;

    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      mainAxisSize: MainAxisSize.min,
      spacing: appTheme.verticalSpaceLarge,
      children: [
        if (!_isObfuscatedEnabled) DynamicThemeImage("results_not_found.svg"),
        Text(
          _isObfuscatedEnabled
              ? t.ui.obfuscationErrorNoServerFound
              : t.ui.noResultsFound,
          style: serversListTheme.searchErrorStyle,
          textAlign: TextAlign.center,
        ),
        if (_isObfuscatedEnabled)
          TextButton(
            child: Text(t.ui.goToSettings),
            onPressed: () =>
                context.navigateToRoute(AppRoute.settingsSecurityAndPrivacy),
          ),
      ],
    );
  }

  Widget? _warningObfuscationEnabled(BuildContext context) {
    if (!_isObfuscatedEnabled) {
      return null;
    }

    final serversListTheme = context.serversListTheme;

    return Center(
      child: Text(
        t.ui.obfuscationSearchWarning,
        textAlign: TextAlign.center,
        style: serversListTheme.obfuscationSearchWarningStyle,
      ),
    );
  }

  Widget _buildSearchListItem(
    BuildContext context,
    dynamic item, {
    ServerType? specialtyGroup,
    required void Function(ConnectArguments) onTap,
  }) {
    final appTheme = context.appTheme;
    final serversListTheme = context.serversListTheme;

    if (item is _QuickConnectTile) {
      return CustomExpansionTile(
        leading: DynamicThemeImage("fastest_server.svg"),
        title: Text(fastestServerLabel, style: appTheme.body),
        onTap: () => onTap(ConnectArguments()),
      );
    }

    if (item is CountryServersGroup) {
      // this is used at building the specialty details popup
      return itemFactory.forCountry(
        context: context,
        specialtyGroup: specialtyGroup,
        country: item,
        onTap: (args) => onTap(args),
      );
    }

    // this is an item used in search
    final group = item as Pair<_SearchGroup, dynamic>;

    String groupName;
    List<Widget> widgets = [];

    switch (group.first) {
      case _SearchGroup.countries:
        groupName = t.ui.countries;
        for (final item in group.second as List<CountryServersGroup>) {
          widgets.add(
            itemFactory.forCountry(
              context: context,
              specialtyGroup: specialtyGroup,
              country: item,
              onTap: (args) => onTap(args),
            ),
          );
        }
        break;

      case _SearchGroup.cities:
        groupName = t.ui.cities;
        for (final item in group.second as List<CountryServersGroup>) {
          assert(item.cities.length == 1);

          widgets.add(
            itemFactory.forCityAtSearch(
              context: context,
              specialtyGroup: specialtyGroup,
              country: item,
              onTap: (args) => onTap(args),
            ),
          );
        }
        break;

      case _SearchGroup.specialty:
        groupName = t.ui.specialServers;

        for (final type in group.second as List<ServerType>) {
          widgets.add(
            itemFactory.forSpecialtyServer(
              context: context,
              type: type,
              servers: [],
              onTap: (args) => onTap(args),
              showDetails: () {},
            ),
          );
        }

        break;

      case _SearchGroup.servers:
        groupName = t.ui.servers;
        for (final item in group.second) {
          final pairCountryServers = item as Pair<Country, List<ServerInfo>>;
          for (final server in pairCountryServers.second) {
            widgets.add(
              itemFactory.forServerInfo(
                context: context,
                country: pairCountryServers.first,
                server: server,
                onTap: (args) => onTap(args),
              ),
            );
          }
        }

        break;
    }

    return CustomExpansionTile(
      key: ValueKey(groupName),
      expanded: true,
      hideExpandButton: true,
      minTileHeight: serversListTheme.listItemHeight,
      title: Padding(
        padding: serversListTheme.paddingSearchGroupsLabel,
        child: Text(groupName, style: appTheme.caption),
      ),
      children: widgets,
    );
  }

  // This is used at search for all the servers and groups
  List<Pair<_SearchGroup, List<dynamic>>> _filterServersList(
    String query,
    ServersList serversList,
  ) {
    final results = _filterServers(
      query,
      _isObfuscatedEnabled
          ? serversList.obfuscatedServersList
          : serversList.standardServersList,
    );

    // search after specialty servers names
    final specialtyServersOrder = [
      ServerType.dedicatedIP,
      ServerType.doubleVpn,
      ServerType.onionOverVpn,
      ServerType.p2p,
    ];

    List<ServerType> matchedSpecialtyServers = [];
    for (final type in specialtyServersOrder) {
      if (labelForServerType(type).toLowerCase().startsWith(query)) {
        final servers = serversList.groups[type];
        if (servers == null || servers.isEmpty) {
          continue;
        }
        matchedSpecialtyServers.add(type);
      }
    }

    if (matchedSpecialtyServers.isNotEmpty) {
      results.add(Pair(_SearchGroup.specialty, matchedSpecialtyServers));
    }

    return results;
  }

  List<Pair<_SearchGroup, List<dynamic>>> _filterServers(
    String query,
    List<CountryServersGroup> countriesList,
  ) {
    if (query.startsWith("#")) {
      if (!allowServerNameSearch) {
        return [];
      }
      // search after server name
      final searchedServerNumber = query.substring(1);

      if (int.tryParse(searchedServerNumber, radix: 10) == null) {
        logger.d("failed to parse server number $query");
        return [];
      }

      // number of servers found
      var serversCount = 0;
      // keep track of servers found in exact match search
      final Set<ServerInfo> foundServers = {};

      // do exact search
      List<Pair<Country, List<ServerInfo>>> results = [];
      for (final countryGroup in countriesList) {
        List<ServerInfo> servers = [];
        for (final city in countryGroup.cities) {
          for (final server in city.servers) {
            if (server.serverNumber == searchedServerNumber) {
              servers.add(server);
              serversCount += 1;
              foundServers.add(server);
            }
          }
        }

        if (servers.isNotEmpty) {
          results.add(Pair(countryGroup.country, servers));
        }
      }

      // continue searching for partial search
      if (serversCount < maxNumberOfServersResults) {
        countriesLoop:
        for (final countryGroup in countriesList) {
          List<ServerInfo> servers = [];
          citiesLoop:
          for (final city in countryGroup.cities) {
            for (final server in city.servers) {
              if (foundServers.contains(server)) {
                // ignore exact match servers
                continue;
              }

              if (server.serverNumber.startsWith(searchedServerNumber)) {
                servers.add(server);
                serversCount += 1;
                if (serversCount >= maxNumberOfServersResults) {
                  break citiesLoop;
                }
              }
            }
          }
          if (servers.isNotEmpty) {
            results.add(Pair(countryGroup.country, servers));
          }

          if (serversCount >= maxNumberOfServersResults) {
            break countriesLoop;
          }
        }

        assert(serversCount <= maxNumberOfServersResults);
      }

      return [if (results.isNotEmpty) Pair(_SearchGroup.servers, results)];
    }

    List<CountryServersGroup> countries = [];
    List<CountryServersGroup> cities = [];

    for (final countryGroup in countriesList) {
      if ((countryGroup.code.toLowerCase().startsWith(query)) ||
          countryGroup.localizedName.toLowerCase().startsWith(query)) {
        countries.add(countryGroup);
      } else {
        for (final city in countryGroup.cities) {
          if (city.localizedName.toLowerCase().startsWith(query)) {
            cities.add(
              CountryServersGroup(
                country: countryGroup.country,
                isVirtual: countryGroup.isVirtual,
                cities: [city],
              ),
            );
          }
        }
      }
    }

    return [
      if (countries.isNotEmpty) Pair(_SearchGroup.countries, countries),
      if (cities.isNotEmpty) Pair(_SearchGroup.cities, cities),
    ];
  }
}

// It's just a marker class describing "Quick connect" tile
final class _QuickConnectTile {}
