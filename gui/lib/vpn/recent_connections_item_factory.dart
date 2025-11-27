import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart';
import 'package:nordvpn/pb/daemon/server_selection_rule.pb.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/servers_list_theme.dart';
import 'package:nordvpn/vpn/server_item_image.dart';
import 'package:nordvpn/widgets/custom_list_tile.dart';

/// Factory for building list items for recent connections
final class RecentConnectionsItemFactory {
  final ImagesManager imagesManager;

  RecentConnectionsItemFactory({required this.imagesManager});

  /// Build a list item for a recent connection
  Widget forRecentConnection({
    required BuildContext context,
    required RecentConnection model,
    required void Function(ConnectArguments) onTap,
  }) {
    final appTheme = context.appTheme;
    final serversListTheme = context.serversListTheme;

    final isSpecialtyServer =
        model.group != ServerGroup.UNDEFINED &&
        model.group != ServerGroup.STANDARD_VPN_SERVERS;

    // Pre-compute connect arguments to avoid recalculation on each tap
    final connectArgs = _buildConnectArgs(model, isSpecialtyServer);

    return CustomListTile(
      minTileHeight: serversListTheme.listItemHeight,
      contentPadding: EdgeInsets.only(left: 0),
      leading: ServerItemImage(image: _buildImage(model, isSpecialtyServer)),
      title: _buildTitle(appTheme, model, isSpecialtyServer),
      onTap: () => onTap(connectArgs),
    );
  }

  Widget _buildImage(RecentConnection model, bool isSpecialtyServer) {
    final isCountry = model.countryCode.isNotEmpty && model.country.isNotEmpty;

    // early return for specialty server without country
    if (isSpecialtyServer && !isCountry) {
      final serverType = toServerType(model.group);
      return serverType != null
          ? imagesManager.forSpecialtyServer(serverType)
          : const Icon(Icons.history);
    }

    // handle country-based images (works for both specialty and standard servers)
    if (isCountry) {
      return imagesManager.forCountry(Country.fromCode(model.countryCode));
    }

    // fallback: try to get specialty server image or default icon
    final serverType = toServerType(model.group);
    return serverType != null
        ? imagesManager.forSpecialtyServer(serverType)
        : const Icon(Icons.history);
  }

  Widget _buildTitle(
    AppTheme appTheme,
    RecentConnection model,
    bool isSpecialtyServer,
  ) {
    if (isSpecialtyServer) {
      var specialtyTitle = Text(model.specialtyServer, style: appTheme.body);
      if (model.country.isNotEmpty) {
        final country = Country.fromCode(model.countryCode);
        var subtitle = country.localizedName;
        final city = model.city;
        subtitle += " - ${city.isEmpty ? t.ui.fastestServer : City(city).localizedName}";

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            specialtyTitle,
            Text(
              _maybeAddVirtualLabel(subtitle, model.isVirtual),
              style: appTheme.caption,
            ),
          ],
        );
      }

      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          specialtyTitle,
          Text(t.ui.fastestServer, style: appTheme.caption),
        ],
      );
    }

    final isCity =
        model.city.isNotEmpty &&
        model.connectionType == ServerSelectionRule.CITY;

    if (isCity) {
      final country = Country.fromCode(model.countryCode);
      final city = City(model.city);
      final cityText = _maybeAddVirtualLabel(
        city.localizedName,
        model.isVirtual,
      );
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(country.localizedName, style: appTheme.body),
          Text(cityText, style: appTheme.caption),
        ],
      );
    }

    final isSpecificServer =
        model.specificServerName.isNotEmpty &&
        model.connectionType == ServerSelectionRule.SPECIFIC_SERVER;

    if (isSpecificServer) {
      final country = Country.fromCode(model.countryCode);
      final serverId = model.serverId;
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(country.localizedName, style: appTheme.body),
          if (serverId != null)
            Text(
              _maybeAddVirtualLabel(serverId, model.isVirtual),
              style: appTheme.caption,
            ),
        ],
      );
    }

    final country = Country.fromCode(model.countryCode);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(country.localizedName, style: appTheme.body),
        Text(t.ui.fastestServer, style: appTheme.caption),
      ],
    );
  }

  String _maybeAddVirtualLabel(String text, bool isVirtual) {
    return isVirtual ? "$text - ${t.ui.virtual}" : text;
  }

  ConnectArguments _buildConnectArgs(
    RecentConnection model,
    bool isSpecialtyServer,
  ) {
    if (model.connectionType == ServerSelectionRule.SPECIFIC_SERVER &&
        model.specificServerName.isNotEmpty) {
      return ConnectArguments(
        server: ServerInfo(
          id: 0,
          hostname: model.specificServerName,
          isVirtual: model.isVirtual,
        ),
      );
    } else {
      Country? country;
      if (model.countryCode.isNotEmpty) {
        country = Country.fromCode(model.countryCode);
      }
      return ConnectArguments(
        country: country,
        city: model.city.isNotEmpty ? City(model.city) : null,
        specialtyGroup: isSpecialtyServer ? toServerType(model.group) : null,
      );
    }
  }
}
