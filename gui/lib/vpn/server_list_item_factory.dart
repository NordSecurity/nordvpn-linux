import 'package:flutter/material.dart';
import 'package:collection/collection.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/data/models/servers_list.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/theme/servers_list_theme.dart';
import 'package:nordvpn/widgets/custom_expansion_tile.dart';
import 'package:nordvpn/widgets/custom_list_tile.dart';
import 'package:nordvpn/vpn/recent_server_list_item.dart';
import 'package:nordvpn/vpn/server_item_image.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

// Factory for building the ListItem for servers
final class ServerListItemFactory {
  final ImagesManager imagesManager;

  ServerListItemFactory({required this.imagesManager});

  // Build the item for a country. including the cities
  Widget forCountry({
    required BuildContext context,
    ServerType? specialtyGroup,
    required CountryServersGroup country,
    required void Function(ConnectArguments) onTap,
    bool enabled = true,
  }) {
    final appTheme = context.appTheme;
    final serversListThemeData = context.serversListTheme;

    String subtitle;
    if (country.cities.length > 1) {
      subtitle = t.ui.citiesAvailable(n: country.cities.length);
    } else {
      subtitle = country.cities.first.localizedName;
      if (country.isVirtual) {
        subtitle += " - ${t.ui.virtual}";
      }
    }

    assert(
      country.cities.isSorted(
        (a, b) => a.localizedName.compareTo(b.localizedName),
      ),
      "cities are not sorted",
    );

    return CustomExpansionTile(
      minTileHeight: serversListThemeData.listItemHeight,
      childrenPadding: EdgeInsets.only(left: serversListThemeData.flagSize),
      leading: ServerItemImage(
        image: imagesManager.forCountry(country.country),
        shouldHighlight: (status) =>
            _shouldHighlight(specialtyGroup, country, status),
      ),
      title: Text(country.countryName, style: appTheme.body),
      subtitle: Text(subtitle, style: appTheme.caption),
      onTap: enabled
          ? () {
              assert(country.cities.isNotEmpty);
              onTap(
                ConnectArguments(
                  country: country.country,
                  specialtyGroup: specialtyGroup,
                  city: country.cities.length == 1
                      ? City(country.cities.first.name)
                      : null,
                ),
              );
            }
          : null,
      children: (country.cities.length == 1)
          ? null
          : [
              for (final city in country.cities)
                ListTile(
                  minTileHeight: serversListThemeData.listItemHeight,
                  leading: ServerItemImage(
                    image: DynamicThemeImage("city_pin.svg"),
                    shouldHighlight: (status) =>
                        _shouldHighlight(specialtyGroup, country, status) &&
                        city.name == status.city?.name,
                  ),
                  title: Text(city.localizedName, style: appTheme.body),
                  onTap: enabled
                      ? () => onTap(
                          ConnectArguments(
                            country: country.country,
                            city: City(city.name),
                            specialtyGroup: specialtyGroup,
                          ),
                        )
                      : null,
                ),
            ],
    );
  }

  bool _shouldHighlight(
    ServerType? serverType,
    CountryServersGroup connectedTo,
    VpnStatus status,
  ) {
    if (!status.isConnected()) return false;

    final countryMatches = connectedTo.country == status.country;
    final anyCityMatches = connectedTo.cities.any(
      (group) => group.city == status.city,
    );
    final groupMatches =
        serverType?.toServerGroup() == status.connectionParameters.group ||
        status.connectionParameters.group == ServerGroup.P2P;

    if (status.connectionParameters.group != ServerGroup.UNDEFINED) {
      return (countryMatches || anyCityMatches) && groupMatches;
    }

    return countryMatches || anyCityMatches;
  }

  // Build item for a specialty server
  Widget forSpecialtyServer({
    required BuildContext context,
    required ServerType type,
    required List<CountryServersGroup> servers,
    bool? enabled,
    String? subtitle,
    required void Function(ConnectArguments) onTap,
    required void Function() showDetails,
  }) {
    final theme = Theme.of(context);
    final appTheme = theme.extension<AppTheme>()!;
    final serversListThemeData = theme.extension<ServersListTheme>()!;

    final isEnabled = enabled ?? true;

    final styleTitle = isEnabled
        ? appTheme.body
        : appTheme.body.copyWith(color: theme.disabledColor);

    final styleSubtitle = isEnabled
        ? appTheme.caption
        : appTheme.caption.copyWith(color: theme.disabledColor);

    return CustomListTile(
      enabled: isEnabled,
      minTileHeight: serversListThemeData.listItemHeight,
      leading: ServerItemImage(
        image: imagesManager.forSpecialtyServer(type),
        shouldHighlight: (status) =>
            status.connectionParameters.group == type.toServerGroup(),
      ),
      title: Text(labelForServerType(type), style: styleTitle),
      subtitle: (subtitle != null)
          ? Text(subtitle, style: styleSubtitle)
          : null,
      onTap: () => onTap(ConnectArguments(specialtyGroup: type)),
      trailing: servers.isNotEmpty
          ? IconButton(
              icon: DynamicThemeImage("three_dots.svg"),
              onPressed: () => showDetails(),
              hoverColor: Colors.transparent,
              splashColor: Colors.transparent,
              highlightColor: Colors.transparent,
            )
          : null,
    );
  }

  // Build item for a server that is a city into the search list
  Widget forCityAtSearch({
    required BuildContext context,
    ServerType? specialtyGroup,
    required CountryServersGroup country,
    required void Function(ConnectArguments) onTap,
  }) {
    final theme = Theme.of(context);
    final appTheme = theme.extension<AppTheme>()!;
    final serversListThemeData = theme.extension<ServersListTheme>()!;

    final city = country.cities.first.city;

    return CustomListTile(
      minTileHeight: serversListThemeData.listItemHeight,
      contentPadding: EdgeInsets.symmetric(
        horizontal: serversListThemeData.flagSize,
      ),
      leading: ServerItemImage(
        image: imagesManager.forCountry(country.country),
        shouldHighlight: (status) => city == status.city,
      ),
      title: Text(city.localizedName, style: appTheme.body),
      subtitle: Text(country.country.localizedName, style: appTheme.caption),
      // children: null,
      onTap: () => onTap(
        ConnectArguments(
          country: country.country,
          city: city,
          specialtyGroup: specialtyGroup,
        ),
      ),
    );
  }

  // Build item for a server. Used when searching after a server name
  Widget forServerInfo({
    required BuildContext context,
    required Country country,
    required ServerInfo server,
    required void Function(ConnectArguments) onTap,
  }) {
    final appTheme = context.appTheme;
    final serversListTheme = context.serversListTheme;

    return CustomListTile(
      minTileHeight: serversListTheme.listItemHeight,
      contentPadding: EdgeInsets.symmetric(
        horizontal: serversListTheme.flagSize,
      ),
      leading: ServerItemImage(
        image: imagesManager.forCountry(country),
        shouldHighlight: (status) => server.hostname == status.hostname,
      ),
      title: Text(country.localizedName, style: appTheme.body),
      subtitle: Text("#${server.serverNumber}", style: appTheme.caption),
      onTap: () => onTap(ConnectArguments(server: server)),
    );
  }

  Widget forRecent({
    required RecentConnection recentConnection,
    required void Function(ConnectArguments) onTapFunc,
    bool enabled = true,
  }) {
    return RecentServerListItem(
      model: recentConnection,
      onTap: onTapFunc,
      enabled: enabled,
      imagesManager: imagesManager,
    );
  }
}
